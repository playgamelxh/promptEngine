package controllers

import (
	"codeagent-backend/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetTaskStatus(c *gin.Context) {
	taskID := c.Query("task_id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task_id is required"})
		return
	}

	task, exists := services.GlobalTaskManager.GetTask(taskID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func StopTask(c *gin.Context) {
	var req struct {
		TaskID string `json:"task_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	services.GlobalTaskManager.StopTask(req.TaskID)
	c.JSON(http.StatusOK, gin.H{"message": "Task stopped"})
}
