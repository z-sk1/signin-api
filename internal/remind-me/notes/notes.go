package notes

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

	_, err = db.DB.Exec("INSERT INTO notes(username, title, content, user_id) VALUES (?, ?, ?, ?)", username, note.Title, note.Content, userID)
	if err != nil {
		fmt.Println("Error inserting note:", err)
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
			fmt.Println("Error reading notes:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not read notes"})
			return
		}
		note.Username = username
		notes = append(notes, note)
	}

	c.JSON(http.StatusOK, gin.H{"notes": notes})
}

func DeleteNote(c *gin.Context) {
	tokenStr := c.GetHeader("Authorization")
	if tokenStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		c.Abort()
		return
	}

	if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
		tokenStr = tokenStr[7:]
	}

	// Parse JWT
	claims := &auth.Claims{} // assuming you have your Claims struct in package "auth"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid note id"})
		return
	}

	_, err = db.DB.Exec("DELETE FROM notes WHERE id = ? AND username = ?", id, username)
	if err != nil {
		fmt.Println("Error deleting note:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete note"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "note deleted succesfully"})
}

func UpdateNote(c *gin.Context) {
	username := c.GetString("username")
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid note id"})
		return
	}

	var note struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := c.BindJSON(&note); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// make sure the note belongs to the user
	res, err := db.DB.Exec("UPDATE notes SET title = ?, content = ? WHERE id = ? AND user_id = (SELECT id FROM users WHERE username = ?)", note.Title, note.Content, id, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update note"})
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "note updated successfully"})
}
