package mtgdb

import (
	"fmt"
	"path/filepath"
	"strings"
)

func CardImagesDir(imagesDir string) string {
	return filepath.Join(imagesDir, "cards")
}

func SetImagesDir(imagesDir string) string {
	return filepath.Join(imagesDir, "sets")
}

func CardImagePath(imagesDir, setCode, collectorNumber string) string {
	return strings.ToLower(filepath.Join(CardImagesDir(imagesDir), setCode, fmt.Sprintf("%s_%s.jpg", setCode, collectorNumber)))
}

func SetImagePath(imagesDir, setCode string) string {
	return strings.ToLower(filepath.Join(SetImagesDir(imagesDir), fmt.Sprintf("%s.jpg", setCode)))
}
