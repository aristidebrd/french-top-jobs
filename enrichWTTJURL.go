package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Define a function to get the URL of a company on Welcome to the Jungle.
func getCompanyWTTJURL(companyName string) string {

	// Escape the company name to fit it in the http request query
	query := url.QueryEscape("companyName")

	// Build the search URL.
	searchURL := fmt.Sprintf("https://www.welcometothejungle.com/fr/companies?query=%s", query)

	// Create a new HTTP client.
	client := &http.Client{}

	// Make a GET request to the search URL.
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	// Send the request and get the response.
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	// Close the response body when we're done with it.
	defer resp.Body.Close()

	// Read the response body into a byte slice.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	// Create a new Goquery document from the response body.
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		fmt.Println(err)
		return ""
	}

	// Find the first search result.
	result := doc.Find(".search-result-item").First()

	// Get the link to the company page.
	WTTJURL, _ := result.Find(".company-name a").Attr("href")

	if !validateWTTJURL(WTTJURL, companyName) {
		WTTJURL = ""
	}

	return WTTJURL
}

// Define a function to scrape a company Welcome to the Jungle page and verify that the company name is present in its content.
func validateWTTJURL(companyURL string, companyName string) bool {

	// Create a new HTTPS client.
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	client := &http.Client{Transport: customTransport}

	// Make a GET request to the company page URL.
	req, err := http.NewRequest("GET", companyURL, nil)
	if err != nil {
		fmt.Println(err)
		return false
	}

	// Send the request and get the response.
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return false
	}

	// Close the response body when we're done with it.
	defer resp.Body.Close()

	// Read the response body into a byte slice.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return false
	}

	// Create a new goquery document from the response body.
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	// Select the div with the class "sc-1m9biip-1 jcWSWv cms-content text".
	text := doc.Find(".sc-1m9biip-1.jcWSWv.cms-content.text").Text()

	if strings.Contains(text, companyName) {
		return true
	} else {
		return false
	}
}

// Define a function to enrich a list of companies with their welcome to the jungle url.
func enrichCompaniesListWTTJURL(companiesList []Company) []Company {

	for _, companyToEnrich := range companiesList {
		if companyToEnrich.Website == "" {
			continue
		} else {
			if companyToEnrich.WTTJURL == "" {
				companyToEnrich.WTTJURL = getCompanyWTTJURL(companyToEnrich.Name)
			}
		}
	}
	return companiesList
}
