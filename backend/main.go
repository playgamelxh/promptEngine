package main

import (
	"codeagent-backend/config"
	"codeagent-backend/routes"
	"codeagent-backend/utils"
	"fmt"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database
	utils.InitDB(cfg.DatabaseDSN)

	// Setup router
	r := routes.SetupRouter()

	// Start server
	fmt.Printf("Server starting on port %s\n", cfg.ServerPort)
	r.Run(":" + cfg.ServerPort)
}
