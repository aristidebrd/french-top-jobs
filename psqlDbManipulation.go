package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"gopkg.in/yaml.v2"
)

func initDbConnection() (*pgxpool.Pool, error) {

	// Open the YAML file containing the API key.
	file, err := os.Open("secrets/db-infos.yaml")
	if err != nil {
		log.Fatal(err)
	}

	// Read the YAML file into a map.
	var data map[string]interface{}
	err = yaml.NewDecoder(file).Decode(&data)
	if err != nil {
		log.Fatal(err)
	}

	// Get the API key from the map.
	psql_db_host := data["psql_db_host"].(string)
	psql_db_port := data["psql_db_port"].(string)
	psql_db_user := data["psql_db_user"].(string)
	psql_db_password := data["psql_db_password"].(string)
	psql_db_name := data["psql_db_name"].(string)

	db_url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", psql_db_user, psql_db_password, psql_db_host, psql_db_port, psql_db_name)
	dbpool, err := pgxpool.New(context.Background(), db_url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}

	return dbpool, err
}
