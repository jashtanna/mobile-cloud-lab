package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/joho/godotenv"
)

func init() {
	if os.Getenv("NODE_ENV") == "local" {
		if err := godotenv.Load(); err != nil {
			fmt.Printf("Error loading .env: %s\n", err.Error())
		}
	}
}

func main() {
	functionName := os.Getenv("AWS_LAMBDA_FUNCTION_NAME")
	switch functionName {
	case "UploadProductFunction":
		lambda.Start(uploadHandler)
	case "GetAllProductsFunction":
		lambda.Start(getAllHandler)
	default:
		fmt.Println("Unknown function name:", functionName)
		os.Exit(1)
	}
}
