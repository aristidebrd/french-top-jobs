package main

import (
	"bytes"
	"crypto/tls"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"gopkg.in/yaml.v2"
)

// This variable stores the name of a company and the website associated
type Company struct {
	Name        string
	Website     string
	LinkedInURL string
	WTTJURL     string
	JobsPageURL string
}

// This variable helps to create custom json body to make the call on the crunchbase organization search API
type SearchRequest struct {
	FieldIDs []string `json:"field_ids"`
	Query    []struct {
		Type       string   `json:"type"`
		FieldID    string   `json:"field_id"`
		OperatorID string   `json:"operator_id"`
		Values     []string `json:"values"`
	} `json:"query"`
	Limit int `json:"limit"`
}

type SearchResponse struct {
	Count    int `json:"count"`
	Entities []struct {
		UUID       string `json:"uuid"`
		Properties struct {
			Identifier struct {
				Permalink   string `json:"permalink"`
				ImageID     string `json:"image_id"`
				UUID        string `json:"uuid"`
				EntityDefID string `json:"entity_def_id"`
				Value       string `json:"value"`
			} `json:"identifier"`
			WebsiteURL  string `json:"website_url"`
			LinkedInURL struct {
				Value string `json:"linkedin"`
			}
		} `json:"properties"`
	} `json:"entities"`
}

func ReadFromWebsitesCSV(filename string) []Company {
	var companiesFromFile []Company

	// Checks if file exists, if it does not then creates it
	_, err := os.Stat(filename)
	if err != nil {
		fmt.Println("File does not exist, creating it.")
		_, err := os.Create(filename)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		return nil
	}

	// Open the file for reading.
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer file.Close()

	// Create a new CSV reader.
	reader := csv.NewReader(file)

	// Read all of the records from the CSV file.
	for {
		var company Company

		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err)
			return nil
		}

		company.Name = record[0]
		company.Website = record[1]
		company.LinkedInURL = record[2]
		company.WTTJURL = record[3]
		company.JobsPageURL = record[4]
		companiesFromFile = append(companiesFromFile, company)
	}

	// If the company name is not found in the CSV file, return an empty string.
	return companiesFromFile
}

func WriteToWebsitesCSV(filename string, companies []Company) {
	// Open the file for writing.
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	// Create a new CSV writer.
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write each company name and website URL to the CSV file.
	for _, company := range companies {

		err = writer.Write([]string{company.Name, company.Website, company.LinkedInURL, company.WTTJURL, company.JobsPageURL})
		if err != nil {
			fmt.Println(err)
		}
	}
}

// This function retrieves the list of the 500 top french (in terms of growth and interest) starts and scales up, returns them as a slice of strings.
func retrieveCompaniesNames() []string {
	// Create the companies names slice
	var companiesNames []string

	// Create a new colly collector
	c := colly.NewCollector()

	// Define a callback to execute when the collector finds a table with id `tablepress-9`.
	c.OnHTML("table#tablepress-9 > tbody", func(h *colly.HTMLElement) {
		h.ForEach("tr", func(_ int, el *colly.HTMLElement) {
			// Append to a slice each lign company name.
			companiesNames = append(companiesNames, el.ChildText("td:nth-child(2)"))
		})
	})

	// Visiter la page web.
	c.Visit("https://datarecrutement.fr/actualites/nos-actualites/tech500/")
	return companiesNames
}

