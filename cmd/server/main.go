package main

import (
	"fmt"

	"github.com/meetsuhagiya/lifeline-backend/internal/database" // <--- Update this if your module name is different!
)

func main() {
	fmt.Println("🚀 Starting LifeLine Backend...")

	// 1. Initialize Database
	database.Connect()

	// 2. Keep the app running (for now)
	fmt.Println("Server is running on port 8080 (Simulation)")
	
    // Later, we will add the HTTP server here using Fiber or Gin
    // select {} // Blocks forever
}