package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
)

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