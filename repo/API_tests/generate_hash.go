// +build ignore

// Helper to generate bcrypt hash for test seeding
package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	hash, err := bcrypt.GenerateFromPassword([]byte("AdminPass123!"), 12)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(hash))
}
