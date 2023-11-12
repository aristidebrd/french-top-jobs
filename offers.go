package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func createJobOffer(db *pgxpool.Pool, offer Offer) error {
	query := `INSERT INTO offers (company_name, offer_url) VALUES (@company_name, @offer_url)`
	args := pgx.NamedArgs{
		"company_name": offer.companyName,
		"offer_url":    offer.offerUrl,
	}
	_, err := db.Exec(context.TODO(), query, args)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch {
			case pgErr.Code == "23505":
				log.Println("The entry already exist, not adding it")
				return nil
			default:
				return fmt.Errorf("unable to insert row: %w", err)
			}
		}
	}
	return err
}

func deleteJobOffer(db *pgxpool.Pool, offer_url string) error {
	query := `DELETE FROM offers WHERE offer_url = @offerIdToDelete`
	args := pgx.NamedArgs{
		"offerIdToDelete": offer_url,
	}
	_, err := db.Exec(context.Background(), query, args)
	if err != nil {
		return fmt.Errorf("unable to delete row: %w", err)
	}

	return err
}

func getCompanyOffers(db *pgxpool.Pool, companyName string) []Offer {
	var offers []Offer

	query := "select * from offers where name = @companyName"
	args := pgx.NamedArgs{
		"companyName": companyName,
	}
	rows, err := db.Query(context.TODO(), query, args)

	switch {
	case err == pgx.ErrNoRows:
		return offers
	case err != nil:
		log.Printf("Database query failed because of %s :", err)
	default:
		for rows.Next() {
			var offer Offer
			err = rows.Scan(&offer.companyName, &offer.offerUrl)
			if err != nil {
				log.Printf("Database query scan failed because of %s :", err)
			}
			offers = append(offers, offer)
		}
	}

	return offers
}

func getAllOffers(db *pgxpool.Pool) []Offer {
	var offers []Offer

	query := "select * from offers"
	rows, err := db.Query(context.TODO(), query)

	switch {
	case err == pgx.ErrNoRows:
		return offers
	case err != nil:
		log.Printf("Database query failed because of %s :", err)
	default:
		for rows.Next() {
			var offer Offer
			err = rows.Scan(&offer.companyName, &offer.offerUrl)
			if err != nil {
				log.Printf("Database query scan failed because of %s :", err)
			}
			offers = append(offers, offer)
		}
	}

	return offers
}
