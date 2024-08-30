package godaddy

import (
	"domaining/database"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Base URL for GoDaddy API
const baseURL = "https://auctions.godaddy.com/beta/findApiProxy/v4/aftermarket/find/auction/recommend"

// FetchAndParseCSV fetches the CSV directly from GoDaddy and inserts it into a SQLite database
func FetchAndParseCSV(dbFilePath string) error {
	// Get the current time and format it as ISO 8601 with Zulu time
	currentTime := time.Now()
	formattedDate := currentTime.Format("2006-01-02T15:04:05.000Z")

	// Construct the full URL with query parameters
	params := url.Values{}
	params.Add("endTimeAfter", formattedDate)
	params.Add("exportCSV", "true")

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// Create HTTP client and request
	client := &http.Client{}
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	// Set headers for the request
	req.Header.Set("accept", "text/csv, application/json, text/plain, */*")
	req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36")

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Parse the CSV directly from the response body
	err = parseCSVToSQLite(resp.Body, dbFilePath)
	if err != nil {
		return fmt.Errorf("error parsing CSV to SQLite: %v", err)
	}

	return nil
}

// parseCSVToSQLite reads a CSV from an io.Reader and inserts the data into a SQLite database
func parseCSVToSQLite(csvReader io.Reader, dbFilePath string) error {
	// Create a new CSV reader
	reader := csv.NewReader(csvReader)
	reader.FieldsPerRecord = -1 // Allow variable number of fields per record

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("error reading CSV: %v", err)
	}

	if len(records) < 2 {
		return fmt.Errorf("CSV file has no data rows")
	}

	// Open the database
	db, err := database.New(dbFilePath)
	if err != nil {
		return fmt.Errorf("error opening database: %v", err)
	}
	defer db.Close()

	// Define hard-coded column names based on the CSV structure
	columns := []string{"item_id", "domain_name", "traffic", "bids", "price", "estimated_value", "domain_age", "auction_end_time", "sale_type", "majestic_tf", "majestic_cf", "backlinks", "referring_domains"}

	// Define the SQL column types (assuming all are TEXT for simplicity, adjust types as needed)
	columnsSQL := []string{
		"`item_id` INTEGER",
		"`domain_name` TEXT",
		"`traffic` INTEGER",
		"`bids` INTEGER",
		"`price` INTEGER",
		"`estimated_value` INTEGER",
		"`domain_age` INTEGER",
		"`auction_end_time` TEXT",
		"`sale_type` TEXT",
		"`majestic_tf` INTEGER",
		"`majestic_cf` INTEGER",
		"`backlinks` INTEGER",
		"`referring_domains` INTEGER",
	}

	// Create the table with hard-coded column definitions
	err = db.CreateTable("godaddy", columnsSQL)
	if err != nil {
		return fmt.Errorf("error creating table: %v", err)
	}

	// Insert the data
	for _, record := range records[1:] { // Skip the header row
		// Convert []string to []interface{} and handle potential conversion errors
		values := []interface{}{
			record[0], // item_id
			record[1], // domain_name
			record[2], // traffic
			record[3], // bids
			record[4], // price
			record[5], // estimated_value
			record[6], // domain_age
			record[7], // auction_end_time
			record[8], // sale_type
			record[9], // majestic_tf
			record[10], // majestic_cf
			record[11], // backlinks
			record[12], // referring_domains
		}

		err := db.InsertRecord("godaddy", columns, values)
		if err != nil {
			return fmt.Errorf("error inserting record: %v", err)
		}
	}

	fmt.Printf("Successfully inserted %d records into the database.\n", len(records)-1)
	return nil
}
