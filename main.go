package main

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/z-sk1/signin-api/internal/auth"
	"github.com/z-sk1/signin-api/internal/db"
	"github.com/z-sk1/signin-api/internal/remind-me/notes"
	"github.com/z-sk1/signin-api/internal/remind-me/reminders"
)

func main() {
	db.InitDB()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // or ["http://localhost:5173"] if you want to be strict
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

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

		// reminders
		authRoutes.POST("/reminders", reminders.CreateReminder)
		authRoutes.GET("/reminders", reminders.GetAllReminders)
	}

	r.Run(":8080")
}
