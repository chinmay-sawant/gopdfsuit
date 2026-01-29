package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Record struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
	Desc  string `json:"desc"`
}

func main() {
	count := 2000 // 2000 rows for a solid benchmark
	var data []Record

	for i := 1; i <= count; i++ {
		data = append(data, Record{
			ID:    i,
			Name:  fmt.Sprintf("User %d", i),
			Email: fmt.Sprintf("user%d@example.com", i),
			Role:  "Administrator",
			Desc:  "Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
		})
	}

	file, err := os.Create("benchmarks/data.json")
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("Error closing file: %v\n", err)
		}
	}()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		fmt.Printf("Error encoding data: %v\n", err)
		return
	}
	fmt.Printf("Generated benchmarks/data.json with %d records\n", count)
}
