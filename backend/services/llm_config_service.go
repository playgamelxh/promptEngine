package services

import (
	"codeagent-backend/models"
	"codeagent-backend/utils"
)

type LLMConfigService struct{}

func (s *LLMConfigService) CreateLLMConfig(config *models.LLMConfig) error {
	tx := utils.DB.Begin()

	if config.IsDefault {
		// Set all others to false
		if err := tx.Model(&models.LLMConfig{}).Where("1 = 1").Update("is_default", false).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Create(config).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *LLMConfigService) GetLLMConfigs(page, pageSize int) ([]models.LLMConfig, int64, error) {
	var configs []models.LLMConfig
	var total int64

	query := utils.DB.Model(&models.LLMConfig{})
	query.Count(&total)

	err := query.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&configs).Error
	return configs, total, err
}

func (s *LLMConfigService) GetLLMConfig(id string) (*models.LLMConfig, error) {
	var config models.LLMConfig
	err := utils.DB.First(&config, id).Error
	return &config, err
}

func (s *LLMConfigService) UpdateLLMConfig(config *models.LLMConfig, input models.LLMConfig) error {
	tx := utils.DB.Begin()

	// Update fields
	config.Name = input.Name
	config.APIKey = input.APIKey
	config.BaseURL = input.BaseURL
	config.ModelName = input.ModelName
	config.Temperature = input.Temperature
	config.Tags = input.Tags

	// If setting to default, unset others
	if input.IsDefault {
		config.IsDefault = true
		if err := tx.Model(&models.LLMConfig{}).Where("id != ?", config.ID).Update("is_default", false).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		config.IsDefault = false
	}

	if err := tx.Save(config).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *LLMConfigService) DeleteLLMConfig(config *models.LLMConfig) error {
	return utils.DB.Delete(config).Error
}

func (s *LLMConfigService) BatchDeleteLLMConfigs(ids []uint) error {
	return utils.DB.Delete(&models.LLMConfig{}, ids).Error
}
