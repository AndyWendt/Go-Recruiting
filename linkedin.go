package main

import (
	"fmt"
	"io"
	"os"
	"encoding/csv"
	"strings"
	"log"
	"github.com/joho/godotenv"
	"net/http"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	//fmt.Println(os.Getenv("WORKABLE_API_KEY"))
	//s3Bucket := os.Getenv("S3_BUCKET")
	//secretKey := os.Getenv("SECRET_KEY")

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
	companyIndex := 3
	positionIndex := 4
	//connectedOnIndex := 5
	//tagsIndex := 6
	devs := make([]map[string]string, 0)


	//lineCount := 0
	for {
		record, err := reader.Read()

		if err == io.EOF {
			break
		} else if err != nil {
			printError(err)
		}

		if false == hasPosition(record[positionIndex]) {
			continue
		}


		person := map[string]string {
			"first_name": record[firstNameIndex],
			"last_name": record[lastNameIndex],
			"email": record[emailIndex],
			"company": record[companyIndex],
			"position": record[positionIndex],
		}

		devs = append(devs, person)
	}



	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://vynyl.workable.com/spi/v3/candidates", nil)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer " + os.Getenv("WORKABLE_API_KEY"))
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)


	if err != nil {
		printError(err)
	}

	io.Copy(os.Stdout, resp.Body)

	data := make([][]string, 0)

	data = append(data, []string{"First Name", "Last Name", "Email", "Company", "Position"})

	for _, dev := range devs {
		data = append(data, []string{dev["first_name"], dev["last_name"], dev["email"], dev["company"], dev["position"]})
	}

	file, err = os.Create("data/result.csv")
	checkError("Cannot create file", err)
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, value := range data {
		err := writer.Write(value)
		checkError("Cannot write to file", err)
	}



	fmt.Println(len(devs))
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

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}
