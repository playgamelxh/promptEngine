package services

import (
	"codeagent-backend/models"
	"codeagent-backend/utils"
	"context"
	"crypto/md5"
	"encoding/hex"
)

type TestCaseService struct{}

func (s *TestCaseService) calculateMD5(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func (s *TestCaseService) CreateTestCase(testCase *models.TestCase) (bool, error) {
	testCase.InputMD5 = s.calculateMD5(testCase.Input)

	// Check for duplicates in the same project
	if testCase.ProjectID != 0 {
		var count int64
		utils.DB.Model(&models.TestCase{}).
			Where("project_id = ? AND input_md5 = ?", testCase.ProjectID, testCase.InputMD5).
			Count(&count)
		if count > 0 {
			return false, nil // Duplicate
		}
	}

	err := utils.DB.Create(testCase).Error
	return true, err
}

func (s *TestCaseService) GetTestCases(projectID, promptID string, page, pageSize int) ([]models.TestCase, int64, error) {
	var testCases []models.TestCase
	var total int64

	query := utils.DB.Model(&models.TestCase{})

	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}
	if promptID != "" {
		query = query.Where("prompt_id = ?", promptID)
	}

	query.Count(&total)
	err := query.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&testCases).Error
	return testCases, total, err
}

func (s *TestCaseService) GetTestCase(id string) (*models.TestCase, error) {
	var testCase models.TestCase
	err := utils.DB.First(&testCase, id).Error
	return &testCase, err
}

func (s *TestCaseService) UpdateTestCase(testCase *models.TestCase, originalInput string) (bool, error) {
	if testCase.Input != originalInput {
		testCase.InputMD5 = s.calculateMD5(testCase.Input)
		// Check uniqueness if input changed
		if testCase.ProjectID != 0 {
			var count int64
			utils.DB.Model(&models.TestCase{}).
				Where("project_id = ? AND input_md5 = ? AND id != ?", testCase.ProjectID, testCase.InputMD5, testCase.ID).
				Count(&count)
			if count > 0 {
				return false, nil // Duplicate
			}
		}
	}

	err := utils.DB.Save(testCase).Error
	return true, err
}

func (s *TestCaseService) DeleteTestCase(testCase *models.TestCase) error {
	return utils.DB.Delete(testCase).Error
}

func (s *TestCaseService) BatchDeleteTestCases(ids []uint) error {
	return utils.DB.Delete(&models.TestCase{}, ids).Error
}

// GenerateTestCases generates test cases using LLM and saves them
func (s *TestCaseService) GenerateTestCases(ctx context.Context, configID uint, promptIDs []uint, count int) ([]models.TestCase, error) {
	var config models.LLMConfig
	if err := utils.DB.First(&config, configID).Error; err != nil {
		return nil, err
	}

	var allCreatedCases []models.TestCase
	llmService := new(LLMService) // Use local instance or inject

	for _, promptID := range promptIDs {
		var prompt models.Prompt
		if err := utils.DB.First(&prompt, promptID).Error; err != nil {
			continue
		}

		generatedCases, err := llmService.GenerateTestCases(ctx, config, prompt.Content, count)
		if err != nil {
			continue
		}

		for _, generated := range generatedCases {
			inputMD5 := s.calculateMD5(generated.Input)

			// Check for duplicates in the same project
			var count int64
			utils.DB.Model(&models.TestCase{}).
				Where("project_id = ? AND input_md5 = ?", prompt.ProjectID, inputMD5).
				Count(&count)

			if count > 0 {
				continue
			}

			testCase := models.TestCase{
				PromptID:       prompt.ID,
				ProjectID:      prompt.ProjectID,
				Input:          generated.Input,
				InputMD5:       inputMD5,
				ExpectedOutput: generated.ExpectedOutput,
			}
			utils.DB.Create(&testCase)
			allCreatedCases = append(allCreatedCases, testCase)
		}
	}

	return allCreatedCases, nil
}
