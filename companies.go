package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// This variable stores the name of a company and the website associated
type Company struct {
	Name        string
	IsTop500    bool
	Website     string
	LinkedInURL string
	WTTJURL     string
	JobsPageURL string
}

func addCompany(db *pgxpool.Pool, company Company) error {
	query := `INSERT INTO companies (name, is_top_500, website_url, linkedin_url, wttj_url, job_page_url) VALUES (@name, @isTop500, @website_url, @linkedin_url, @wttj_url, @job_page_url)`
	args := pgx.NamedArgs{
		"name":         company.Name,
		"isTop500":     company.IsTop500,
		"website_url":  company.Website,
		"linkedin_url": company.LinkedInURL,
		"wttj_url":     company.WTTJURL,
		"job_page_url": company.JobsPageURL,
	}
	_, err := db.Exec(context.TODO(), query, args)
	if err != nil {
		return fmt.Errorf("unable to insert row: %w", err)
	}

	return err
}

func addMultipleCompanies(db *pgxpool.Pool, companies []Company) error {
	var rows [][]interface{}
	for _, company := range companies {
		companySlice := []interface{}{company.Name, company.IsTop500, company.Website, company.LinkedInURL, company.WTTJURL, company.JobsPageURL}
		rows = append(rows, companySlice)
	}
	_, err := db.CopyFrom(
		context.TODO(),
		pgx.Identifier{"name"},
		[]string{"name", "is_top_500", "website_url", "linkedin_url", "wttj_url", "job_page_url"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return err
	}

	return err
}

func updatecompanyValue(db *pgxpool.Pool, companyName string, value string, content any) error {
	query := `UPDATE companies SET ` + value + ` = @content WHERE name = @companyName`
	args := pgx.NamedArgs{
		"companyName": companyName,
		"content":     content,
	}
	_, err := db.Exec(context.Background(), query, args)
	if err != nil {
		return fmt.Errorf("unable to update row: %w", err)
	}

	return err
}

func updateCompanyLinkedinURL(db *pgxpool.Pool, companyName string, content string) error {
	err := updatecompanyValue(db, companyName, "linkedin_url", content)
	return err
}

func updateCompanyWTTJURL(db *pgxpool.Pool, companyName string, content string) error {
	err := updatecompanyValue(db, companyName, "wttj_url", content)
	return err
}

func updateCompanyWebsiteURL(db *pgxpool.Pool, companyName string, content string) error {
	err := updatecompanyValue(db, companyName, "website_url", content)
	return err
}

func updateCompany(db *pgxpool.Pool, company Company) error {
	query := `UPDATE companies SET name = @name, is_top_500 = @isTop500, website_url = @website_url, linkedin_url = @linkedin_url, wttj_url = @wttj_url, job_page_url = @job_page_url WHERE name = @companyToUpdate`
	args := pgx.NamedArgs{
		"name":            company.Name,
		"isTop500":        company.IsTop500,
		"website_url":     company.Website,
		"linkedin_url":    company.LinkedInURL,
		"wttj_url":        company.WTTJURL,
		"job_page_url":    company.JobsPageURL,
		"companyToUpdate": company.Name,
	}
	_, err := db.Exec(context.Background(), query, args)
	if err != nil {
		return fmt.Errorf("unable to update row: %w", err)
	}

	return err
}

func getCompany(db *pgxpool.Pool, companyName string) (Company, bool, error) {
	var company Company

	exists := false

	query := "select * from companies where name = @companyName"
	args := pgx.NamedArgs{
		"companyName": companyName,
	}
	row := db.QueryRow(context.TODO(), query, args)
	err := row.Scan(&company.Name, &company.IsTop500, &company.Website, &company.LinkedInURL, &company.WTTJURL, &company.JobsPageURL)
	switch {
	case err == pgx.ErrNoRows:
		err = nil
	case err != nil:
		log.Printf("Database query failed because of %s :", err)
	default:
		exists = true
	}

	return company, exists, err
}

func getAllCompanies(db *pgxpool.Pool) []Company {
	var companies []Company

	query := "select * from companies"

	rows, err := db.Query(context.TODO(), query)
	if err != nil {
		log.Printf("Database query failed because of %s :", err)
	}
	for rows.Next() {
		var company Company
		err = rows.Scan(&company.Name, &company.IsTop500, &company.Website, &company.LinkedInURL, &company.WTTJURL, &company.JobsPageURL)
		if err != nil {
			log.Printf("Database query scan failed because of %s :", err)
		}
		companies = append(companies, company)
	}

	return companies
}

func deleteCompany(db *pgxpool.Pool, companyName string) error {
	query := `DELETE FROM companies WHERE name = @companyToDelete`
	args := pgx.NamedArgs{
		"companyToDelete": companyName,
	}
	_, err := db.Exec(context.Background(), query, args)
	if err != nil {
		return fmt.Errorf("unable to delete row: %w", err)
	}

	return err
}

func getAllCompaniesAPI(c *gin.Context) {
	companies := getAllCompanies(dbpoolapi)
	c.JSON(http.StatusOK, gin.H{"data": companies})
}

func getCompanyAPI(c *gin.Context) {
	company, exists, _ := getCompany(dbpoolapi, c.Query("name"))
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": company})
}
