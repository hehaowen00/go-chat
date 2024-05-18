package main

import (
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("ADDR")

	r := gin.Default()

	r.Run(":" + port)
}
