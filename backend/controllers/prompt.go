package controllers

import (
	"codeagent-backend/models"
	"codeagent-backend/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var promptService = &services.PromptService{LLMService: new(services.LLMService)}

func CreatePrompt(c *gin.Context) {
	var prompt models.Prompt
	if err := c.ShouldBindJSON(&prompt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := promptService.CreatePrompt(&prompt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, prompt)
}

func GetPrompts(c *gin.Context) {
	projectID := c.Query("project_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "30"))
	if pageSize > 30 {
		pageSize = 30
	}

	prompts, total, err := promptService.GetPrompts(projectID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":     prompts,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func UpdatePrompt(c *gin.Context) {
	prompt, err := promptService.GetPrompt(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Prompt not found"})
		return
	}

	if err := c.ShouldBindJSON(prompt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := promptService.UpdatePrompt(prompt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, prompt)
}

func DeletePrompt(c *gin.Context) {
	prompt, err := promptService.GetPrompt(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Prompt not found"})
		return
	}

	if err := promptService.DeletePrompt(prompt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Prompt deleted"})
}

func BatchDeletePrompts(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No IDs provided"})
		return
	}

	if err := promptService.BatchDeletePrompts(req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Prompts deleted"})
}

func BatchGeneratePrompts(c *gin.Context) {
	var req struct {
		ConfigID    uint   `json:"config_id"`
		Instruction string `json:"instruction"`
		Count       int    `json:"count"`
		ProjectID   uint   `json:"project_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdPrompts, err := promptService.BatchGeneratePrompts(c.Request.Context(), req.ConfigID, req.Instruction, req.Count, req.ProjectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, createdPrompts)
}
