package main

import (
	"fmt"
)

func main() {

	// Read the list of websites from the file.
	companies := ReadFromWebsitesCSV("websites.csv")

	// Retrieve the list of the names of the companies to add to the database
	newCompaniesNamesList := retrieveCompaniesNames()

	// Remove companies already present in the database from the list of the companies to be looked for
	newCompaniesNamesList = isCompanyPresentInDatabase(newCompaniesNamesList, companies)

	// Find the websites of the companies.
	newCompaniesList := GetCompaniesWebsites(newCompaniesNamesList)
	fmt.Println("\n Here is the list of the companies that will be added to the database :")
	fmt.Println(newCompaniesList)

	WriteToWebsitesCSV("websites.csv", newCompaniesList)
	fmt.Println("\n The database has been saved in the websites.csv file.")

	companies = ReadFromWebsitesCSV("websites.csv")
	// Enrich the companies with their welcome to the jungle url
	companies = enrichCompaniesListWTTJURL(companies)

	// Enrich the companies with their jobs page url
	companies = enrichCompaniesListJobURL(companies)

	// Write to the websites-full.csv file the list of the companies fully enriched
	WriteToWebsitesCSV("websites-full.csv", companies)
}
