package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// TodoStatus represents the valid states of a Todo item
type TodoStatus string

const (
	StatusPending   TodoStatus = "pending"
	StatusCompleted TodoStatus = "completed"
)

// Todo represents a single todo item in the application
type Todo struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      TodoStatus `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// decodeJSON is a helper function that decodes JSON request body into a target struct
// using generics for type-safe JSON decoding
func decodeJSON[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("failed to decode request body: %w", err)
	}
	defer r.Body.Close()
	return v, nil
}

// respondJSON is a helper function that writes JSON response with proper headers
func respondJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func main() {
	todos := []Todo{}
	fmt.Println("Hello, World!")

	// Register routes before starting the server
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	//POST /todos
	http.HandleFunc("POST /todos", func(w http.ResponseWriter, r *http.Request) {
		// Use helper function to decode request body
		todo, err := decodeJSON[Todo](r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// create new todo with ID, CreatedAt, UpdatedAt
		todo.ID = uuid.New().String()
		todo.CreatedAt = time.Now()
		todo.UpdatedAt = time.Now()
		todo.Status = StatusPending

		//Write todo to local todos array
		todos = append(todos, todo)

		// Use helper function to respond with JSON
		if err := respondJSON(w, http.StatusCreated, todo); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	//GET /todos
	http.HandleFunc("GET /todos", func(w http.ResponseWriter, r *http.Request) {
		//get all todos
		if err := respondJSON(w, http.StatusOK, todos); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	//PATCH /todos/:id status
	http.HandleFunc("PATCH /todos/", func(w http.ResponseWriter, r *http.Request) {
		//get id from path
		id := r.URL.Path[len("/todos/"):]

		// Use helper function to decode status update
		update, err := decodeJSON[struct {
			Status TodoStatus `json:"status"`
		}](r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		now := time.Now()
		found := false
		for i, todo := range todos {
			if todo.ID == id {
				todo.Status = update.Status
				todo.UpdatedAt = now
				if update.Status == StatusCompleted {
					todo.CompletedAt = &now
				}
				todos[i] = todo
				found = true

				// Respond with updated todo
				if err := respondJSON(w, http.StatusOK, todo); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				break
			}
		}

		if !found {
			http.Error(w, "Todo not found", http.StatusNotFound)
			return
		}
	})

	// Start the server with error handling
	fmt.Println("Listening on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
