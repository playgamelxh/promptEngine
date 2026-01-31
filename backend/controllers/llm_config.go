package controllers

import (
	"codeagent-backend/models"
	"codeagent-backend/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var llmConfigService = new(services.LLMConfigService)

func CreateLLMConfig(c *gin.Context) {
	var config models.LLMConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := llmConfigService.CreateLLMConfig(&config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

func GetLLMConfigs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "30"))
	if pageSize > 30 {
		pageSize = 30
	}

	configs, total, err := llmConfigService.GetLLMConfigs(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":     configs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func UpdateLLMConfig(c *gin.Context) {
	config, err := llmConfigService.GetLLMConfig(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
		return
	}

	var input models.LLMConfig
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := llmConfigService.UpdateLLMConfig(config, input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

func DeleteLLMConfig(c *gin.Context) {
	config, err := llmConfigService.GetLLMConfig(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
		return
	}

	if err := llmConfigService.DeleteLLMConfig(config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Config deleted"})
}

func BatchDeleteLLMConfigs(c *gin.Context) {
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

	if err := llmConfigService.BatchDeleteLLMConfigs(req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Configs deleted"})
}
