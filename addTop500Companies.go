package main

import (
	"log"

	"github.com/gocolly/colly"
	"github.com/jackc/pgx/v5/pgxpool"
)

// This function retrieves the list of the 500 top french (in terms of growth and interest) starts and scales up, returns them as a slice of strings.
func addTop500Companies(db *pgxpool.Pool) {
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

	newCompaniesNamesList := checkNewCompaniesList(db, companiesNames)

	for _, companyName := range newCompaniesNamesList {
		var company Company
		company.Name = companyName
		company.IsTop500 = true
		company.Website = ""
		company.LinkedInURL = ""
		company.WTTJURL = ""
		company.JobsPageURL = ""
		err := addCompany(db, company)
		if err != nil {
			log.Printf("An error happened while adding the company to the database : %v", err)
		}
	}

}

// Based on a list of companies, returns the list of the companies that aren't already present in the database
func checkNewCompaniesList(db *pgxpool.Pool, newCompaniesNamesList []string) []string {
	var newCompaniesNamesApprovedList []string
	for _, newCompanyName := range newCompaniesNamesList {
		_, exist, err := getCompany(db, newCompanyName)
		if err != nil {
			log.Printf("An error happened with the query : %s", err)
		}

		if !exist {
			log.Printf("%s does not exist in the database, it will be created", newCompanyName)
			newCompaniesNamesApprovedList = append(newCompaniesNamesApprovedList, newCompanyName)
		}
	}
	return newCompaniesNamesApprovedList
}
