// handlers/todo.go
package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"todo-web/models"

	"github.com/gin-gonic/gin"
)

func GetTodos(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")

		rows, err := db.Query("SELECT id, user_id, title, done FROM todos WHERE user_id = $1", userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		todos := []models.Todo{}
		for rows.Next() {
			var t models.Todo
			rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Done)
			todos = append(todos, t)
		}

		c.JSON(http.StatusOK, todos)
	}
}

func CreateTodo(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var id uint
		userID := c.GetUint("user_id")

		var input struct {
			Title string `json:"title" binding:"required"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := db.QueryRow("INSERT INTO todos (user_id, title, done) VALUES ($1, $2, $3) RETURNING id", userID, input.Title, false).Scan(&id)
		if err != nil {
			fmt.Println("CREATE TODO ERROR:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		todo := models.Todo{
			ID:     uint(id),
			UserID: userID,
			Title:  input.Title,
			Done:   false,
		}

		c.JSON(http.StatusCreated, todo)
	}
}

func UpdateTodo(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		result, err := db.Exec("UPDATE todos SET done = TRUE WHERE id = $1 AND user_id = $2", id, userID)
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
	}
}

func DeleteTodo(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		result, err := db.Exec("DELETE FROM todos WHERE id = $1 AND user_id = $2", id, userID)
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
	}
}
