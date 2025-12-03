package reminders

import (
	"net/http"

	"time"

	"fmt"

	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/z-sk1/signin-api/internal/auth"
	"github.com/z-sk1/signin-api/internal/db"
)

type Reminder struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Due       string    `json:"due"`
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

	dueTime, err := time.Parse(time.RFC3339, reminder.Due)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid due date"})
		return
	}

	_, err = db.DB.Exec("INSERT INTO reminders(username, title, content, due, user_id) VALUES (?, ?, ?, ?, ?)", username, reminder.Title, reminder.Content, dueTime, userID)
	if err != nil {
		fmt.Println("Error inserting reminder:", err)
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
			fmt.Println("Error reading reminders:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not read reminders"})
			return
		}
		reminder.Username = username
		reminders = append(reminders, reminder)
	}

	c.JSON(http.StatusOK, gin.H{"reminders": reminders})
}

func GetReminderCount(c *gin.Context) {
	username := c.GetString("username")

	// find user id 
	var userID int 
	err := db.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not find user"})
		return
	}

	var total int
	err = db.DB.QueryRow("SELECT COUNT(*) AS total FROM reminders WHERE user_id = ?", userID).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not get reminder count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"total": total})
}

func DeleteReminder(c *gin.Context) {
	tokenStr := c.GetHeader("Authorization")

	if tokenStr == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "missing token"})
		c.Abort()
		return
	}

	if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
		tokenStr = tokenStr[7:]
	}

	// Parse JWT
	claims := &auth.Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return auth.JwtKey, nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	username := claims.Username

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid reminder id"})
		return
	}

	_, err = db.DB.Exec("DELETE FROM reminders WHERE id = ? AND username = ?", id, username)
	if err != nil {
		fmt.Println("Error deleting reminder:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete reminder"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "reminder deleted succesfully"})
}

func UpdateReminder(c *gin.Context) {
	username := c.GetString("username")
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid reminder id"})
		return
	}

	var reminder struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		Due     string `json:"due"`
	}
	if err := c.BindJSON(&reminder); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	dueTime, err := time.Parse(time.RFC3339, reminder.Due)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid due date"})
		return
	}

	res, err := db.DB.Exec("UPDATE reminders SET title = ?, content = ?, due = ? WHERE id = ? AND user_id = (SELECT id FROM users WHERE username = ?)", reminder.Title, reminder.Content, dueTime, id, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update reminder"})
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "reminder not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "reminder updated successfully"})
}
