package auth

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/z-sk1/signin-api/internal/db"
)

var jwtKey []byte

func init() {
	jwtKey = getOrCreateSecret()

	fmt.Printf("jwt key: %s...", string(jwtKey)[:5])
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// jwt claims
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func SignUp(c *gin.Context) {
	var newUser User
	if err := c.BindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// check if user exists

	var exists int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", newUser.Username).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	if exists > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user already exists"})
		return
	}

	// insert new user
	_, err = db.DB.Exec("INSERT INTO users(username, password) VALUES(?, ?)", newUser.Username, newUser.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user created succesfully"})
}

func Login(c *gin.Context) {
	var creds User

	if err := c.BindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	var storedPassword string
	err := db.DB.QueryRow("SELECT password FROM users WHERE username = ?", creds.Username).Scan(&storedPassword)
	if err == sql.ErrNoRows || storedPassword != creds.Password {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid username or password"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: creds.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(jwtKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not create token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenStr})
}

func DeleteAccount(c *gin.Context) {
	tokenStr := c.GetHeader("Authorization")
	if tokenStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		c.Abort()
		return
	}

	if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
		tokenStr = tokenStr[7:]
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		c.Abort()
		return
	}

	username := claims.Username

	_, err = db.DB.Exec("DELETE FROM users WHERE username = ?", username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "account deleted successfully"})
}

func RequireAuth(c *gin.Context) {
	tokenStr := c.GetHeader("Authorization")
	if tokenStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		c.Abort()
		return
	}

	if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
		tokenStr = tokenStr[7:]
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		c.Abort()
		return
	}

	c.Set("username", claims.Username)
	c.Next()
}

func Me(c *gin.Context) {
	username := c.GetString("username")
	c.JSON(http.StatusOK, gin.H{"username": username})
}
