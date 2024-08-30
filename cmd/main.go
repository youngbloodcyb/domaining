package main

import (
	"fmt"
	"log"
	"os"

	"domaining/database"
	"domaining/dropcatch"
	"domaining/utils"
)

func main() {
	// Step 1: Fetch the CSV URL
	fileUrl, fileName, err := dropcatch.FetchCSVUrl()
	if err != nil {
		log.Fatalf("Error fetching CSV URL: %v", err)
	}

	fmt.Printf("Fetched file URL: %s\n", fileUrl)
	fmt.Printf("File name: %s\n", fileName)

	// Step 2: Download the CSV file
	err = dropcatch.DownloadCSVFile(fileUrl, fileName)
	if err != nil {
		log.Fatalf("Error downloading CSV file: %v", err)
	}

	fmt.Printf("Successfully downloaded: %s\n", fileName)

	// Step 3: Unzip the file
	csvFileName, err := utils.UnzipFile(fileName)
	if err != nil {
		log.Fatalf("Error unzipping file: %v", err)
	}

	fmt.Printf("Successfully unzipped to: %s\n", csvFileName)

	// Step 4: Initialize the database
	dbPath := "domains.db"
	db, err := database.New(dbPath)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	defer db.Close()

	// Step 5: Parse CSV and load into SQLite
	err = dropcatch.ParseCSVToSQLite(csvFileName, dbPath)
	if err != nil {
		log.Fatalf("Error parsing CSV to SQLite: %v", err)
	}

	fmt.Println("CSV data successfully loaded into SQLite database.")

	// Optional: Clean up downloaded files
	err = os.Remove(fileName)
	if err != nil {
		fmt.Printf("Warning: Could not remove zip file: %v\n", err)
	}

	err = os.Remove(csvFileName)
	if err != nil {
		fmt.Printf("Warning: Could not remove CSV file: %v\n", err)
	}

	fmt.Println("Process completed successfully.")
}