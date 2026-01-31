package controllers

import (
	"codeagent-backend/models"
	"codeagent-backend/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var testCaseService = new(services.TestCaseService)

func CreateTestCase(c *gin.Context) {
	var testCase models.TestCase
	if err := c.ShouldBindJSON(&testCase); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	created, err := testCaseService.CreateTestCase(&testCase)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !created {
		c.JSON(http.StatusOK, gin.H{"message": "Duplicate test case skipped", "skipped": true})
		return
	}

	c.JSON(http.StatusOK, testCase)
}

type GenerateRequest struct {
	PromptIDs []uint `json:"prompt_ids"`
	ConfigID  uint   `json:"config_id"`
	Count     int    `json:"count"`
}

func GenerateTestCases(c *gin.Context) {
	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	allCreatedCases, err := testCaseService.GenerateTestCases(c.Request.Context(), req.ConfigID, req.PromptIDs, req.Count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, allCreatedCases)
}

func GetTestCases(c *gin.Context) {
	projectID := c.Query("project_id")
	promptID := c.Query("prompt_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "30"))
	if pageSize > 30 {
		pageSize = 30
	}

	testCases, total, err := testCaseService.GetTestCases(projectID, promptID, page, pageSize)
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

func UpdateTestCase(c *gin.Context) {
	testCase, err := testCaseService.GetTestCase(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "TestCase not found"})
		return
	}

	// Backup original input to check if it changed
	originalInput := testCase.Input

	if err := c.ShouldBindJSON(testCase); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := testCaseService.UpdateTestCase(testCase, originalInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !updated {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Duplicate test case input in this project"})
		return
	}

	c.JSON(http.StatusOK, testCase)
}

func DeleteTestCase(c *gin.Context) {
	testCase, err := testCaseService.GetTestCase(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "TestCase not found"})
		return
	}

	if err := testCaseService.DeleteTestCase(testCase); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "TestCase deleted"})
}

func BatchDeleteTestCases(c *gin.Context) {
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

	if err := testCaseService.BatchDeleteTestCases(req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "TestCases deleted"})
}
