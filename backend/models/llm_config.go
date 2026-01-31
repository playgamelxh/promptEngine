package models

// LLMConfig stores configuration for LLM interfaces
type LLMConfig struct {
	BaseModel
	Name        string  `json:"name"`
	APIKey      string  `json:"api_key"`
	BaseURL     string  `json:"base_url"`
	ModelName   string  `json:"model_name"`
	Temperature float64 `json:"temperature" gorm:"default:0.7"`
	Tags        string  `json:"tags"` // Comma separated tags
	IsDefault   bool    `json:"is_default" gorm:"default:false"`
}
