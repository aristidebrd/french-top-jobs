package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"gopkg.in/yaml.v2"
)

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
				Value string `json:"value"`
			} `json:"linkedin"`
		} `json:"properties"`
	} `json:"entities"`
}

func enrichWebsiteAndLinkedinURL(db *pgxpool.Pool, companyName string) error {

	company, _, err := getCompany(db, companyName)
	if err != nil {
		log.Printf("An error occured : %v", err)
	}

	if company.Website != "" && company.LinkedInURL != "" {
		return nil
	}

	// Open the YAML file containing the API key.
	file, err := os.Open("secrets/crunchbase-api-key.yaml")
	if err != nil {
		log.Printf("Error while opening the yaml file contaning the API key : %v", err)
	}

	// Read the YAML file into a map.
	var data map[string]interface{}
	err = yaml.NewDecoder(file).Decode(&data)
	if err != nil {
		log.Printf(" Error while decoding the yaml file into a map : %v", err)
	}

	// Get the API key from the map.
	apiKey := data["api_key"].(string)

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
				Values:     []string{companyName},
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
	client := http.Client{}

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
				if company.Website == "" {
					WebsiteURL := removeTrailingSlash(searchResponse.Entities[0].Properties.WebsiteURL)
					updateCompanyWebsiteURL(db, company.Name, WebsiteURL)
					log.Printf("%s has been enriched with it's website url : %s", companyName, WebsiteURL)
				}
				if company.LinkedInURL == "" {
					LinkedInURL := removeTrailingSlash(searchResponse.Entities[0].Properties.LinkedInURL.Value)
					updateCompanyLinkedinURL(db, company.Name, LinkedInURL)
					log.Printf("%s has been enriched with it's wttj url : %s", companyName, LinkedInURL)
				}
			}

			break
		} else {
			// handle the error here
			fmt.Println("\n The crunchbase API limit might has been reached, stopping for 50 secs")
			time.Sleep(50 * time.Second)
			fmt.Println("\n Resuming the search")
		}
	}

	return err
}
