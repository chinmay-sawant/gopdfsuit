package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
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
	data := make([]Record, 0, count)

	var idBuf []byte
	for i := 1; i <= count; i++ {
		// PERF-15: AppendInt into reusable buffer instead of Itoa in loop
		idBuf = strconv.AppendInt(idBuf[:0], int64(i), 10)
		idStr := string(idBuf)
		data = append(data, Record{
			ID:    i,
			Name:  "User " + idStr,
			Email: "user" + idStr + "@example.com",
			Role:  "Administrator",
			Desc:  "Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
		})
	}

	file, err := os.Create("data.json")
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
