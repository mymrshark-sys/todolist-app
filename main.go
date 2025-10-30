package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type Note struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Status    string    `json:"status"`
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	CreatedAt time.Time `json:"created_at"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

var db *sql.DB
var store = sessions.NewCookieStore([]byte("todolist-secret-key-change-this-in-production"))

func main() {
	// Print working directory untuk debug
	wd, _ := os.Getwd()
	fmt.Println("📁 Working Directory:", wd)

	// Koneksi ke PostgreSQL
	var err error
	connStr := "postgresql://postgres:ZdchCwkDVmKKpkiJKVOSFrTGFkpEDBVW@ballast.proxy.rlwy.net:27126/railway?sslmode=disable"
	
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("❌ Error connecting to database:", err)
	}
	defer db.Close()

	// Test koneksi
	err = db.Ping()
	if err != nil {
		log.Fatal("❌ Error pinging database:", err)
	}
	fmt.Println("✅ PostgreSQL connected successfully!")

	// Setup router
	router := mux.NewRouter()

	// Public Routes (tidak perlu login)
	router.HandleFunc("/login", loginPage).Methods("GET")
	router.HandleFunc("/register", registerPage).Methods("GET")
	router.HandleFunc("/api/login", handleLogin).Methods("POST")
	router.HandleFunc("/api/register", handleRegister).Methods("POST")
	router.HandleFunc("/api/logout", handleLogout).Methods("POST")

	// Protected Routes (perlu login)
	router.HandleFunc("/", authMiddleware(homePage)).Methods("GET")
	router.HandleFunc("/api/notes", authMiddleware(getNotes)).Methods("GET")
	router.HandleFunc("/api/notes", authMiddleware(createNote)).Methods("POST")
	router.HandleFunc("/api/notes/{id}", authMiddleware(updateNote)).Methods("PUT")
	router.HandleFunc("/api/notes/{id}", authMiddleware(deleteNote)).Methods("DELETE")
	router.HandleFunc("/api/notes/{id}/toggle", authMiddleware(toggleStatus)).Methods("PATCH")
	router.HandleFunc("/api/user", authMiddleware(getCurrentUser)).Methods("GET")

	// Static files
	staticDir := "./static"
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		log.Println("⚠️  Warning: static directory not found at", staticDir)
	} else {
		fmt.Println("✅ Static directory found:", staticDir)
	}
	
	fs := http.FileServer(http.Dir(staticDir))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	// Start server
	fmt.Println("🚀 Server running on http://localhost:8080")
	fmt.Println("📝 Default login: admin / admin123")
	fmt.Println("📝 Press Ctrl+C to stop")
	log.Fatal(http.ListenAndServe(":8080", router))
}

// Middleware untuk autentikasi
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        session, _ := store.Get(r, "todolist-session")
        
        // Cek apakah user sudah login
        userID, ok := session.Values["user_id"]
        if !ok || userID == nil {
            // Jika request dari API, return JSON error
            if r.Header.Get("Content-Type") == "application/json" || strings.HasPrefix(r.URL.Path, "/api") {
                w.WriteHeader(http.StatusUnauthorized)
                json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
                return
            }
            // Jika request halaman, redirect ke login
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return
        }
        
        // Set user_id ke context untuk digunakan di handler
        next(w, r)
    }
}

// Get user ID from session
func getUserID(r *http.Request) int {
	session, _ := store.Get(r, "todolist-session")
	userID, _ := session.Values["user_id"].(int)
	return userID
}

// Handler untuk halaman login
func loginPage(w http.ResponseWriter, r *http.Request) {
	// Jika sudah login, redirect ke home
	session, _ := store.Get(r, "todolist-session")
	if userID, ok := session.Values["user_id"]; ok && userID != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	tmpl, err := template.ParseFiles("templates/login.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// Handler untuk halaman register
func registerPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/register.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// Handler untuk halaman utama
func homePage(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "todolist-session")
	username := session.Values["username"]

	templatePaths := []string{
		"templates/index.html",
		"./templates/index.html",
		filepath.Join("templates", "index.html"),
	}
	
	var tmpl *template.Template
	var err error
	
	for _, path := range templatePaths {
		if _, statErr := os.Stat(path); statErr == nil {
			tmpl, err = template.ParseFiles(path)
			if err == nil {
				break
			}
		}
	}
	
	if tmpl == nil || err != nil {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}
	
	data := map[string]interface{}{
		"Username": username,
	}
	
	tmpl.Execute(w, data)
}

// Handle login
func handleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
		return
	}

	// Get user from database
	var user User
	err = db.QueryRow(`
		SELECT id, username, password, email, full_name 
		FROM users 
		WHERE username = $1
	`, req.Username).Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.FullName)
	
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid username or password"})
		return
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid username or password"})
		return
	}

	// Create session
	session, _ := store.Get(r, "todolist-session")
	session.Values["user_id"] = user.ID
	session.Values["username"] = user.Username
	session.Save(r, w)

	fmt.Printf("✅ User %s logged in\n", user.Username)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Login successful",
		"user": map[string]interface{}{
			"id":        user.ID,
			"username":  user.Username,
			"email":     user.Email,
			"full_name": user.FullName,
		},
	})
}

// Handle register
func handleRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
		return
	}

	// Validate input
	if req.Username == "" || req.Password == "" || req.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "All fields are required"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error processing password"})
		return
	}

	// Insert user
	var userID int
	err = db.QueryRow(`
		INSERT INTO users (username, password, email, full_name) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id
	`, req.Username, string(hashedPassword), req.Email, req.FullName).Scan(&userID)
	
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Username or email already exists"})
		return
	}

	fmt.Printf("✅ New user registered: %s\n", req.Username)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Registration successful"})
}

// Handle logout
func handleLogout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "todolist-session")
	session.Values["user_id"] = nil
	session.Values["username"] = nil
	session.Options.MaxAge = -1
	session.Save(r, w)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Logout successful"})
}

// Get current user
func getCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	
	var user User
	err := db.QueryRow(`
		SELECT id, username, email, full_name, created_at 
		FROM users 
		WHERE id = $1
	`, userID).Scan(&user.ID, &user.Username, &user.Email, &user.FullName, &user.CreatedAt)
	
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "User not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// Get semua notes (filter by user)
func getNotes(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	
	rows, err := db.Query(`
		SELECT id, title, content, status, created_at, updated_at 
		FROM notes 
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var note Note
		err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.Status, &note.CreatedAt, &note.UpdatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		notes = append(notes, note)
	}

	if notes == nil {
		notes = []Note{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes)
}