func GetCompaniesWebsites(companyNames []string) []Company {

	// Open the YAML file containing the API key.
	file, err := os.Open("crunchbase-api-key.yaml")
	if err != nil {
		log.Fatal(err)
	}

	// Read the YAML file into a map.
	var data map[string]interface{}
	err = yaml.NewDecoder(file).Decode(&data)
	if err != nil {
		log.Fatal(err)
	}

	// Get the API key from the map.
	apiKey := data["api_key"].(string)

	// Get the websites of the companies in the slice
	var companies []Company

	for _, companyName := range companyNames {

		fmt.Println("Currently working on :", companyName)

		// Define company variable
		var company Company

		company.Name = companyName

		// Create a new SearchRequest object
		searchRequest := SearchRequest{
			FieldIDs: []string{"identifier", "website_url", "linkedin"},
			Query: []struct {
				Type       string   `json:"type"`
				FieldID    string   `json:"field_id"`
				OperatorID string   `json:"operator_id"`
				Values     []string `json:"values"`
			}{
				{
					Type:       "predicate",
					FieldID:    "identifier",
					OperatorID: "contains",
					Values:     []string{company.Name},
				},
				{
					Type:       "predicate",
					FieldID:    "location_identifiers",
					OperatorID: "includes",
					Values:     []string{"f134827e-36a1-fd31-a82f-950489e103ef"},
				},
			},
			Limit: 1,
		}

		// Marshal the SearchRequest object into JSON
		jsonBytes, err := json.Marshal(searchRequest)
		if err != nil {
			fmt.Println(err)
			return nil
		}

		// Create a new HTTPS client
		customTransport := http.DefaultTransport.(*http.Transport).Clone()
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		client := &http.Client{Transport: customTransport}

		// Create a new HTTP GET request
		req, err := http.NewRequest("POST", "https://api.crunchbase.com/api/v4/searches/organizations", bytes.NewReader(jsonBytes))

		// Set the API key in the HTTP header
		req.Header.Add("X-cb-user-key", apiKey)
		if err != nil {
			fmt.Println(err)
			return nil
		}

		// Set the HTTP request header `Content-Type` to `application/json`
		req.Header.Set("Content-Type", "application/json")

		// Set the HTTP request header 'accept' to 'application/json'
		req.Header.Set("accept", "application/json")

		// Run the request, if the body is not valid then the API limit has been reached so stops for 50 seconds then retry the request.
		// If the body is valid then append it to the companies list.
		for {
			// Execute the HTTP request
			resp, err := client.Do(req)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			// Close the HTTP response body
			defer resp.Body.Close()

			// Read the HTTP response body
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			if json.Valid([]byte(body)) {
				// Create a new SearchResponse variable to store the unmarshaled response
				var searchResponse SearchResponse

				// Unmarshal the JSON response body into a Company object
				if err := json.Unmarshal(body, &searchResponse); err != nil {
					fmt.Println(err)
					return nil
				}

				if searchResponse.Count != 0 {
					company.Website = strings.TrimSpace(searchResponse.Entities[0].Properties.WebsiteURL)
					company.LinkedInURL = strings.TrimSpace(searchResponse.Entities[0].Properties.LinkedInURL.Value)
					company.WTTJURL = ""
					company.JobsPageURL = ""
					companies = append(companies, company)
				} else {
					company.Website = ""
					company.LinkedInURL = ""
					company.WTTJURL = ""
					company.JobsPageURL = ""
					companies = append(companies, company)
				}
				break
			} else {
				// handle the error here
				fmt.Println("\n The crunchbase API limit might has been reached, stopping for 50 secs")
				time.Sleep(50 * time.Second)
				fmt.Println("\n Resuming the search")
			}
		}
	}

	return companies
}

// Based on a list of companies, returns the list of the companies that aren't already present in the database
func isCompanyPresentInDatabase(newCompaniesNamesList []string, companies []Company) []string {
	var newCompaniesNamesApprovedList []string
	var alreadyExist bool
	for _, newCompanyName := range newCompaniesNamesList {
		alreadyExist = false
		for _, company := range companies {
			if newCompanyName == company.Name {
				alreadyExist = true
				break
			}
		}
		if !alreadyExist {
			newCompaniesNamesApprovedList = append(newCompaniesNamesApprovedList, newCompanyName)
		}
	}
	return newCompaniesNamesApprovedList
}
