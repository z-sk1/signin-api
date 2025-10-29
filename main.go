package main

import (
	"github.com/gin-gonic/gin"
	"github.com/z-sk1/signin-api/internal/auth"
	"github.com/z-sk1/signin-api/internal/db"
)

func main() {
	db.InitDB()

	r := gin.Default()

	// auth routes
	r.POST("/signup", auth.SignUp)
	r.POST("/login", auth.Login)

	// Test protected route
	r.GET("/me", auth.RequireAuth, auth.Me)
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "hello"})
	})

	r.Run(":8080")
}
