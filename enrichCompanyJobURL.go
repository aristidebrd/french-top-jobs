package main

import (
	"context"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Check if the job page of a company is present on their own website
func checkCompanyJobPage(ctx context.Context, website string) string {
	jobsPageURL := ""

	// Create the request context
	ctx, cancel := createTab(ctx)
	defer cancel()

	if !isUrlWorking(website) {
		return jobsPageURL
	}

	html := ""
	var err error
	for i := 0; i < 5; i++ {
		err := chromedp.Run(ctx,
			// visit the target page
			chromedp.Navigate(website),
			// wait for the page to load
			chromedp.Sleep(2000*time.Millisecond),
		)
		if err != nil {
			i = i + 1
		} else {
			break
		}
	}

	if err != nil {
		log.Fatal("Error while performing the automation logic:", err)
	}

	c := chromedp.FromContext(ctx)
	rootNode, err := dom.GetDocument().Do(cdp.WithExecutor(ctx, c.Target))
	if err != nil {
		log.Fatal(err)
	}

	html, err = dom.GetOuterHTML().WithNodeID(rootNode.NodeID).Do(cdp.WithExecutor(ctx, c.Target))
	if err != nil {
		log.Fatal(err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		log.Fatal(err)
	}

	// Find all the href links in the HTML document
	links := doc.Find("a")

	// Create a regular expression to match the URL of the jobs page
	regex := regexp.MustCompile(".*(jobs|careers|carreers).*")

	// Iterate over the links and print them to the console
	for i := 0; i < links.Length(); i++ {
		link := links.Eq(i)

		// Get the href attribute of the link
		href, ok := link.Attr("href")
		if ok {
			if regex.MatchString(href) {
				jobsPageURL = removeTrailingSlash(getAbsoluteUrl(website, href))
				break
			}
		}
	}

	return jobsPageURL
}

// Enrich the job url for a company
func enrichJobURL(ctx context.Context, companyToEnrich Company) (Company, error) {

	var err error
	if companyToEnrich.Website != "" {
		companyToEnrich.JobsPageURL = checkCompanyJobPage(ctx, companyToEnrich.Website)
	}
	if companyToEnrich.JobsPageURL != "" {
		log.Printf("%s careers page was found on it's website : %v", companyToEnrich.Name, companyToEnrich)
	} else if companyToEnrich.WTTJURL != "" {
		companyToEnrich.JobsPageURL = companyToEnrich.WTTJURL + "/jobs"
		log.Printf("%s careers page was not found on it's website, using wttj jobs page: %v", companyToEnrich.Name, companyToEnrich)
	} else if companyToEnrich.LinkedInURL != "" {
		companyToEnrich.JobsPageURL = companyToEnrich.LinkedInURL + "/jobs"
		log.Printf("%s careers page was not found on it's website, using linkedin jobs page : %v", companyToEnrich.Name, companyToEnrich)
	} else {
		log.Printf("%s careers page was not found on it's website, neither wttj or linked pages were found, not enriching job page url : %v", companyToEnrich.Name, companyToEnrich)
	}

	// Return the company enriched with it's jobs page URL
	return companyToEnrich, err
}

// Define a function to enrich a list of companies with their welcome to the jungle url.
func enrichCompanyJobUrl(ctx context.Context, db *pgxpool.Pool, companyName string) error {
	var err error

	company, exists, err := getCompany(db, companyName)
	switch {
	case err != nil:
		log.Printf("Database query failed because of : %s", err)
	case !exists:
		log.Printf("The company does not exist")
	default:
		if company.JobsPageURL == "" {
			company, err = enrichJobURL(ctx, company)
			if err != nil {
				return err
			}
			updateCompany(db, company)
			if err != nil {
				log.Printf("An error happened with the query : %s", err)
			}
		} else {
			log.Printf("Company already has it's Job url enriched.")
		}
	}

	return err
}
