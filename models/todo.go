package models

// models/todo.go
type Todo struct {
	ID     uint   `json:"id"`
	UserID uint   `json:"user_id"`
	Title  string `json:"title"`
	Done   bool   `json:"done"`
}
