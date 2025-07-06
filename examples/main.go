package main

import (
	"fmt"
	"log"

	"github.com/0x6b/sqlfmt"
)

func main() {
	// Simple example using Format
	fmt.Println("=== Simple Format Example ===")
	sql := "SELECT id, name FROM users WHERE active = 1 ORDER BY created_at DESC"
	formatted, err := sqlfmt.Format(sql, sqlfmt.DefaultFormatOptions)
	if err != nil {
		log.Fatalf("Error formatting SQL: %v", err)
	}
	fmt.Printf("%s\n\n", formatted)

	// Reusable formatter example
	fmt.Println("=== Reusable Formatter Example ===")
	formatter, err := sqlfmt.NewFormatter()
	if err != nil {
		log.Fatalf("Failed to create formatter: %v", err)
	}
	defer func(formatter *sqlfmt.Formatter) {
		err := formatter.Close()
		if err != nil {
			log.Fatalf("Failed to close formatter: %v", err)
		}
	}(formatter)

	// Format multiple queries with the same formatter (more efficient)
	queries := []string{
		"SELECT * FROM products WHERE price > 100",
		"INSERT INTO logs (message, created_at) VALUES ('test', NOW())",
		"UPDATE users SET last_login = CURRENT_TIMESTAMP WHERE id = 123",
		"DELETE FROM sessions WHERE expired_at < NOW() - INTERVAL '7 days'",
	}

	for i, sql := range queries {
		formatted, err := formatter.Format(sql, sqlfmt.DefaultFormatOptions)
		if err != nil {
			log.Printf("Error formatting query %d: %v", i+1, err)
			continue
		}
		fmt.Printf("# Query %d:\n%s\n\n", i+1, formatted)
	}
}
