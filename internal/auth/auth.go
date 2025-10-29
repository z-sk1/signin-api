package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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

// in memory store for now
var users = map[string]string{}

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

	if _, exists := users[newUser.Username]; exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user already exists"})
		return
	}

	users[newUser.Username] = newUser.Password
	c.JSON(http.StatusOK, gin.H{"message": "user created succesfully"})
}

func Login(c *gin.Context) {
	var creds User

	if err := c.BindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	storedPassword, ok := users[creds.Username]
	if !ok || storedPassword != creds.Password {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid username or password"})
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

func RequireAuth(c *gin.Context) {
	tokenStr := c.GetHeader("Authorization")
	if tokenStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		c.Abort()
		return
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
