package controllers

import (
	"codeagent-backend/models"
	"codeagent-backend/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var llmTestCaseService = &services.LLMTestCaseService{LLMService: new(services.LLMService)}

// GenerateRequest is defined in test_case.go

func GenerateLLMTestCases(c *gin.Context) {
	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	allCreatedCases, err := llmTestCaseService.GenerateLLMTestCases(c.Request.Context(), req.ConfigID, req.PromptIDs, req.Count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, allCreatedCases)
}

type RunRequest struct {
	TestCaseIDs []uint `json:"test_case_ids"`
	ConfigID    uint   `json:"config_id"`
}

func RunLLMTestCases(c *gin.Context) {
	var req RunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	taskID, err := llmTestCaseService.RunLLMTestCases(req.TestCaseIDs, req.ConfigID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"task_id": taskID, "message": "Run started"})
}

func EvaluateLLMTestCases(c *gin.Context) {
	var req RunRequest // Reuse RunRequest structure
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	taskID, err := llmTestCaseService.EvaluateLLMTestCases(req.TestCaseIDs, req.ConfigID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"task_id": taskID, "message": "Evaluation started"})
}

func GetLLMTestCases(c *gin.Context) {
	projectID := c.Query("project_id")
	promptID := c.Query("prompt_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "30"))
	if pageSize > 30 {
		pageSize = 30
	}

	testCases, total, err := llmTestCaseService.GetLLMTestCases(projectID, promptID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":     testCases,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func DeleteLLMTestCase(c *gin.Context) {
	if err := llmTestCaseService.DeleteLLMTestCase(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}

func BatchDeleteLLMTestCases(c *gin.Context) {
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

	if err := llmTestCaseService.BatchDeleteLLMTestCases(req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "LLMTestCases deleted"})
}

type RunFromDefinitionsRequest struct {
	PromptID uint `json:"prompt_id"`
	ConfigID uint `json:"config_id"`
}

func RunLLMTestCasesFromDefinitions(c *gin.Context) {
	var req RunFromDefinitionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	taskID, err := llmTestCaseService.RunLLMTestCasesFromDefinitions(req.PromptID, req.ConfigID)
	if err != nil {
		if err.Error() == "no test cases found for this project" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"task_id": taskID, "message": "Run from definitions started"})
}

func UpdateLLMTestCase(c *gin.Context) {
	testCase, err := llmTestCaseService.GetLLMTestCase(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "LLMTestCase not found"})
		return
	}

	var input models.LLMTestCase
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Only allow updating specific fields if needed, but for now allow full update
	testCase.IsPass = input.IsPass
	// Allow updating evaluation text too if user wants to add notes
	if input.Evaluation != "" {
		testCase.Evaluation = input.Evaluation
	}

	if err := llmTestCaseService.UpdateLLMTestCase(testCase); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, testCase)
}
