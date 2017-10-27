package main

import (
	"fmt"
	"io"
	"os"
	"encoding/csv"
	"strings"
)

func main() {
	file, err := os.Open("data/Connections.csv")

	if err != nil {
		printError(err)
		return
	}

	defer file.Close()

	reader := csv.NewReader(file)

	firstNameIndex := 0
	lastNameIndex := 1
	emailIndex := 2
	//companyIndex := 3
	positionIndex := 4
	//connectedOnIndex := 5
	//tagsIndex := 6
	out := make([][]string, 0)


	lineCount := 0
	for {
		record, err := reader.Read()

		if err == io.EOF {
			break
		} else if err != nil {
			printError(err)
		}


		fmt.Println(record[firstNameIndex], record[lastNameIndex], record[emailIndex])


		if hasPosition(record[positionIndex]) {
			tt := make([]string, 0)
			out = append(out, append(tt, record[firstNameIndex]))
		}

		fmt.Println("Record", lineCount, "is", record, "and has", len(record), "fields")

		fmt.Println()
		lineCount += 1
	}

	fmt.Println(len(out))
}

func printError(err error) (n int, error error) {
	return fmt.Println("Error: ", err)
}

func hasPosition(testPosition string) (hasPosition bool) {
	positions := [5]string{"dev", "developer", "engineer", "programmer", "code"}

	for _, position := range positions {
		containsPosition := strings.Contains(strings.ToLower(testPosition), position)

		if containsPosition == true {
			return true
		}
	}

	return false
}
