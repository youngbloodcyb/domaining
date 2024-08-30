package parser

import (
	"archive/zip"
	"domaining/database"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type DropCatchResponse struct {
	Result struct {
		FileUrl  string `json:"fileUrl"`
		FileName string `json:"fileName"`
	} `json:"result"`
	Success    bool   `json:"success"`
	StatusCode string `json:"statusCode"`
}

func FetchDropcatchUrl() (string, string, error) {
	url := "https://client.dropcatch.com/GetFileUrl?FileType=csv&RequestType=Auction&AuctionType=AllAuctions"

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", fmt.Errorf("error creating request: %v", err)
	}

	// Set headers (same as before)
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "en-US,en;q=0.9")
	req.Header.Set("dnt", "1")
	req.Header.Set("origin", "https://www.dropcatch.com")
	req.Header.Set("referer", "https://www.dropcatch.com/")
	req.Header.Set("sec-ch-ua", `"Chromium";v="127", "Not)A;Brand";v="99"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", "macOS")
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-site")
	req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("error reading response body: %v", err)
	}

	var dropCatchResp DropCatchResponse
	err = json.Unmarshal(body, &dropCatchResp)
	if err != nil {
		return "", "", fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	if !dropCatchResp.Success || dropCatchResp.StatusCode != "OK" {
		return "", "", fmt.Errorf("API request was not successful")
	}

	return dropCatchResp.Result.FileUrl, dropCatchResp.Result.FileName, nil
}

func DownloadDropcatchFile(fileUrl, fileName string) error {
	// Create the file
	out, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(fileUrl)
	if err != nil {
		return fmt.Errorf("error downloading file: %v", err)
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}

	return nil
}

func UnzipFile(zipFile string) (string, error) {
	// Open the zip file
	reader, err := zip.OpenReader(zipFile)
	if err != nil {
		return "", fmt.Errorf("error opening zip file: %v", err)
	}
	defer reader.Close()

	// We're assuming there's only one file in the zip
	if len(reader.File) != 1 {
		return "", fmt.Errorf("expected 1 file in zip, found %d", len(reader.File))
	}

	zippedFile := reader.File[0]
	csvFileName := zippedFile.Name

	// Create the file
	outFile, err := os.Create(csvFileName)
	if err != nil {
		return "", fmt.Errorf("error creating output file: %v", err)
	}
	defer outFile.Close()

	// Open the zipped file
	zippedData, err := zippedFile.Open()
	if err != nil {
		return "", fmt.Errorf("error opening zipped file: %v", err)
	}
	defer zippedData.Close()

	// Write the unzipped data to the output file
	_, err = io.Copy(outFile, zippedData)
	if err != nil {
		return "", fmt.Errorf("error writing unzipped data: %v", err)
	}

	return csvFileName, nil
}

func ParseCSVToSQLite(csvFilePath, dbFilePath string) error {
	// Open the CSV file
	file, err := os.Open(csvFilePath)
	if err != nil {
		return fmt.Errorf("error opening CSV file: %v", err)
	}
	defer file.Close()

	// Create a new CSV reader
	reader := csv.NewReader(file)

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("error reading CSV: %v", err)
	}

	if len(records) == 0 {
		return fmt.Errorf("CSV file is empty")
	}

	// Open the database
	db, err := database.New(dbFilePath)
	if err != nil {
		return fmt.Errorf("error opening database: %v", err)
	}
	defer db.Close()

	// Create the table
	columns := records[0]
	columnsSQL := make([]string, len(columns))
	for i, col := range columns {
		columnsSQL[i] = fmt.Sprintf("%s TEXT", col)
	}
	err = db.CreateTable("domains", columnsSQL)
	if err != nil {
		return fmt.Errorf("error creating table: %v", err)
	}

	// Insert the data
	for _, record := range records[1:] { // Skip the header row
		// Convert []string to []interface{}
		values := make([]interface{}, len(record))
		for i, v := range record {
			values[i] = v
		}

		err := db.InsertRecord("domains", columns, values)
		if err != nil {
			return fmt.Errorf("error inserting record: %v", err)
		}
	}

	fmt.Printf("Successfully inserted %d records into the database.\n", len(records)-1)
	return nil
}