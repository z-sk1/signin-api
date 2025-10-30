package notes

import (
	"net/http"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/z-sk1/signin-api/internal/db"
)

type Note struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

func CreateNote(c *gin.Context) {
	username := c.GetString("username")
	var note Note

	if err := c.BindJSON(&note); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// find user_id from username
	var userID int
	err := db.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not find user"})
		return
	}

	_, err = db.DB.Exec("INSERT INTO notes(username, title, content) VALUES (?, ?, ?)", username, note.Title, note.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save note"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "note created succesfully"})
}

func GetAllNotes(c *gin.Context) {
	username := c.GetString("username")

	// find user id 
	var userID int
	err := db.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not find user"})
		return
	}

	// get all notes for user
	rows, err := db.DB.Query("SELECT id, title, content, created_at FROM notes WHERE user_id = ?", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not find notes"})
		return
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var note Note
		if err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not read notes"})
			return
		}
		note.Username = username 
		notes = append(notes, note)
	}

	c.JSON(http.StatusOK, gin.H{"notes": notes})
}