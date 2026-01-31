package routes

import (
	"codeagent-backend/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	api := r.Group("/api")
	{
		// Project Routes
		api.POST("/projects", controllers.CreateProject)
		api.GET("/projects", controllers.GetProjects)
		api.PUT("/projects/:id", controllers.UpdateProject)
		api.DELETE("/projects/batch", controllers.BatchDeleteProjects)
		api.DELETE("/projects/:id", controllers.DeleteProject)

		// LLM Config Routes
		api.POST("/llm-configs", controllers.CreateLLMConfig)
		api.GET("/llm-configs", controllers.GetLLMConfigs)
		api.PUT("/llm-configs/:id", controllers.UpdateLLMConfig)
		api.DELETE("/llm-configs/batch", controllers.BatchDeleteLLMConfigs)
		api.DELETE("/llm-configs/:id", controllers.DeleteLLMConfig)

		// Prompt Routes
		api.POST("/prompts", controllers.CreatePrompt)
		api.POST("/prompts/generate", controllers.BatchGeneratePrompts)
		api.GET("/prompts", controllers.GetPrompts)
		api.PUT("/prompts/:id", controllers.UpdatePrompt)
		api.DELETE("/prompts/batch", controllers.BatchDeletePrompts)
		api.DELETE("/prompts/:id", controllers.DeletePrompt)

		// TestCase Routes
		api.POST("/test-cases", controllers.CreateTestCase)
		api.POST("/test-cases/generate", controllers.GenerateTestCases)
		api.GET("/test-cases", controllers.GetTestCases)
		api.PUT("/test-cases/:id", controllers.UpdateTestCase)
		api.DELETE("/test-cases/batch", controllers.BatchDeleteTestCases)
		api.DELETE("/test-cases/:id", controllers.DeleteTestCase)

		// LLM Test Case Routes
		api.POST("/llm-test-cases/generate", controllers.GenerateLLMTestCases)
		api.POST("/llm-test-cases/run", controllers.RunLLMTestCases)
		api.POST("/llm-test-cases/run-from-definitions", controllers.RunLLMTestCasesFromDefinitions)
		api.GET("/llm-test-cases/task/status", controllers.GetTaskStatus)
		api.POST("/llm-test-cases/task/stop", controllers.StopTask)
		api.POST("/llm-test-cases/evaluate", controllers.EvaluateLLMTestCases)
		api.GET("/llm-test-cases", controllers.GetLLMTestCases)
		api.PUT("/llm-test-cases/:id", controllers.UpdateLLMTestCase)
		api.DELETE("/llm-test-cases/batch", controllers.BatchDeleteLLMTestCases)
		api.DELETE("/llm-test-cases/:id", controllers.DeleteLLMTestCase)
	}

	return r
}
