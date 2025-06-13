package main

import (
	"fmt"
	"log"
	"os"

	"github.com/darkphotonKN/community-builds/config"
	"github.com/darkphotonKN/community-builds/internal/validation"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

/**
* Main entry point to entire application.
* NOTE: Keep code here as clean and little as possible.
**/
func main() {
	// env setup
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, using system environment variables")
	}

	// database setup
	db := config.InitDB()
	defer db.Close()

	// router setup
	router := config.SetupRouter()

	// Register custom validators
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		validation.RegisterValidators(v)
	}

	defaultDevPort := ":7001"
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultDevPort
	}

	// starts server and listen on port
	if err := router.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatal("Failed to start server")
	}
}
