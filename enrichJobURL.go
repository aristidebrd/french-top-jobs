package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"regexp"

	"github.com/PuerkitoBio/goquery"
)

// Check if the job page of a company is present on their own website
func checkCompanyJobPage(website string) string {
	jobsPageURL := ""

	// Create a new HTTPS client
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	client := &http.Client{Transport: customTransport}

	// Make a GET request to the company website URL
	req, err := http.NewRequest("GET", website, nil)
	if err != nil {
		fmt.Println(err)
	}

	// Send the request and get the response
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	// Close the response body
	defer resp.Body.Close()

	// Create a new goquery document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	// Find all the href links in the HTML document
	links := doc.Find("a")

	// Create a regular expression to match the URL of the jobs page
	regex := regexp.MustCompile(".*?/(jobs|careers)*$")

	// Iterate over the links and print them to the console
	for i := 0; i < links.Length(); i++ {
		link := links.Eq(i)

		// Get the href attribute of the link
		href, ok := link.Attr("href")
		if ok {
			if regex.MatchString(href) {
				jobsPageURL = href
				break
			}
		}
	}

	return jobsPageURL
}

// Enrich the job url for a company
func enrichJobURL(companyToEnrich Company) Company {

	companyToEnrich.Website = checkCompanyJobPage(companyToEnrich.Website)
	if companyToEnrich.JobsPageURL != "" {
		return companyToEnrich
	} else if companyToEnrich.WTTJURL != "" {
		companyToEnrich.JobsPageURL = companyToEnrich.WTTJURL + "/jobs"
	} else if companyToEnrich.LinkedInURL != "" {
		companyToEnrich.JobsPageURL = companyToEnrich.LinkedInURL + "/jobs"
	}

	// Return the company enriched with it's jobs page URL
	return companyToEnrich
}

// Enrich the job url for a list of companies
func enrichCompaniesListJobURL(companiesList []Company) []Company {

	for index, companyToEnrich := range companiesList {
		if companyToEnrich.Website == "" {
			continue
		} else {
			companiesList[index] = enrichJobURL(companyToEnrich)
		}
	}

	// Return the list of companies enriched with their job's page url
	return companiesList
}
