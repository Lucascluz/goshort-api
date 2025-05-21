package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"math/rand"
	"golang.org/x/crypto/bcrypt"


	"github.com/gin-gonic/gin"
	"gopkg.in/gomail.v2"

	"github.com/google/uuid"
	"goshort-api/configs"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateAlphanumericOTP(length int) string {
	rand.Seed(time.Now().UnixNano())
	otp := make([]byte, length)
	for i := range otp {
		otp[i] = charset[rand.Intn(len(charset))]
	}
	return string(otp)
}

type PasswordResetHandler struct {
	DB *sql.DB
}

func NewPasswordResetHandler(db *sql.DB) *PasswordResetHandler {
	return &PasswordResetHandler{DB: db}
}

func (h *PasswordResetHandler) RequestReset(c *gin.Context) {
	var request struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email"})
		return
	}

	var userID int
	err := h.DB.QueryRow("SELECT id FROM users WHERE email = ?", request.Email).Scan(&userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Load config
	cfg, err := configs.LoadConfig()
	if err != nil {
		log.Println("Failed to load config:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	// Gererate 6 digit OTP
	otp := generateAlphanumericOTP(6)

	// Compose email
	m := gomail.NewMessage()
	m.SetHeader("From", cfg.SMTP_ACCOUNT)
	m.SetHeader("To", request.Email)
	m.SetHeader("Subject", "Password Recovery")
	m.SetBody("text/html", fmt.Sprintf(`
		<p>Hi,</p>
		<p>We received a request to reset your password.</p>
		<p>Your OTP is: <strong>%s</strong></p>
		<p>If you didn't request this, please ignore this email.</p>
		<p>Thanks,</p>
		<p>GoShort Team</p>
	`, otp))

	// Send email
	d := gomail.NewDialer(cfg.SMTP_HOST, cfg.SMTP_PORT, cfg.SMTP_ACCOUNT, cfg.SMTP_PASSWORD)
	if err := d.DialAndSend(m); err != nil {
		log.Println("Failed to send email:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email"})
		return
	}

	// Store OTP in database
	_, err = h.DB.Exec("INSERT INTO otps (user_id, otp, expires_at) VALUES (?, ?, ?)", userID, otp, time.Now().Add(15*time.Minute))
	if err != nil {
		log.Println("Failed to store OTP:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store OTP"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Recovery email sent"})
}

// ConfirmReset verifies the OTP and checks if it's valid
func (h *PasswordResetHandler) ConfirmReset(c *gin.Context) {
	var request struct {
		Email string `json:"email" binding:"required,email"`
		OTP   string `json:"otp" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var userID int
	err := h.DB.QueryRow("SELECT id FROM users WHERE email = ?", request.Email).Scan(&userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var storedOTP string
	var expiresAt time.Time
	err = h.DB.QueryRow("SELECT otp, expires_at FROM otps WHERE user_id = ? AND used = FALSE", userID).Scan(&storedOTP, &expiresAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No valid OTP found"})
		return
	}

	if request.OTP != storedOTP || time.Now().After(expiresAt) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired OTP"})
		return
	}

	resetToken := uuid.NewString()

	// Save token to DB with expiration (e.g., 15 mins)
	_, err = h.DB.Exec(`
		INSERT INTO password_resets (user_id, token, expires_at)
		VALUES (?, ?, ?)
	`, userID, resetToken, time.Now().Add(15*time.Minute))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store reset token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "OTP verified successfully",
		"token": resetToken, // frontend must keep this
	})
}

func (h *PasswordResetHandler) SubmitNewPassword(c *gin.Context) {
	var request struct {
		Email       string `json:"email" binding:"required,email"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
		ResetToken  string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Validate token
	var userID int
	var expiresAt time.Time
	err := h.DB.QueryRow(`
		SELECT user_id, expires_at FROM password_resets
		WHERE token = ? AND used = FALSE
	`, request.ResetToken).Scan(&userID, &expiresAt)

	if err != nil || time.Now().After(expiresAt) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired reset token"})
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password"})
		return
	}

	// Update password
	_, err = h.DB.Exec("UPDATE users SET password_hash = ? WHERE id = ?", hashedPassword, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	// Mark token as used
	_, _ = h.DB.Exec("UPDATE password_resets SET used = TRUE WHERE token = ?", request.ResetToken)

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}



