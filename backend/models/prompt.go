package models

// Prompt stores prompt templates
type Prompt struct {
	BaseModel
	ProjectID uint   `json:"project_id"`
	Name      string `json:"name"`
	Content   string `gorm:"type:text" json:"content"`
	Tags      string `json:"tags"` // Comma separated tags
}
