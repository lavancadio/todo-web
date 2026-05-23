package main

import (
	"database/sql"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
)

type Todo struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
	Done bool   `json:"done"`
}

func initDB() *sql.DB {
	db, err := sql.Open("sqlite", "todos.db")
	if err != nil {
		panic(err)
	}

	db.Exec(`
		CREATE TABLE IF NOT EXISTS todos (
			id   INTEGER PRIMARY KEY AUTOINCREMENT,
			text TEXT NOT NULL,
			done BOOLEAN NOT NULL DEFAULT 0
		)
	`)

	return db
}

func main() {
	db := initDB()
	defer db.Close()

	r := gin.Default()

	// CORS
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// serve index.html
	r.GET("/", func(c *gin.Context) {
		html, err := os.ReadFile("index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "index.html not found")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", html)
	})

	// GET /todos
	r.GET("/todos", func(c *gin.Context) {
		rows, err := db.Query("SELECT id, text, done FROM todos")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		todos := []Todo{}
		for rows.Next() {
			var t Todo
			rows.Scan(&t.ID, &t.Text, &t.Done)
			todos = append(todos, t)
		}

		c.JSON(http.StatusOK, todos)
	})

	// POST /todos
	r.POST("/todos", func(c *gin.Context) {
		var t Todo
		if err := c.ShouldBindJSON(&t); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
			return
		}

		result, err := db.Exec("INSERT INTO todos (text, done) VALUES (?, ?)", t.Text, false)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		id, _ := result.LastInsertId()
		t.ID = int(id)
		t.Done = false

		c.JSON(http.StatusCreated, t)
	})

	// PUT /todos/:id
	r.PUT("/todos/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		result, err := db.Exec("UPDATE todos SET done = 1 WHERE id = ?", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		affected, _ := result.RowsAffected()
		if affected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "todo not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "marked as done"})
	})

	// DELETE /todos/:id
	r.DELETE("/todos/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		result, err := db.Exec("DELETE FROM todos WHERE id = ?", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		affected, _ := result.RowsAffected()
		if affected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "todo not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	r.Run(":8080")
}
