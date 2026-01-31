package services

import (
	"codeagent-backend/models"
	"codeagent-backend/utils"
	"context"
	"fmt"
)

type LLMTestCaseService struct {
	LLMService *LLMService
}

func (s *LLMTestCaseService) GetLLMTestCases(projectID, promptID string, page, pageSize int) ([]models.LLMTestCase, int64, error) {
	var testCases []models.LLMTestCase
	var total int64

	query := utils.DB.Model(&models.LLMTestCase{})

	if projectID != "" {
		query = query.Joins("JOIN prompts ON prompts.id = llm_test_cases.prompt_id").Where("prompts.project_id = ?", projectID)
	}
	if promptID != "" {
		query = query.Where("prompt_id = ?", promptID)
	}

	query.Count(&total)
	err := query.Order("llm_test_cases.id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&testCases).Error
	return testCases, total, err
}

func (s *LLMTestCaseService) DeleteLLMTestCase(id string) error {
	return utils.DB.Delete(&models.LLMTestCase{}, id).Error
}

func (s *LLMTestCaseService) BatchDeleteLLMTestCases(ids []uint) error {
	return utils.DB.Delete(&models.LLMTestCase{}, ids).Error
}

func (s *LLMTestCaseService) UpdateLLMTestCase(testCase *models.LLMTestCase) error {
	return utils.DB.Save(testCase).Error
}

func (s *LLMTestCaseService) GetLLMTestCase(id string) (*models.LLMTestCase, error) {
	var testCase models.LLMTestCase
	err := utils.DB.First(&testCase, id).Error
	return &testCase, err
}

func (s *LLMTestCaseService) GenerateLLMTestCases(ctx context.Context, configID uint, promptIDs []uint, count int) ([]models.LLMTestCase, error) {
	var config models.LLMConfig
	if err := utils.DB.First(&config, configID).Error; err != nil {
		return nil, err
	}

	var allCreatedCases []models.LLMTestCase

	for _, promptID := range promptIDs {
		var prompt models.Prompt
		if err := utils.DB.First(&prompt, promptID).Error; err != nil {
			continue
		}

		generatedCases, err := s.LLMService.GenerateTestCases(ctx, config, prompt.Content, count)
		if err != nil {
			continue
		}

		for _, generated := range generatedCases {
			testCase := models.LLMTestCase{
				PromptID: prompt.ID,
				Input:    generated.Input,
			}
			utils.DB.Create(&testCase)
			allCreatedCases = append(allCreatedCases, testCase)
		}
	}
	return allCreatedCases, nil
}

func (s *LLMTestCaseService) RunLLMTestCases(testCaseIDs []uint, configID uint) (string, error) {
	var config models.LLMConfig
	if err := utils.DB.First(&config, configID).Error; err != nil {
		return "", err
	}

	taskID := GlobalTaskManager.StartTask(len(testCaseIDs), func(ctx context.Context, updateProgress func(int, string) error) error {
		for i, id := range testCaseIDs {
			if err := updateProgress(i, fmt.Sprintf("Running test case %d/%d", i+1, len(testCaseIDs))); err != nil {
				return err
			}

			var testCase models.LLMTestCase
			if err := utils.DB.First(&testCase, id).Error; err != nil {
				continue
			}

			var prompt models.Prompt
			utils.DB.First(&prompt, testCase.PromptID)

			output, err := s.LLMService.RunPrompt(ctx, config, prompt.Content, testCase.Input)
			if err == nil {
				testCase.Output = output
				utils.DB.Save(&testCase)
			}
		}
		return nil
	})

	return taskID, nil
}

func (s *LLMTestCaseService) EvaluateLLMTestCases(testCaseIDs []uint, configID uint) (string, error) {
	var config models.LLMConfig
	if err := utils.DB.First(&config, configID).Error; err != nil {
		return "", err
	}

	taskID := GlobalTaskManager.StartTask(len(testCaseIDs), func(ctx context.Context, updateProgress func(int, string) error) error {
		for i, id := range testCaseIDs {
			if err := updateProgress(i, fmt.Sprintf("Evaluating test case %d/%d", i+1, len(testCaseIDs))); err != nil {
				return err
			}

			var testCase models.LLMTestCase
			if err := utils.DB.First(&testCase, id).Error; err != nil {
				continue
			}

			if testCase.Output == "" {
				continue
			}

			var prompt models.Prompt
			utils.DB.First(&prompt, testCase.PromptID)

			reason, isPass, err := s.LLMService.EvaluateTestCase(ctx, config, prompt.Content, testCase.Input, testCase.Output)
			if err == nil {
				testCase.Evaluation = reason
				testCase.IsPass = isPass
				utils.DB.Save(&testCase)
			}
		}
		return nil
	})

	return taskID, nil
}

func (s *LLMTestCaseService) RunLLMTestCasesFromDefinitions(promptID, configID uint) (string, error) {
	var prompt models.Prompt
	if err := utils.DB.First(&prompt, promptID).Error; err != nil {
		return "", err
	}

	var config models.LLMConfig
	if err := utils.DB.First(&config, configID).Error; err != nil {
		return "", err
	}

	var testCases []models.TestCase
	if err := utils.DB.Joins("JOIN prompts ON prompts.id = test_cases.prompt_id").
		Where("prompts.project_id = ?", prompt.ProjectID).
		Find(&testCases).Error; err != nil {
		return "", err
	}

	if len(testCases) == 0 {
		return "", fmt.Errorf("no test cases found for this project")
	}

	taskID := GlobalTaskManager.StartTask(len(testCases), func(ctx context.Context, updateProgress func(int, string) error) error {
		for i, tc := range testCases {
			if err := updateProgress(i, fmt.Sprintf("Running and Evaluating %d/%d", i+1, len(testCases))); err != nil {
				return err
			}

			output, err := s.LLMService.RunPrompt(ctx, config, prompt.Content, tc.Input)
			if err != nil {
				output = "Error: " + err.Error()
			}

			reason, isPass, err := s.LLMService.EvaluateTestCase(ctx, config, prompt.Content, tc.Input, output)
			if err != nil {
				reason = "Evaluation Error: " + err.Error()
				isPass = false
			}

			result := models.LLMTestCase{
				PromptID:   promptID,
				Input:      tc.Input,
				Output:     output,
				Evaluation: reason,
				IsPass:     isPass,
			}
			utils.DB.Create(&result)
		}
		return nil
	})

	return taskID, nil
}
