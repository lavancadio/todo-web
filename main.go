package main

import (
	"database/sql"
	"os"

	"todo-web/handlers"
	"todo-web/middleware"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func initDB() *sql.DB {
	db, err := sql.Open("postgres", "postgres://todo:secret@localhost:5432/tododb?sslmode=disable")
	if err != nil {
		panic(err)
	}

	db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id       SERIAL PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL
		)
	`)

	db.Exec(`
		CREATE TABLE IF NOT EXISTS todos (
			id      SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			title   TEXT NOT NULL,
			done    BOOLEAN NOT NULL DEFAULT FALSE,
			FOREIGN KEY (user_id) REFERENCES users(id)
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
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
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
			c.String(500, "index.html not found")
			return
		}
		c.Data(200, "text/html; charset=utf-8", html)
	})

	// Auth routes (no JWT needed)
	r.POST("/register", handlers.Register(db))
	r.POST("/login", handlers.Login(db))

	// Todo routes (JWT protected)
	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/todos", handlers.GetTodos(db))
		protected.POST("/todos", handlers.CreateTodo(db))
		protected.PUT("/todos/:id", handlers.UpdateTodo(db))
		protected.DELETE("/todos/:id", handlers.DeleteTodo(db))
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
