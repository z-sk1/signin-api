package expenses

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

type Expense struct {
	ID       int     `json:"id"`
	Username string  `json:"username"`
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
	Date     string  `json:"date"`
	Note     string  `json:"note"`
}

func CreateExpense(c *gin.Context) {
	username := c.GetString("username")
	var expense Expense

	if err := c.BindJSON(&expense); err != nil {
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

	date, err := time.Parse(time.RFC3339, expense.Date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid date"})
		return
	}

	_, err = db.DB.Exec("INSERT INTO expenses(username, amount, category, date, note, user_id) VALUES (?, ?, ?, ?, ?, ?)", username, expense.Amount, expense.Category, date, expense.Note, userID)
	if err != nil {
		fmt.Println("Error insterting expense", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save expense"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "expense saved successfully"})
}

func GetAllExpenses(c *gin.Context) {
	username := c.GetString("username")

	// find user id
	var userID int
	err := db.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not find user"})
		return
	}

	// get all expenses for user
	rows, err := db.DB.Query("SELECT id, amount, category, date, note FROM expenses WHERE user_id = ?", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not find expenses"})
		return
	}
	defer rows.Close()

	var expenses []Expense
	for rows.Next() {
		var expense Expense
		if err := rows.Scan(&expense.ID, &expense.Amount, &expense.Category, &expense.Date, &expense.Note); err != nil {
			fmt.Println("Error reading expenses:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not read expenses"})
			return
		}
		expense.Username = username
		expenses = append(expenses, expense)
	}

	c.JSON(http.StatusOK, gin.H{"expenses": expenses})
}

func DeleteExpense(c *gin.Context) {
	tokenStr := c.GetHeader("Authorization")

	if tokenStr == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "missing token"})
		c.Abort()
		return
	}

	if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
		tokenStr = tokenStr[7:]
	}

	// parse JWT
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expense id"})
		return
	}

	_, err = db.DB.Exec("DELETE FROM expenses WHERE id = ? AND username = ?", id, username)
	if err != nil {
		fmt.Println("Error deleting expense:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete expense"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "expense deleted succesfully"})
}

func UpdateExpense(c *gin.Context) {
	username := c.GetString("username")
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expense id"})
		return
	}

	var expense struct {
		Amount   float64 `json:"amount"`
		Category string  `json:"category"`
		Date     string  `json:"date"`
		Note     string  `json:"note"`
	}

	if err := c.BindJSON(&expense); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	date, err := time.Parse(time.RFC3339, expense.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date"})
		return
	}

	res, err := db.DB.Exec("UPDATE expenses SET amount = ?, category = ?, date = ?, note = ? WHERE id = ? AND user_id = (SELECT id FROM users WHERE username = ?)", expense.Amount, expense.Category, date, expense.Note, id, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update expense"})
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "expense not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "expense updated successfully"})
}
