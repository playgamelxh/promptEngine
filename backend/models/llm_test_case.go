package models

// LLMTestCase stores generated test cases
type LLMTestCase struct {
	BaseModel
	PromptID   uint   `json:"prompt_id"`
	Input      string `gorm:"type:text" json:"input"`
	Output     string `gorm:"type:text" json:"output"`
	Evaluation string `gorm:"type:text" json:"evaluation"` // JSON or text evaluation result
	IsPass     bool   `json:"is_pass"`
}
