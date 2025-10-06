package processor

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/extractor"
	"github.com/unidoc/unipdf/v3/model"
)

type PDFInfo struct {
	Format       string `json:"format"`
	SizeBytes    int64  `json:"size_bytes"`
	Pages        int    `json:"pages_count"`
	Title        string `json:"title,omitempty"`
	Author       string `json:"author,omitempty"`
	Subject      string `json:"subject,omitempty"`
	Creator      string `json:"creator,omitempty"`
	Producer     string `json:"producer,omitempty"`
	Created      string `json:"created,omitempty"`
	Modified     string `json:"modified,omitempty"`
	CharCount    int    `json:"char_count,omitempty"`
	WordCount    int    `json:"word_count,omitempty"`
	ErrorMessage string `json:"error,omitempty"`
}

func ProcessPDF(filePath string) ([]byte, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	pdfReader, _, err := model.NewPdfReaderFromFile(filePath, nil)
	if err != nil {
		return json.Marshal(PDFInfo{
			Format:       "pdf",
			SizeBytes:    fileInfo.Size(),
			ErrorMessage: "Invalid PDF: " + err.Error(),
		})
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		numPages = 0
	}

	info := PDFInfo{
		Format:    "pdf",
		SizeBytes: fileInfo.Size(),
		Pages:     numPages,
	}

	// Getting metadata, may be empty, cuz unipdf requires payment(
	trailer, err := pdfReader.GetTrailer()
	if err != nil {
		if infoDict, ok := trailer.Get("Info").(*core.PdfObjectDictionary); ok {
			info.Title = getString(infoDict, core.PdfObjectName("Title"))
			info.Author = getString(infoDict, core.PdfObjectName("Author"))
			info.Subject = getString(infoDict, core.PdfObjectName("Subject"))
			info.Creator = getString(infoDict, core.PdfObjectName("Creator"))
			info.Producer = getString(infoDict, core.PdfObjectName("Producer"))
			info.Created = getString(infoDict, core.PdfObjectName("CreationDate"))
			info.Modified = getString(infoDict, core.PdfObjectName("ModDate"))
		}
	}

	var fullText strings.Builder
	for i := 1; i <= numPages; i++ {
		page, err := pdfReader.GetPage(i)
		if err != nil {
			continue
		}

		ex, err := extractor.New(page)
		if err != nil {
			continue
		}

		text, err := ex.ExtractText()
		if err != nil {
			continue
		}
		fullText.WriteString(text)
	}

	textStr := fullText.String()
	info.CharCount = len(textStr)
	info.WordCount = len(strings.Fields(textStr))

	return json.MarshalIndent(info, "", "  ")
}

func getString(dict *core.PdfObjectDictionary, key core.PdfObjectName) string {
	if obj := dict.Get(key); obj != nil {
		if str, ok := obj.(*core.PdfObjectString); ok {
			return str.Decoded()
		}
	}
	return ""
}
