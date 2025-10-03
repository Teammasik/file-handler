package main

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"os"
	"strings"
)

type CSVRecord map[string]string

// CSVToJSON читает CSV-файл по inputPath и записывает JSON-результат в outputPath
func CSVtoJSON(inputPath, outputPath string) error {
	jsonData, err := CSVtoJSONBytes(inputPath)
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, jsonData, 0644)
}

func CSVtoJSONBytes(inputPath string) ([]byte, error) {
	file, err := os.Open(inputPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Автоопределение разделителя: проверим первую строку
	firstLine, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			return []byte("[]"), nil // пустой файл
		}
		return nil, err
	}

	// Определяем разделитель по первой строке
	var separator rune = ','
	if strings.Contains(firstLine[0], "\t") {
		separator = '\t'
	} else if strings.Contains(firstLine[0], ";") {
		separator = ';'
	}

	// Пересоздаём reader с правильным разделителем
	file.Seek(0, 0) // возвращаемся в начало файла
	reader = csv.NewReader(file)
	reader.Comma = separator

	// Читаем все записи
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