// Create note baru (with user_id)
func createNote(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	
	var note Note
	err := json.NewDecoder(r.Body).Decode(&note)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if note.Title == "" || note.Content == "" {
		http.Error(w, "Title and content are required", http.StatusBadRequest)
		return
	}

	err = db.QueryRow(`
		INSERT INTO notes (title, content, status, user_id) 
		VALUES ($1, $2, 'pending', $3) 
		RETURNING id, created_at, updated_at
	`, note.Title, note.Content, userID).Scan(&note.ID, &note.CreatedAt, &note.UpdatedAt)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	note.Status = "pending"
	note.UserID = userID

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(note)
}

// Update note (verify ownership)
func updateNote(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	userID := getUserID(r)

	var note Note
	err = json.NewDecoder(r.Body).Decode(&note)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if note.Title == "" || note.Content == "" {
		http.Error(w, "Title and content are required", http.StatusBadRequest)
		return
	}

	err = db.QueryRow(`
		UPDATE notes 
		SET title = $1, content = $2, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $3 AND user_id = $4
		RETURNING status, updated_at
	`, note.Title, note.Content, id, userID).Scan(&note.Status, &note.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Note not found or unauthorized", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	note.ID = id

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(note)
}

// Delete note (verify ownership)
func deleteNote(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	userID := getUserID(r)

	result, err := db.Exec("DELETE FROM notes WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Note not found or unauthorized", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Toggle status note (verify ownership)
func toggleStatus(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	userID := getUserID(r)

	var currentStatus string
	err = db.QueryRow("SELECT status FROM notes WHERE id = $1 AND user_id = $2", id, userID).Scan(&currentStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Note not found or unauthorized", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newStatus := "pending"
	if currentStatus == "pending" {
		newStatus = "completed"
	}

	_, err = db.Exec(`
		UPDATE notes 
		SET status = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2 AND user_id = $3
	`, newStatus, id, userID)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": newStatus})
}