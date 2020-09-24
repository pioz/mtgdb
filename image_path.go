package mtgdb

import (
	"fmt"
	"path/filepath"
)

func CardImagesDir(imagesDir string) string {
	return filepath.Join(imagesDir, "cards")
}

func SetImagesDir(imagesDir string) string {
	return filepath.Join(imagesDir, "sets")
}

func CardImagePath(imagesDir, setCode, collectorNumber, locale string, backImage bool) string {
	var fileName string
	if !backImage {
		fileName = fmt.Sprintf("%s_%s_%s.jpg", setCode, collectorNumber, locale)
	} else {
		fileName = fmt.Sprintf("%s_%s_%s_back.jpg", setCode, collectorNumber, locale)
	}
	return filepath.Join(CardImagesDir(imagesDir), setCode, fileName)
}

func SetImagePath(imagesDir, setCode string) string {
	return filepath.Join(SetImagesDir(imagesDir), fmt.Sprintf("%s.jpg", setCode))
}
