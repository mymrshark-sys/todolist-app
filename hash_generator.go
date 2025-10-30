package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func Savemain() {
	// Ganti password di sini
	password := "admin123"
	
	// Generate hash
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	
	fmt.Println("Password:", password)
	fmt.Println("Hash:", string(hash))
	fmt.Println("\nSQL Query:")
	fmt.Printf("INSERT INTO users (username, password, email, full_name) VALUES ('admin', '%s', 'admin@todolist.com', 'Administrator');\n", string(hash))
}