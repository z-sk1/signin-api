package reminders

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/z-sk1/signin-api/internal/db"
)

type Reminder struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Due       time.Time `json:"due"`
	CreatedAt time.Time `json:"created_at"`
}

func CreateReminder(c *gin.Context) {
	username := c.GetString("username")
	var reminder Reminder

	if err := c.BindJSON(&reminder); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// find user_id from username
	var userID int
	err := db.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not find users"})
		return
	}

	_, err = db.DB.Exec("INSERT INTO reminders(username, title, content, due) (?, ?, ?, ?)", username, reminder.Title, reminder.Content, reminder.Due)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save reminder"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "reminder created succesfully"})
}

func GetAllReminders(c *gin.Context) {
	username := c.GetString("username")

	// find user id
	var userID int
	err := db.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not find user"})
		return
	}

	// get all reminders for user
	rows, err := db.DB.Query("SELECT id, title, content, due, created_at FROM reminders WHERE user_id = ?", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not find reminders"})
		return
	}
	defer rows.Close()

	var reminders []Reminder
	for rows.Next() {
		var reminder Reminder
		if err := rows.Scan(&reminder.ID, &reminder.Title, &reminder.Content, &reminder.Due, &reminder.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not read reminders"})
			return
		}
		reminder.Username = username
		reminders = append(reminders, reminder)
	}

	c.JSON(http.StatusOK, gin.H{"reminders": reminders})
}
