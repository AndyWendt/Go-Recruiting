package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"encoding/csv"
	"strings"
	"log"
	"github.com/joho/godotenv"
	"net/http"
	"time"
	"encoding/json"
	"net/url"
	//"net/http/httputil"
)

type WorkableCandidatesResponse struct {
	Candidates []struct {
		ID        string      `json:"id"`
		Name      string      `json:"name"`
		Firstname string      `json:"firstname"`
		Lastname  string      `json:"lastname"`
		Headline  interface{} `json:"headline"`
		Account   struct {
			Subdomain string `json:"subdomain"`
			Name      string `json:"name"`
		} `json:"account"`
		Job struct {
			Shortcode string `json:"shortcode"`
			Title     string `json:"title"`
		} `json:"job"`
		Stage                  string      `json:"stage"`
		Disqualified           bool        `json:"disqualified"`
		DisqualificationReason interface{} `json:"disqualification_reason"`
		HiredAt                interface{} `json:"hired_at"`
		Sourced                bool        `json:"sourced"`
		ProfileURL             string      `json:"profile_url"`
		Address                interface{} `json:"address"`
		Phone                  interface{} `json:"phone"`
		Email                  string      `json:"email"`
		Domain                 interface{} `json:"domain"`
		CreatedAt              time.Time   `json:"created_at"`
		UpdatedAt              time.Time   `json:"updated_at"`
	} `json:"candidates"`
	Paging struct {
		Next string `json:"next"`
	} `json:"paging"`
}

func main() {
	loadEnv()

	devsFromLinkedInMaps := loadLinkedInDevs()
	candidatesInWorkable := getWorkableCandidates(devsFromLinkedInMaps)
	devsFromLinkedInSlice := convertDevsFromLinkedInToSlice(devsFromLinkedInMaps)
	devsNotInWorkable := findDevsInLinkedInButNotWorkable(devsFromLinkedInSlice, candidatesInWorkable)

	writeToFile("workable-candidates.csv", candidatesInWorkable)
	writeToFile("devsInLinkedInButNotInWorkable.csv", devsNotInWorkable)
	writeToFile("devsFromLinkedIn.csv", devsFromLinkedInSlice)

	fmt.Println(len(devsFromLinkedInMaps))
	fmt.Println(len(devsNotInWorkable))
	fmt.Println(len(candidatesInWorkable))
}

func convertDevsFromLinkedInToSlice(devsFromLinkedInMaps []map[string]string) [][]string {
	data := make([][]string, 0)
	data = append(data, []string{"First Name", "Last Name", "Email", "Company", "Position"})

	for _, dev := range devsFromLinkedInMaps {
		data = append(data, []string{dev["first_name"], dev["last_name"], dev["email"], dev["company"], dev["position"]})
	}

	return data
}

func getWorkableCandidates(devsFromLinkedIn []map[string]string) [][]string {
	candidatesInWorkable := make([][]string, 0)
	candidatesInWorkable = append(candidatesInWorkable, []string{"First Name", "Last Name", "Email", "Company", "Position"})
	fmt.Println(len(devsFromLinkedIn))
	candidatesInWorkable = getCandidates(os.Getenv("WORKABLE_URL")+"/candidates", devsFromLinkedIn, candidatesInWorkable)
	fmt.Println(len(candidatesInWorkable))
	return candidatesInWorkable
}

func loadLinkedInDevs() []map[string]string {
	file := loadLinkedInConnections()
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
	for {
		record, err := reader.Read()

		if err == io.EOF {
			break
		}

		checkError("error reading file", err)

		if false == hasPosition(record[positionIndex]) {
			continue
		}

		person := map[string]string{
			"first_name": record[firstNameIndex],
			"last_name":  record[lastNameIndex],
			"email":      record[emailIndex],
			"company":    record[companyIndex],
			"position":   record[positionIndex],
		}

		devs = append(devs, person)
	}
	return devs
}

func loadLinkedInConnections() *os.File {
	file, err := os.Open("data/Connections.csv")
	checkError("Failed to open connections file", err)

	return file
}

func loadEnv() {
	err := godotenv.Load()
	checkError("ENV failed to load", err)
}

func printError(err error) (n int, error error) {
	return fmt.Println("Error: ", err)
}

func hasPosition(testPosition string) bool {
	positions := [7]string{"dev", "developer", "engineer", "programmer", "code", "consultant", "freelance"}

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
		os.Exit(1)
	}
}

func getCandidates(url string, devs []map[string]string, workableCandidates [][]string) [][]string {
	fmt.Println("Calling: " + url)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer " + os.Getenv("WORKABLE_API_KEY"))
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

	//dump, err := httputil.DumpResponse(resp, true)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//fmt.Printf("%q", dump)

	checkError("error making request", err)

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	checkError("couldn't read response", err)

	var candidates WorkableCandidatesResponse
	err = json.Unmarshal(body, &candidates)
	checkError("couldn't unmarshall response", err)

	for _, candidate := range candidates.Candidates {
		for _, dev := range devs {
			if strings.ToLower(dev["first_name"]) == strings.ToLower(candidate.Firstname) && strings.ToLower(dev["last_name"]) == strings.ToLower(candidate.Lastname) {
				workableCandidates = append(workableCandidates, []string{dev["first_name"], dev["last_name"], dev["email"], dev["company"], dev["position"]})
			}
		}
	}

	if isValidUrl(candidates.Paging.Next) {
		fmt.Println("Calling getCandidates again")
		return getCandidates(candidates.Paging.Next, devs, workableCandidates)
	}

	return workableCandidates
}

func writeToFile(fileName string, data [][]string) bool {
	file, err := os.Create("data/" +  fileName)
	checkError("Cannot create file", err)
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, value := range data {
		err := writer.Write(value)
		checkError("Cannot write to file", err)
	}

	return true
}

func isValidUrl(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	} else {
		return true
	}
}

func findDevsInLinkedInButNotWorkable(linkedInDevs [][]string, workableCandidates [][]string) [][]string {
	out := make([][]string, 0)
	for _, dev := range linkedInDevs {
		found := false

		for _, candidate := range workableCandidates {
			firstNameMatch := strings.ToLower(dev[0]) == strings.ToLower(candidate[0])
			lastNameMatch := strings.ToLower(dev[1]) == strings.ToLower(candidate[1])
			if firstNameMatch && lastNameMatch {
				found = true
			}
		}

		if found == false {
			out = append(out, dev)
		}
	}

	return out
}
