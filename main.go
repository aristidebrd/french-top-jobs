package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// change this for your situation, 20 or 30, 1,000 or 10,000 may be too high
const MAX_CONCURRENT_JOBS = 20

var dbpoolapi *pgxpool.Pool

func main() {

	var err error

	// Initiate db connection
	dbpoolapi, err = initDbConnection()
	if err != nil {
		log.Fatalf("Connection initialisation failed because of : %s", err)
	}
	defer dbpoolapi.Close()

	r := gin.Default()

	r.GET("/companies", getAllCompaniesAPI)
	r.GET("/company", getCompanyAPI)

	r.Run()

}

func enrichmentEngine() {
	// Initiate db connection
	dbpool, err := initDbConnection()
	if err != nil {
		log.Fatalf("Connection initialisation failed because of : %s", err)
	}
	defer dbpool.Close()

	// Create chrome browser initial context
	ctx, cancel := createBrowser()
	defer cancel()

	// Retrieve the list of the names of the companies to add to the database
	addTop500Companies(dbpool)

	companiesList := getAllCompanies(dbpool)

	var wg sync.WaitGroup
	wg.Add(len(companiesList))

	waitChan := make(chan struct{}, MAX_CONCURRENT_JOBS)

	// Enrich and add the companies that are not present in the database
	for _, company := range companiesList {

		waitChan <- struct{}{}
		go func(company Company) {

			fmt.Println("")
			log.Println("Currently working on :", company.Name)

			// Multi threading the 2 next enrich functions
			var wg2 sync.WaitGroup
			wg2.Add(2)

			go func() {
				defer wg2.Done()
				err = enrichWebsiteAndLinkedinURL(dbpool, company.Name)
				if err != nil {
					log.Printf("An error happened while enriching the company website url : %v", err)
				}
			}()

			go func() {
				defer wg2.Done()
				err = enrichCompanyWTTJUrl(ctx, dbpool, company.Name)
				if err != nil {
					log.Printf("An error happened while enriching the company WTTJ url : %v", err)
				}
			}()

			// Waiting for the 2 functions to end before enriching the company job page url
			wg2.Wait()

			err = enrichCompanyJobUrl(ctx, dbpool, company.Name)
			if err != nil {
				log.Printf("An error happened while enriching the company job's page url : %v", err)
			}

			wg.Done()
			<-waitChan
		}(company)
	}
	wg.Wait()

	// Update companies list
	companiesListUpdated := getAllCompanies(dbpool)

	// Creating waitgroup for offers discovery concurrence search
	var wgOffers sync.WaitGroup
	wgOffers.Add(len(companiesListUpdated))
	waitChanOffers := make(chan struct{}, 2*MAX_CONCURRENT_JOBS)

	// Adding jobs urls to the offers table in the database by looping through all the companies
	for _, company := range companiesListUpdated {

		waitChanOffers <- struct{}{}
		go func(company Company) {

			log.Printf("Now working on %s jobs", company.Name)
			if err != nil {
				log.Printf("An error happened with the query : %s", err)
			}
			addJobs(ctx, dbpool, company)

			wgOffers.Done()
			<-waitChanOffers
		}(company)
	}
	wgOffers.Wait()

}
