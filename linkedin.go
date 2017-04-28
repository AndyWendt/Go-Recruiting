package main

import (
	"fmt"
	"io"
	"os"
	"encoding/csv"
)

func main() {
	file, err := os.Open("data/Connections.csv")

	if err != nil {
		printError(err)
		return
	}

	defer file.Close()

	reader := csv.NewReader(file)

	lineCount := 0
	for {
		record, err := reader.Read()

		if err == io.EOF {
			break
		} else if err != nil {
			printError(err)
		}

		fmt.Println("Record", lineCount, "is", record, "and has", len(record), "fields")

		fmt.Println()
		lineCount += 1
	}
}

func printError(err error) (n int, error error) {
	return fmt.Println("Error: ", err)
}
