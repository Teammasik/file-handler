package processor

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"os"
	"strings"
)

type CSVRecord map[string]string

func CSVToJSON(inputPath, outputPath string) error {
	jsonData, err := CSVToJSONBytes(inputPath)
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, jsonData, 0644)
}

func CSVToJSONBytes(inputPath string) ([]byte, error) {
	file, err := os.Open(inputPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	firstLine, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			return []byte("[]"), nil
		}
		return nil, err
	}

	var separator rune = ','
	if strings.Contains(firstLine[0], "\t") {
		separator = '\t'
	} else if strings.Contains(firstLine[0], ";") {
		separator = ';'
	}

	file.Seek(0, 0)
	reader = csv.NewReader(file)
	reader.Comma = separator

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return []byte("[]"), nil
	}

	headers := records[0]
	dataRows := records[1:]

	var jsonRecords []CSVRecord
	for _, row := range dataRows {
		record := make(CSVRecord)
		for i, value := range row {
			if i < len(headers) {
				record[headers[i]] = value
			}
		}
		jsonRecords = append(jsonRecords, record)
	}

	return json.MarshalIndent(jsonRecords, "", "  ")
}
