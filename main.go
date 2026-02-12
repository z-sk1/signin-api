package main

import (
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/z-sk1/signin-api/internal/auth"
	leaderboard "github.com/z-sk1/signin-api/internal/bateenfest"
	"github.com/z-sk1/signin-api/internal/db"
	"github.com/z-sk1/signin-api/internal/keep-track/expenses"
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
	r.POST("/forgot-password", auth.ForgotPassword)
	r.POST("/reset-password", auth.ResetPassword)

	// unprotected routes
	r.GET("/leaderboard/:section", leaderboard.GetAllLeaderboardScores)

	// protected routes
	authRoutes := r.Group("/")
	authRoutes.Use(auth.RequireAuth)
	{
		authRoutes.GET("/me", auth.Me)
		authRoutes.DELETE("/delete", auth.DeleteAccount)

		// admin
		adminRoutes := r.Group("/admin")
		adminRoutes.Use(auth.RequireAdmin)
		{
			adminRoutes.POST("/leaderboard", leaderboard.AddLeaderboardScore)
			adminRoutes.DELETE("/leaderboard/:id", leaderboard.DeleteLeaderboardScore)
			adminRoutes.PUT("/leaderboard/:id", leaderboard.UpdateLeaderboardScore)
		}

		// notes
		authRoutes.POST("/notes", notes.CreateNote)
		authRoutes.GET("/notes", notes.GetAllNotes)
		authRoutes.GET("/notes/total", notes.GetNoteCount)
		authRoutes.DELETE("/notes/:id", notes.DeleteNote)
		authRoutes.PUT("/notes/:id", notes.UpdateNote)

		// reminders
		authRoutes.POST("/reminders", reminders.CreateReminder)
		authRoutes.GET("/reminders", reminders.GetAllReminders)
		authRoutes.GET("/reminders/total", reminders.GetReminderCount)
		authRoutes.DELETE("/reminders/:id", reminders.DeleteReminder)
		authRoutes.PUT("/reminders/:id", reminders.UpdateReminder)

		// expenses
		authRoutes.POST("/expenses", expenses.CreateExpense)

		authRoutes.GET("/expenses", expenses.GetAllExpenses)
		authRoutes.GET("/expenses/total", expenses.GetTotalExpenses)
		authRoutes.GET("/expenses/categories", expenses.GetExpenseCategories)
		authRoutes.DELETE("/expenses/:id", expenses.DeleteExpense)
		authRoutes.PUT("/expenses/:id", expenses.UpdateExpense)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
