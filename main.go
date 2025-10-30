package main

import (
	"github.com/gin-gonic/gin"
	"github.com/z-sk1/signin-api/internal/auth"
	"github.com/z-sk1/signin-api/internal/db"
	"github.com/z-sk1/signin-api/internal/remind-me/notes"
	"github.com/z-sk1/signin-api/internal/remind-me/reminders"
)

func main() {
	db.InitDB()

	r := gin.Default()

	// auth routes
	r.POST("/signup", auth.SignUp)
	r.POST("/login", auth.Login)

	// protected routes 
	authRoutes := r.Group("/")
	authRoutes.Use(auth.RequireAuth) 
	{
		authRoutes.GET("/me", auth.Me)
		authRoutes.DELETE("/delete", auth.DeleteAccount)

		// notes 
		authRoutes.POST("/notes", notes.CreateNote)
		authRoutes.GET("/notes", notes.GetAllNotes)
	}

	r.Run(":8080")
}
