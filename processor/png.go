package processor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/png"
	"os"
)

type PNGInfo struct {
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	ColorModel   string `json:"color_model"`
	BitDepth     int    `json:"bit_depth,omitempty"`
	Format       string `json:"format"`
	ErrorMessage string `json:"error,omitempty"`
}

// ProcessPNG reads PNG, returns info about it
func ProcessPNG(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// reading only metadata without image decoding
	img, err := png.DecodeConfig(file)
	if err != nil {
		return json.Marshal(PNGInfo{
			Format:       "png",
			ErrorMessage: "Invalid PNG file: " + err.Error(),
		})
	}

	info := PNGInfo{
		Width:      img.Width,
		Height:     img.Height,
		Format:     "png",
		ColorModel: colorModelToString(img.ColorModel),
	}

	// attempt to get bit depth (may not always work thru DecodeConfig)
	if bitDepth, ok := getBitDepthFromPNG(filePath); ok {
		info.BitDepth = bitDepth
	}

	return json.MarshalIndent(info, "", "  ")
}

func colorModelToString(cm interface{}) string {
	switch cm {
	case nil:
		return "unknown"
	default:
		return fmt.Sprintf("%T", cm)
	}
}

// getBitDepthFromPNG tries to get bit depth from PNG heading
func getBitDepthFromPNG(filePath string) (int, bool) {
	data, err := os.ReadFile(filePath)
	if err != nil || len(data) < 24 {
		return 0, false
	}

	// checking PNG's signature (first 8 bytes)
	pngSignature := []byte{137, 80, 78, 71, 13, 10, 26, 10}
	if !bytes.Equal(data[:8], pngSignature) {
		return 0, false
	}

	// IHDR chunk starts from 8th byte
	// IHDR struc: width(4), height(4), bit_depth(1), color_type(1)
	if len(data) < 25 {
		return 0, false
	}

	bitDepth := int(data[24]) // 8 + 16 (len of chunk + "IHDR") + 8 (width+height) = 24
	return bitDepth, true
}
