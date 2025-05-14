package handlers

import (
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
	"goshort-api/internal/models"
)

const (
	charset   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	keyLength = 6
)

type URLHandler struct {
	DB *sql.DB
}

func NewURLHandler(db *sql.DB) *URLHandler {
	return &URLHandler{DB: db}
}

func generateShortKey() string {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	shortKey := make([]byte, keyLength)
	for i := range shortKey {
		shortKey[i] = charset[rand.Intn(len(charset))]
	}
	return string(shortKey)
}

// ShortenURL handles the URL shortening request
func (h *URLHandler) ShortenURL(c *gin.Context) {
	userID := c.MustGet("user_id") // From middleware

	var request struct {
		OriginalURL string `json:"original_url" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL inválida"})
		return
	}

	shortKey := generateShortKey()
	
	_, err := h.DB.Exec(
		"INSERT INTO url_keys (short_key, original_url, user_id) VALUES (?, ?, ?)", 
		shortKey, request.OriginalURL, userID,
	)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar no banco"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"short_url": fmt.Sprintf("http://localhost:8080/%s", shortKey),
	})
}

// RedirectURL handles the redirection from short URL to original URL
func (h *URLHandler) RedirectURL(c *gin.Context) {
	shortKey := c.Param("shortKey")
	var originalURL string
	
	err := h.DB.QueryRow(
		"SELECT original_url FROM url_keys WHERE short_key = ?", 
		shortKey,
	).Scan(&originalURL)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "URL não encontrada"})
		return
	}
	
	c.Redirect(http.StatusFound, originalURL)
}

// ListURLs handles the request to list all shortened URLs
func (h *URLHandler) ListURLs(c *gin.Context) {
	rows, err := h.DB.Query("SELECT short_key, original_url FROM url_keys")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar URLs"})
		return
	}
	defer rows.Close()

	var urls []models.URL
	for rows.Next() {
		var u models.URL
		if err := rows.Scan(&u.ShortKey, &u.OriginalURL); err == nil {
			u.ShortKey = fmt.Sprintf("http://localhost:8080/%s", u.ShortKey)
			urls = append(urls, u)
		}
	}

	c.JSON(http.StatusOK, urls)
}