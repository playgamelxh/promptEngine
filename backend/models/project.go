package models

// Project stores project information
type Project struct {
	BaseModel
	Name        string `json:"name"`
	Description string `json:"description"`
	Tags        string `json:"tags"` // Comma separated tags
}
