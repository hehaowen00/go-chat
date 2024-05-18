package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
	})

	r.POST("/register", func(c *gin.Context) {
		log.Println(c.ClientIP())
	})

	r.Run()
}
