package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/z-sk1/signin-api/internal/db"
	"golang.org/x/crypto/bcrypt"
)

var JwtKey []byte

func init() {
	JwtKey = getOrCreateSecret()

	fmt.Printf("jwt key: %s...", string(JwtKey)[:5])
}

type User struct {
	Email    string `json:"email"`
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
	err := db.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = ? OR email = ?", newUser.Username, newUser.Email).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	if exists > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user already exists"})
		return
	}

	// hash the password
	hashed, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not hash password"})
		return
	}

	// insert new user
	_, err = db.DB.Exec("INSERT INTO users(email, username, password) VALUES(?, ?, ?)", newUser.Email, newUser.Username, string(hashed))
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

	var storedHashPassword string
	var actualUsername string
	var storedEmail string

	identifier := creds.Username
	if identifier == "" {
		identifier = creds.Email
	}

	err := db.DB.QueryRow("SELECT username, email, password FROM users WHERE username = ? OR email = ?", identifier, identifier).Scan(&actualUsername, &storedEmail, &storedHashPassword)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid username/email or password"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	// verify hashed password
	err = bcrypt.CompareHashAndPassword([]byte(storedHashPassword), []byte(creds.Password))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid username/email or password"})
		return
	}

	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		Username: actualUsername,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(JwtKey)
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
		return JwtKey, nil
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

func ForgotPassword(c *gin.Context) {
	var body struct {
		Email string `json:"email"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	var exists int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", body.Email).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	if exists == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "If this email exists, a reset link was sent"})
		return
	}

	// generate token
	tokenBytes := make([]byte, 32)
	rand.Read(tokenBytes)
	token := hex.EncodeToString(tokenBytes)

	// hash token for storage
	hash, _ := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)

	expiresAt := time.Now().Add(15 * time.Minute).Unix()

	// store token
	db.DB.Exec("INSERT INTO password_resets(email, token_hash, expires_at) VALUES(?, ?, ?)", body.Email, string(hash), expiresAt)

	// send token for testing
	c.JSON(http.StatusOK, gin.H{
		"reset_token": token, // remove later
		"message":     "reset link created",
	})
}

func ResetPassword(c *gin.Context) {
	var body struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	rows, err := db.DB.Query("SELECT email, token_hash, expires_at FROM password_resets")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	defer rows.Close()

	var email string
	var tokenHash string
	var expiresAt int64
	found := false

	for rows.Next() {
		rows.Scan(&email, &tokenHash, &expiresAt)

		if bcrypt.CompareHashAndPassword([]byte(tokenHash), []byte(body.Token)) == nil {
			found = true
			break
		}
	}

	if !found || expiresAt < time.Now().Unix() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired token"})
		return
	}

	newHash, _ := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)

	_, err = db.DB.Exec("UPDATE users SET password = ? WHERE email = ?", string(newHash), email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update password"})
		return
	}

	// delete used token
	db.DB.Exec("DELETE FROM password_resets WHERE email = ?", email)

	c.JSON(http.StatusOK, gin.H{"message": "password reset successful"})
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
		return JwtKey, nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		c.Abort()
		return
	}

	c.Set("username", claims.Username)

	var email string
	err = db.DB.QueryRow("SELECT email FROM users WHERE username = ?", claims.Username).Scan(&email)
	if err == nil {
		c.Set("email", email)
	}

	c.Next()
}

func Me(c *gin.Context) {
	username := c.GetString("username")
	email := c.GetString("email")

	c.JSON(http.StatusOK, gin.H{"username": username, "email": email})
}
