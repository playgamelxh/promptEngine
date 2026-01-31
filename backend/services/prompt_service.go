package services

import (
	"codeagent-backend/models"
	"codeagent-backend/utils"
	"context"
)

type PromptService struct {
	LLMService *LLMService
}

func (s *PromptService) CreatePrompt(prompt *models.Prompt) error {
	return utils.DB.Create(prompt).Error
}

func (s *PromptService) GetPrompts(projectID string, page, pageSize int) ([]models.Prompt, int64, error) {
	var prompts []models.Prompt
	var total int64

	query := utils.DB.Model(&models.Prompt{})
	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}

	query.Count(&total)
	err := query.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&prompts).Error
	return prompts, total, err
}

func (s *PromptService) GetPrompt(id string) (*models.Prompt, error) {
	var prompt models.Prompt
	err := utils.DB.First(&prompt, id).Error
	return &prompt, err
}

func (s *PromptService) UpdatePrompt(prompt *models.Prompt) error {
	return utils.DB.Save(prompt).Error
}

func (s *PromptService) DeletePrompt(prompt *models.Prompt) error {
	return utils.DB.Delete(prompt).Error
}

func (s *PromptService) BatchDeletePrompts(ids []uint) error {
	return utils.DB.Delete(&models.Prompt{}, ids).Error
}

func (s *PromptService) BatchGeneratePrompts(ctx context.Context, configID uint, instruction string, count int, projectID uint) ([]models.Prompt, error) {
	var config models.LLMConfig
	if err := utils.DB.First(&config, configID).Error; err != nil {
		return nil, err
	}

	generatedPrompts, err := s.LLMService.GeneratePrompts(ctx, config, instruction, count)
	if err != nil {
		return nil, err
	}

	var createdPrompts []models.Prompt
	for _, generated := range generatedPrompts {
		prompt := models.Prompt{
			ProjectID: projectID,
			Name:      generated.Name,
			Content:   generated.Content,
			Tags:      generated.Tags,
		}
		utils.DB.Create(&prompt)
		createdPrompts = append(createdPrompts, prompt)
	}

	return createdPrompts, nil
}
