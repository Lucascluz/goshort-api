package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
)

var DB *sql.DB

func setupDatabase() {
	var err error
	DB, err = sql.Open("sqlite", "./app.db")
	if err != nil {
		log.Fatal(err)
	}

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS url_keys (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		short_key TEXT NOT NULL UNIQUE,
		original_url TEXT NOT NULL,
	);`

	_, err = DB.Exec(sqlStmt)
	if err != nil {
		log.Fatalf("Erro ao criar tabela de urls: %q: %s\n", err, sqlStmt)
	}
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const keyLength = 6

func generateShortKey() string {
	rand.Seed(time.Now().UnixNano())
	shortKey := make([]byte, keyLength)
	for i := range shortKey {
		shortKey[i] = charset[rand.Intn(len(charset))]
	}
	return string(shortKey)
}

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.POST("/shorten", func(c *gin.Context) {
		var request struct {
			OriginalURL string `json:"original_url" binding:"required"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "URL inválida"})
			return
		}

		shortKey := generateShortKey()

		_, err := DB.Exec("INSERT INTO url_keys (short_key, original_url) VALUES (?, ?)", shortKey, request.OriginalURL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar no banco"})
			return
		}

		shortURL := fmt.Sprintf("http://localhost:8080/%s", shortKey)
		c.JSON(http.StatusOK, gin.H{"short_url": shortURL})
	})

	r.GET("/:shortKey", func(c *gin.Context) {
		shortKey := c.Param("shortKey")

		var originalURL string
		err := DB.QueryRow("SELECT original_url FROM url_keys WHERE short_key = ?", shortKey).Scan(&originalURL)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "URL não encontrada"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar URL"})
			return
		}

		c.Redirect(http.StatusFound, originalURL)
	})

	r.GET("/list", func(c *gin.Context) {
		rows, err := DB.Query("SELECT short_key, original_url FROM url_keys")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar URLs"})
			return
		}
		defer rows.Close()

		var urls []gin.H
		for rows.Next() {
			var shortKey, originalURL string
			if err := rows.Scan(&shortKey, &originalURL); err == nil {
				urls = append(urls, gin.H{
					"short_url":    fmt.Sprintf("http://localhost:8080/%s", shortKey),
					"original_url": originalURL,
				})
			}
		}

		c.JSON(http.StatusOK, urls)
	})

	return r
}

func main() {
	setupDatabase()
	defer DB.Close()

	r := setupRouter()
	r.Run(":8080")
}
