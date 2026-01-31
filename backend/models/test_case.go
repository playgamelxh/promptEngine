package models

// TestCase stores manual test cases for prompts
type TestCase struct {
	BaseModel
	ProjectID      uint   `json:"project_id" gorm:"index"`
	PromptID       uint   `json:"prompt_id"`
	Input          string `gorm:"type:text" json:"input"`
	InputMD5       string `gorm:"size:32;index" json:"input_md5"`
	ExpectedOutput string `gorm:"type:text" json:"expected_output"`
	Tags           string `json:"tags"` // Comma separated tags
}
