package processor

import (
	"encoding/json"
	"os"
	"time"

	"github.com/xuri/excelize/v2"
)

type ExcelInfo struct {
	Format       string     `json:"format"`
	SizeBytes    int64      `json:"size_bytes"`
	Sheets       int        `json:"sheets_count,omitempty"`
	Title        string     `json:"title,omitempty"`
	Subject      string     `json:"subject,omitempty"`
	Creator      string     `json:"creator,omitempty"`
	Created      *time.Time `json:"created,omitempty"`
	Modified     *time.Time `json:"modified,omitempty"`
	Description  string     `json:"description,omitempty"`
	ErrorMessage string     `json:"error,omitempty"`
}

// ProcessExcel handles .xlsx files, return metadata
func ProcessExcel(filePath string) ([]byte, error) {
	// Получаем размер файла
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	info := ExcelInfo{
		Format:    "xlsx",
		SizeBytes: fileInfo.Size(),
	}

	// attempt to open as .xlsx
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return json.Marshal(ExcelInfo{
			Format:       "xls/xlsx",
			SizeBytes:    fileInfo.Size(),
			ErrorMessage: "Unsupported or corrupted Excel file: " + err.Error(),
		})
	}
	defer f.Close()

	sheets := f.GetSheetList()
	info.Sheets = len(sheets)


	props, err := f.GetDocProps()
	if err != nil {
		info.ErrorMessage = "Could not read document properties: " + err.Error()
	} else {
		info.Title = props.Title
		info.Subject = props.Subject
		info.Creator = props.Creator
		info.Description = props.Description

		if props.Created != "" {
			if t, err := time.Parse(time.RFC3339, props.Created); err == nil {
				info.Created = &t
			}
		}
		if props.Modified != "" {
			if t, err := time.Parse(time.RFC3339, props.Modified); err == nil {
				info.Modified = &t
			}
		}
	}

	return json.MarshalIndent(info, "", "  ")
}
