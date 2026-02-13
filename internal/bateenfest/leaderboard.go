package leaderboard

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/z-sk1/signin-api/internal/db"
)

type LeaderboardEntry struct {
	ID      int    `json:"id"`
	Section string `json:"section"`
	Name    string `json:"name"`
	Points  int    `json:"points"`
	Rank    int    `json:"rank"`
}

func AddLeaderboardScore(c *gin.Context) {
	var entry LeaderboardEntry

	if err := c.BindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if entry.Points < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "points must be positive"})
		return
	}

	username := c.GetString("username")
	var userID int
	err := db.DB.QueryRow("SELECT id FROM users WHERE username = $1", username).Scan(&userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not find user"})
		return
	}

	_, err = db.DB.Exec(`
		INSERT INTO leaderboard (user_id, username, section, name, points)
		VALUES ($1, $2, $3, $4, $5)
	`, userID, username, entry.Section, entry.Name, entry.Points)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add score"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "score added"})
}

func GetAllLeaderboardScores(c *gin.Context) {
	section := c.Param("section")

	rows, err := db.DB.Query(`
		SELECT id, name, points
		FROM leaderboard
		WHERE section = $1
		ORDER BY points DESC
	`, section)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch leaderboard"})
		return
	}
	defer rows.Close()

	var entries []LeaderboardEntry
	rank := 1

	for rows.Next() {
		var entry LeaderboardEntry
		rows.Scan(&entry.ID, &entry.Name, &entry.Points)
		entry.Rank = rank
		rank++
		entries = append(entries, entry)
	}

	c.JSON(http.StatusOK, entries)
}

func DeleteLeaderboardScore(c *gin.Context) {
	id := c.Param("id")

	_, err := db.DB.Exec("DELETE FROM leaderboard WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete score"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "score deleted"})
}

func UpdateLeaderboardScore(c *gin.Context) {
	id := c.Param("id")

	var entry struct {
		Section string `json:"section"`
		Name    string `json:"name"`
		Points  int    `json:"points"`
	}

	if err := c.BindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	_, err := db.DB.Exec(`
		UPDATE leaderboard
		SET section = $1, name = $2, points = $3
		WHERE id = $4
	`, entry.Section, entry.Name, entry.Points, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update score"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "score updated"})
}
