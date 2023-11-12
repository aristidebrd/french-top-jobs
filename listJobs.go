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

type Offer struct {
	companyName string
	offerUrl    string
}

// This function finds all links on a website and returns them as a list
func findAllLinks(ctx context.Context, website string) []string {

	var links []string

	// Create the request context
	ctx, cancel := createTab(ctx)
	defer cancel()

	html := ""
	var err error
	for i := 0; i < 5; i++ {
		err := chromedp.Run(ctx,
			// visit the target page
			chromedp.Navigate(website),
			// wait for the page to load
			chromedp.Sleep(4000*time.Millisecond),
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

	c := chromedp.FromContext(ctx)
	rootNode, err := dom.GetDocument().Do(cdp.WithExecutor(ctx, c.Target))
	if err != nil {
		log.Print(err)
	}

	html, err = dom.GetOuterHTML().WithNodeID(rootNode.NodeID).Do(cdp.WithExecutor(ctx, c.Target))
	if err != nil {
		log.Print(err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		log.Print(err)
	}

	// Find all the href links in the HTML document
	hrefs := doc.Find("a")

	// Iterate over the links and print them to the console
	for i := 0; i < hrefs.Length(); i++ {
		href := hrefs.Eq(i)
		// Get the href attribute of the link
		link, ok := href.Attr("href")
		if ok {
			link = strings.TrimSpace(link)
			if strings.Contains(website, "welcometothejungle") {
				website = "https://www.welcometothejungle.com"
			}
			link = getAbsoluteUrl(website, link)
			links = append(links, link)
		}
	}

	return links
}

// This function verify that the url is about a devops job
func isDevopsJobUrl(website string) bool {
	// Create a regular expression to match the URL of the jobs page
	regex := regexp.MustCompile(".*(devops|dev_ops|dev-ops|devsecops|dev_sec_ops|dev-sec-ops|sre|site_reliability_engineer|site-reliability-engineer).*")
	// pageTitle := getPageTitle(website)
	if regex.MatchString(website) {
		log.Printf("This url is a devops job : %s", website)
		return true
	} else {
		return false
	}
}

// This function take as parameter a company and add to the database the devops jobs url found on it's job page
func addJobs(ctx context.Context, db *pgxpool.Pool, company Company) {

	if company.JobsPageURL == "" {
		log.Printf("%s does not have a job page url, exiting", company.Name)
		return
	}

	log.Printf("%s has this job url : %s", company.Name, company.JobsPageURL)

	links := findAllLinks(ctx, company.JobsPageURL)

	for _, link := range links {
		if isDevopsJobUrl(link) {
			newOffer := Offer{company.Name, link}
			err := createJobOffer(db, newOffer)
			if err != nil {
				log.Printf("An error happened with the query : %s", err)
			}
		}
	}

}
