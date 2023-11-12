package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Define a function to get the URL of a company on Welcome to the Jungle.
func findCompanyWTTJURL(ctx context.Context, companyName string) (string, error) {

	// Escape the company name to fit it in the http request query
	query := url.QueryEscape(companyName)

	// Build the search URL.
	searchURL := fmt.Sprintf("https://www.welcometothejungle.com/fr/companies?query=%s", query)

	// Create the request context
	ctx, cancel := createTab(ctx)
	defer cancel()

	selector := "sc-1wqurwm-0 NyiTc ais-Hits-list"

	var node []cdp.NodeID
	html := ""
	var err error
	for i := 0; i < 5; i++ {
		err := chromedp.Run(ctx,
			// visit the target page
			chromedp.Navigate(searchURL),
			// wait for the page to load
			chromedp.Sleep(4000*time.Millisecond),
			chromedp.ActionFunc(func(ctx context.Context) error {

				id, count, err := dom.PerformSearch(selector).Do(ctx)
				if err != nil {
					return err
				}
				defer func() {
					_ = dom.DiscardSearchResults(id).Do(ctx)
				}()

				if count < 1 {
					node = nil
				} else {
					node, err = dom.GetSearchResults(id, 0, count).Do(ctx)
					if err != nil {
						log.Print("No enterprises were found")
						return err
					}
				}
				return nil
			}),
		)
		if err != nil {
			i = i + 1
		} else {
			break
		}
	}

	if err != nil {
		log.Printf("Error while performing the automation logic : %v", err)
	}

	if node != nil {
		c := chromedp.FromContext(ctx)
		html, err = dom.GetOuterHTML().WithNodeID(node[0]).Do(cdp.WithExecutor(ctx, c.Target))
		if err != nil {
			log.Printf("Error while getting the wttj search page HTML : %v ", err)
		}

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		if err != nil {
			log.Printf("Error while opening the HTML as a goquery document : %v", err)
		}

		link := doc.Find("a")
		url, exist := link.Attr("href")
		if !exist {
			log.Printf("Error while trying to get the content of the href link")
		}

		html = "https://www.welcometothejungle.com" + url
	}

	WTTJURL := html

	// Parse the url to separe it's components
	url, err := url.Parse(WTTJURL)
	if err != nil {
		log.Printf("%s", err)
	}

	// Delete the query part
	url.RawQuery = ""

	// Fetch back the parsed url
	WTTJURL = url.String()

	if !validateWTTJURL(ctx, WTTJURL, companyName) {
		WTTJURL = ""
	}

	return WTTJURL, err
}

// Define a function to scrape a company Welcome to the Jungle page and verify that the company name is present in its content.
func validateWTTJURL(ctx context.Context, companyURL string, companyName string) bool {

	searchURL, err := url.ParseRequestURI(companyURL)
	if err != nil {
		return false
	}

	// Create the request context
	ctx, cancel := createTab(ctx)
	defer cancel()

	selector := "sc-gdfaqJ cUUdcw"

	var node []cdp.NodeID
	html := ""
	for i := 0; i < 5; i++ {
		err = chromedp.Run(ctx,
			// visit the target page
			chromedp.Navigate(searchURL.String()),
			// wait for the page to load
			chromedp.Sleep(4000*time.Millisecond),
			chromedp.ActionFunc(func(ctx context.Context) error {

				id, count, err := dom.PerformSearch(selector).Do(ctx)
				if err != nil {
					return err
				}
				defer func() {
					_ = dom.DiscardSearchResults(id).Do(ctx)
				}()

				if count < 1 {
					log.Printf("There is no wttj url for %s", companyName)
					node = nil
				} else {
					node, err = dom.GetSearchResults(id, 0, count).Do(ctx)
					if err != nil {
						return err
					}
				}
				return nil
			}),
		)
		if err != nil {
			i = i + 1
		} else {
			break
		}
	}
	if err != nil {
		log.Printf("Error while performing the automation logic : %v", err)
	}

	title := ""
	if node != nil {
		c := chromedp.FromContext(ctx)
		html, err = dom.GetOuterHTML().WithNodeID(node[0]).Do(cdp.WithExecutor(ctx, c.Target))
		if err != nil {
			log.Printf("Error while getting the wttj company page HTML : %v ", err)
		}

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		if err != nil {
			log.Printf("Error while opening the company wttj HTML as a goquery document : %v", err)
		}

		title = doc.Find("h1").Text()
	}

	companyName = strings.ToLower(companyName)
	title = strings.ToLower(title)
	companyNameWithoutSpaces := strings.TrimSpace(companyName)

	if (strings.Contains(title, companyName)) || (strings.Contains(title, companyNameWithoutSpaces)) {
		return true
	} else {
		return false
	}
}

// Define a function to enrich a list of companies with their welcome to the jungle url.
func enrichCompanyWTTJUrl(ctx context.Context, db *pgxpool.Pool, companyName string) error {
	var err error

	company, exists, err := getCompany(db, companyName)
	switch {
	case err != nil:
		log.Printf("Database query failed because of : %s", err)
	case !exists:
		log.Printf("The company does not exist")
	default:
		if company.WTTJURL == "" {
			WTTJURL, err := findCompanyWTTJURL(ctx, companyName)
			if err != nil {
				return err
			}

			if WTTJURL == "" {
				log.Printf("%s wttj url has not been found", companyName)
			} else {
				log.Printf("WTTJ url has been found for %s", companyName)
				log.Printf("Updating database")
				err = updateCompanyWTTJURL(db, companyName, WTTJURL)
				if err != nil {
					log.Printf("An error happened with the query : %s", err)
					return err
				}
				log.Printf("%s has been enriched with it's wttj url : %s", companyName, WTTJURL)
			}

		} else {
			log.Printf("Company %s already has it's WTTJ url enriched.", companyName)
		}
	}

	return err
}
