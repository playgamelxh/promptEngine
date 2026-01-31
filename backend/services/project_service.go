package services

import (
	"codeagent-backend/models"
	"codeagent-backend/utils"
)

type ProjectService struct{}

func (s *ProjectService) CreateProject(project *models.Project) error {
	return utils.DB.Create(project).Error
}

func (s *ProjectService) GetProjects(page, pageSize int) ([]models.Project, int64, error) {
	var projects []models.Project
	var total int64

	query := utils.DB.Model(&models.Project{})
	query.Count(&total)

	err := query.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&projects).Error
	return projects, total, err
}

func (s *ProjectService) GetProject(id string) (*models.Project, error) {
	var project models.Project
	err := utils.DB.First(&project, id).Error
	return &project, err
}

func (s *ProjectService) UpdateProject(project *models.Project) error {
	return utils.DB.Save(project).Error
}

func (s *ProjectService) DeleteProject(project *models.Project) error {
	return utils.DB.Delete(project).Error
}

func (s *ProjectService) BatchDeleteProjects(ids []uint) error {
	return utils.DB.Delete(&models.Project{}, ids).Error
}
