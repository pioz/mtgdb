package mtgdb

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pioz/mtgdb/pb"
)

type Importer struct {
	DataDir                  string
	ImagesDir                string
	OnlyTheseSetCodes        []string
	ForceDownloadData        bool
	DownloadAssets           bool
	DownloadOnlyEnAssets     bool
	ForceDownloadOlderAssets bool
	ForceDownloadDiffSha1    bool
	ForceDownloadAssets      bool
	ImageType                string

	cardCollection        map[string]*Card
	setCollection         map[string]*Set
	setIconsDownloaded    map[string]struct{}
	notEnImagesToDownload map[string]*cardJsonStruct

	errorsChan          chan error
	wg                  sync.WaitGroup
	downloaderSemaphore chan struct{}
	bar                 *pb.ProgressBar
}

func NewImporter(dataDir string) *Importer {
	return &Importer{
		DataDir:                  dataDir,
		OnlyTheseSetCodes:        []string{},
		ImagesDir:                filepath.Join(dataDir, "images"),
		ForceDownloadData:        false,
		DownloadAssets:           true,
		DownloadOnlyEnAssets:     true,
		ForceDownloadOlderAssets: false,
		ForceDownloadDiffSha1:    false,
		ForceDownloadAssets:      false,
		ImageType:                "normal",
		errorsChan:               make(chan error, 10),
		downloaderSemaphore:      make(chan struct{}, 50),
	}
}

func (importer *Importer) SetDownloadConcurrency(n int) {
	importer.downloaderSemaphore = make(chan struct{}, n)
}

func (importer *Importer) DownloadData() error {
	createDirIfNotExist(importer.DataDir)

	allSetsJsonFilePath := filepath.Join(importer.DataDir, "all_sets.json")
	allCardsJsonFilePath := filepath.Join(importer.DataDir, "all_cards.json")
	if _, err := os.Stat(allSetsJsonFilePath); importer.ForceDownloadData || os.IsNotExist(err) {
		err := downloadFile(allSetsJsonFilePath, "https://api.scryfall.com/sets")
		if err != nil {
			return err
		}
	}
	if _, err := os.Stat(allCardsJsonFilePath); importer.ForceDownloadData || os.IsNotExist(err) {
		allCardsDataUrl, err := fetchAllCardsDataUrl()
		if err != nil {
			return err
		}
		err = downloadFile(allCardsJsonFilePath, allCardsDataUrl)
		if err != nil {
			return err
		}
	}
	return nil
}

func (importer *Importer) BuildCardsFromJson() []Card {
	defer removeAllFilesByExtension(SetImagesDir(importer.ImagesDir), "svg")

	importer.cardCollection = make(map[string]*Card)
	importer.setCollection = make(map[string]*Set)
	if importer.DownloadAssets {
		createDirIfNotExist(SetImagesDir(importer.ImagesDir))
		importer.setIconsDownloaded = make(map[string]struct{})
		importer.notEnImagesToDownload = make(map[string]*cardJsonStruct)
	}

	setsJson := setsJsonStruct{}
	err := loadFile(filepath.Join(importer.DataDir, "all_sets.json"), &setsJson)
	if err != nil {
		panic(err)
	}
	for _, setJson := range setsJson.Data {
		if len(importer.OnlyTheseSetCodes) != 0 && !contains(importer.OnlyTheseSetCodes, setJson.Code) {
			continue
		}
		importer.buildSet(&setJson)
	}

	if importer.DownloadAssets {
		importer.bar = pb.New("Download images", 0)
	}

	streamer, err := NewJsonStreamer(filepath.Join(importer.DataDir, "all_cards.json"))
	if err != nil {
		panic(err)
	}
	for streamer.Next() {
		var cardJson cardJsonStruct
		err := streamer.Get(&cardJson)
		if err != nil {
			panic(err)
		}
		if len(importer.OnlyTheseSetCodes) != 0 && !contains(importer.OnlyTheseSetCodes, cardJson.SetCode) {
			continue
		}
		importer.buildCard(&cardJson)
	}

	if importer.DownloadAssets {
		for _, cardJson := range importer.notEnImagesToDownload {
			importer.bar.IncrementMax()
			importer.wg.Add(1)
			go importer.downloadCardImage(*cardJson, "en")
		}
		importer.bar.Finishln()
	}

	waitErrors(&importer.wg, importer.errorsChan)
	close(importer.errorsChan)
	cards := make([]Card, 0, len(importer.cardCollection))
	for _, card := range importer.cardCollection {
		cards = append(cards, *card)
	}
	return cards
}

func AutoMigrate(db *gorm.DB) {
	db.AutoMigrate(&Set{})
	db.Model(&Set{}).AddUniqueIndex("idx_sets_code", "code")
	db.Model(&Set{}).AddIndex("idx_sets_parent_code", "parent_code")
	db.AutoMigrate(&Card{})
	db.Model(&Card{}).AddUniqueIndex("idx_cards_set_code_collector_number_is_token", "set_code", "collector_number", "is_token")
	db.Model(&Card{}).AddIndex("idx_cards_en_name", "en_name")
	db.Model(&Card{}).AddForeignKey("set_code", "sets(code)", "RESTRICT", "RESTRICT")
}

func BulkInsert(db *gorm.DB, cards []Card) error {
	sets := make(map[string]*Set)
	for _, card := range cards {
		if _, found := sets[card.SetCode]; !found && card.SetCode != "" {
			sets[card.SetCode] = card.Set
		}
	}
	allSets := make([]Set, 0, len(sets))
	for _, set := range sets {
		allSets = append(allSets, *set)
	}
	err := bulkInsert(db, Set{}, allSets, 1000)
	if err != nil {
		return err
	}
	return bulkInsert(db, Card{}, cards, 1000)
}

// Object must be an array of pointers
func bulkInsert(db *gorm.DB, table interface{}, objects interface{}, bulkSize int) error {
	scope := db.NewScope(table)
	fields := scope.Fields()
	quoted := make([]string, 0, len(fields))
	placeholders := make([]string, 0, len(fields))
	onUpdate := make([]string, 0, len(fields))
	for _, field := range fields {
		if (field.IsPrimaryKey && field.IsBlank) || field.IsIgnored || !field.IsNormal {
			continue
		}
		quoted = append(quoted, field.DBName)
		placeholders = append(placeholders, "?")
		onUpdate = append(onUpdate, fmt.Sprintf("%s = VALUES(%s)", field.DBName, field.DBName))
	}
	slice := reflect.ValueOf(objects)
	if slice.Kind() != reflect.Slice {
		panic("BulkInsert `objects` param given a non-slice type")
	}
	for i := 0; i < slice.Len(); i += bulkSize {
		allPlacehoders := make([]string, 0, bulkSize*len(fields))
		allValues := make([]interface{}, 0, bulkSize*len(fields))
		end := i + bulkSize
		if end > slice.Len() {
			end = slice.Len()
		}
		subslice := slice.Slice(i, end)
		for j := 0; j < subslice.Len(); j++ {
			obj := subslice.Index(j).Interface()
			for _, field := range fields {
				if (field.IsPrimaryKey && field.IsBlank) || field.IsIgnored || !field.IsNormal {
					continue
				}
				allValues = append(allValues, reflect.ValueOf(obj).FieldByName(field.Name).Interface())
			}
			allPlacehoders = append(allPlacehoders, fmt.Sprintf("(%s)", strings.Join(placeholders, ",")))
		}
		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s ON DUPLICATE KEY UPDATE %s;",
			scope.QuotedTableName(),
			strings.Join(quoted, ","),
			strings.Join(allPlacehoders, ","),
			strings.Join(onUpdate, ","),
		)
		err := db.Exec(query, allValues...).Error
		if err != nil {
			return err
		}
	}
	// return FillMissingTranslations(db)
	return nil
}

func FillMissingTranslations(db *gorm.DB) error {
	return db.Exec(`
		UPDATE cards
		JOIN cards AS main_cards ON cards.en_name = main_cards.en_name
		JOIN sets ON cards.set_code = sets.code
		SET
			cards.es_name  = main_cards.es_name,
			cards.fr_name  = main_cards.fr_name,
			cards.de_name  = main_cards.de_name,
			cards.it_name  = main_cards.it_name,
			cards.pt_name  = main_cards.pt_name,
			cards.ja_name  = main_cards.ja_name,
			cards.ko_name  = main_cards.ko_name,
			cards.ru_name  = main_cards.ru_name,
			cards.zhs_name = main_cards.zhs_name,
			cards.zht_name = main_cards.zht_name
		WHERE
		  main_cards.set_code = sets.parent_code
	`).Error
}

// PRIVATE types

type bulkDataJsonStruct struct {
	Type        string `json:"type"`
	DownloadUri string `json:"download_uri"`
}

type bulkDataArrayJsonStruct struct {
	Data []bulkDataJsonStruct `json:"data"`
}

type setsJsonStruct struct {
	Data []setJsonStruct `json:"data"`
}

type setJsonStruct struct {
	Name          string `json:"name"`
	Code          string `json:"code"`
	ReleasedAt    string `json:"released_at"`
	IconSvgUri    string `json:"icon_svg_uri"`
	ParentSetCode string `json:"parent_set_code"`
}

func (setJson *setJsonStruct) getIconName() string {
	basename := filepath.Base(setJson.IconSvgUri)
	return strings.Split(basename, ".")[0]
}

func (setJson *setJsonStruct) getParentCode() (code string) {
	if setJson.ParentSetCode != "" {
		code = setJson.ParentSetCode
	} else {
		code = setJson.Code
	}
	return code
}

type cardJsonStruct struct {
	Id              string               `json:"id"`
	Name            string               `json:"name"`
	PrintedName     string               `json:"printed_name"`
	Lang            string               `json:"lang"`
	ReleasedAt      string               `json:"released_at"`
	ImageUris       imagesCardJsonStruct `json:"image_uris"`
	CardFaces       []cardFaceStruct     `json:"card_faces"`
	SetCode         string               `json:"set"`
	SetName         string               `json:"set_name"`
	SetType         string               `json:"set_type"`
	CollectorNumber string               `json:"collector_number"`
	Foil            bool                 `json:"foil"`
	NonFoil         bool                 `json:"nonfoil"`
}

type imagesCardJsonStruct struct {
	Png    string `json:"png"`
	Large  string `json:"large"`
	Normal string `json:"normal"`
	Small  string `json:"small"`
}

func (imagesCardJson *imagesCardJsonStruct) GetImageByTypeName(name string) string {
	switch name {
	case "png":
		return imagesCardJson.Png
	case "large":
		return imagesCardJson.Large
	case "normal":
		return imagesCardJson.Normal
	case "small":
		return imagesCardJson.Small
	default:
		return imagesCardJson.Normal
	}
}

type cardFaceStruct struct {
	PrintedName string               `json:"printed_name"`
	ImageUris   imagesCardJsonStruct `json:"image_uris"`
}

// PRIVATE functions

func fetchAllCardsDataUrl() (string, error) {
	resp, err := http.Get("https://api.scryfall.com/bulk-data")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var bulkDataArray bulkDataArrayJsonStruct
	var cardsUrl string
	err = json.Unmarshal(body, &bulkDataArray)
	if err != nil {
		return "", err
	}
	for _, bulkData := range bulkDataArray.Data {
		if bulkData.Type == "all_cards" {
			cardsUrl = bulkData.DownloadUri
		}
	}
	if cardsUrl == "" {
		return "", errors.New("Impossible to retrieve all cards data url")
	}
	return cardsUrl, nil
}

func (importer *Importer) buildSet(setJson *setJsonStruct) {
	iconName := setJson.getIconName()
	if _, found := importer.setCollection[setJson.Code]; !found {
		set := &Set{
			Name:       setJson.Name,
			Code:       setJson.Code,
			ParentCode: setJson.getParentCode(),
			IconName:   iconName,
		}
		releasedAt, err := time.Parse("2006-01-02", setJson.ReleasedAt)
		if err == nil {
			set.ReleasedAt = &releasedAt
		}
		importer.setCollection[setJson.Code] = set

		if importer.DownloadAssets {
			createDirIfNotExist(filepath.Join(CardImagesDir(importer.ImagesDir), setJson.Code))
			if _, found := importer.setIconsDownloaded[iconName]; !found {
				importer.setIconsDownloaded[iconName] = struct{}{}
				importer.wg.Add(1)
				go importer.downloadSetIcon(*setJson)
			}
		}
	}
}

func (importer *Importer) buildCard(cardJson *cardJsonStruct) {
	isToken := false
	if cardJson.SetType == "token" {
		isToken = true
	}
	key := fmt.Sprintf("%s-%s", cardJson.SetCode, cardJson.CollectorNumber)
	card, found := importer.cardCollection[key]
	if !found {
		card = &Card{
			ScryfallId:      cardJson.Id,
			EnName:          cardJson.Name,
			SetCode:         cardJson.SetCode,
			CollectorNumber: cardJson.CollectorNumber,
			IsToken:         isToken,
			HasBackSide:     hasBackSide(cardJson),
			Set:             importer.setCollection[cardJson.SetCode],
			Foil:            cardJson.Foil,
			NonFoil:         cardJson.NonFoil,
		}
		importer.cardCollection[key] = card
		if importer.DownloadAssets {
			importer.notEnImagesToDownload[key] = cardJson
		}
	}
	if importer.DownloadAssets && (!importer.DownloadOnlyEnAssets || cardJson.Lang == "en") {
		importer.bar.IncrementMax()
		importer.wg.Add(1)
		go importer.downloadCardImage(*cardJson, cardJson.Lang)
		if cardJson.Lang == "en" {
			delete(importer.notEnImagesToDownload, key)
		}
	}
	printedName := cardJson.PrintedName
	if printedName == "" && len(cardJson.CardFaces) > 1 {
		printedName = fmt.Sprintf("%s // %s", cardJson.CardFaces[0].PrintedName, cardJson.CardFaces[1].PrintedName)
	}

	card.SetName(printedName, cardJson.Lang)

	if !card.IsValid() {
		panic(fmt.Sprintf("Card is not valid: %v", card))
	}
}

// setJson must be a copy (not a pointer) cause this method is called in a go routine
func (importer *Importer) downloadSetIcon(setJson setJsonStruct) {
	defer pushSemaphoreAndDefer(&importer.wg, importer.downloaderSemaphore)()

	iconName := setJson.getIconName()
	svgFilePath := filepath.Join(SetImagesDir(importer.ImagesDir), fmt.Sprintf("%s.svg", iconName))
	setIconFilePath := SetImagePath(importer.ImagesDir, iconName)
	if _, err := os.Stat(setIconFilePath); importer.ForceDownloadAssets || os.IsNotExist(err) {
		err := downloadFile(svgFilePath, setJson.IconSvgUri)
		if err != nil {
			importer.errorsChan <- err
			return
		}

		err = runCmd("rsvg-convert", svgFilePath, "-b", "white", "-o", setIconFilePath)
		if err != nil {
			importer.errorsChan <- err
		}
	}
}

// cardJson must be a copy (not a pointer) cause this method is called in a go routine
func (importer *Importer) downloadCardImage(cardJson cardJsonStruct, saveAsLang string) {
	defer pushSemaphoreAndDefer(&importer.wg, importer.downloaderSemaphore)()

	if hasBackSide(&cardJson) {
		imageUrl := cardJson.CardFaces[0].ImageUris.GetImageByTypeName(importer.ImageType)
		filePath := CardImagePath(importer.ImagesDir, cardJson.SetCode, cardJson.CollectorNumber, saveAsLang, false)
		importer.downloadImage(imageUrl, filePath)
		imageUrl = cardJson.CardFaces[1].ImageUris.GetImageByTypeName(importer.ImageType)
		filePath = CardImagePath(importer.ImagesDir, cardJson.SetCode, cardJson.CollectorNumber, saveAsLang, true)
		importer.downloadImage(imageUrl, filePath)
	} else {
		imageUrl := cardJson.ImageUris.GetImageByTypeName(importer.ImageType)
		filePath := CardImagePath(importer.ImagesDir, cardJson.SetCode, cardJson.CollectorNumber, saveAsLang, false)
		importer.downloadImage(imageUrl, filePath)
	}
	importer.bar.Increment()
}

func (importer *Importer) downloadImage(imageUrl, filePath string) {
	var downloadErr error
	if importer.ForceDownloadAssets {
		downloadErr = downloadFile(filePath, imageUrl)
		if downloadErr != nil {
			importer.errorsChan <- downloadErr
		}
		return
	}

	stat, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		downloadErr = downloadFile(filePath, imageUrl)
		if downloadErr != nil {
			importer.errorsChan <- downloadErr
		}
		return
	}

	if importer.ForceDownloadOlderAssets || importer.ForceDownloadDiffSha1 {
		if !importer.ForceDownloadOlderAssets {
			stat = nil
		}
		sha1 := ""
		if importer.ForceDownloadDiffSha1 {
			sha1 = sha1sum(filePath)
		}
		downloadErr = downloadFileWhenChanged(filePath, imageUrl, stat, sha1)
		if downloadErr != nil {
			importer.errorsChan <- downloadErr
		}
	}
}

func hasBackSide(cardJson *cardJsonStruct) bool {
	return len(cardJson.CardFaces) > 1 && cardJson.CardFaces[0].ImageUris != (imagesCardJsonStruct{}) && cardJson.CardFaces[1].ImageUris != (imagesCardJsonStruct{})
}

func getResponseHeader(url string) (http.Header, error) {
	var (
		resp *http.Response
		err  error
	)
	retryErr := retryOnError(3, 100*time.Millisecond, func() error {
		resp, err = http.Head(url)
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			return httpError(url, resp.StatusCode)
		}
		return nil
	})
	return resp.Header, retryErr
}

func downloadFileWhenChanged(filepath, url string, stat os.FileInfo, sha1 string) error {
	header, err := getResponseHeader(url)
	if err != nil {
		return err
	}

	reDownloadReasons := []string{}
	if stat != nil && remoteFileIsNewer(header, stat.ModTime()) {
		reDownloadReasons = append(reDownloadReasons, "remote timestamp is newer")
	}

	if sha1 != "" && remoteFileHasDiffSha1(header, sha1) {
		reDownloadReasons = append(reDownloadReasons, "remote sha1 is changed")
	}

	if len(reDownloadReasons) == 0 {
		return nil
	}

	log.Printf("Force re-download of image file '%s': %s\n", filepath, strings.Join(reDownloadReasons, " and "))
	return downloadFile(filepath, url)
}

func downloadFile(filepath, url string) error {
	return retryOnError(3, 100*time.Millisecond, func() error {
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return httpError(url, resp.StatusCode)
		}

		file, err := os.Create(filepath)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(file, resp.Body)
		return err
	})
}

func remoteFileIsNewer(header http.Header, fileModTime time.Time) bool {
	lastModified := header.Get("last-modified")
	remoteLastModified, err := time.Parse(time.RFC1123, lastModified)
	if err == nil {
		return remoteLastModified.After(fileModTime)
	}

	uploadTimestamp := header.Get("x-bz-upload-timestamp")
	timestamp, err := strconv.ParseInt(uploadTimestamp, 10, 64)
	if err == nil {
		return time.Unix(timestamp/1000.0, 0).After(fileModTime)
	}
	return true
}

func remoteFileHasDiffSha1(header http.Header, sha1 string) bool {
	remoteSha1 := header.Get("x-bz-content-sha1")
	remoteSha1 = strings.ReplaceAll(remoteSha1, "unverified:", "")
	return remoteSha1 != sha1
}

func sha1sum(filepath string) string {
	f, err := os.Open(filepath)
	if err != nil {
		return ""
	}
	defer f.Close()
	h := sha1.New()
	if _, err = io.Copy(h, f); err != nil {
		return ""
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func httpError(url string, statusCode int) error {
	return fmt.Errorf("Download file `%s` failed with status code %d", url, statusCode)
}

func runCmd(arg string, args ...string) error {
	var stderr bytes.Buffer
	cmd := exec.Command(arg, args...)
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Command `%s` fail: %s\n%s", cmd.String(), err, stderr.String())
	}
	return nil
}

func loadFile(filePath string, out interface{}) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, out)
	if err != nil {
		return err
	}

	return nil
}

// Utils

func contains(collection []string, s string) bool {
	for _, ss := range collection {
		if ss == s {
			return true
		}
	}
	return false
}

func removeAllFilesByExtension(dirPath, ext string) {
	files, err := filepath.Glob(filepath.Join(dirPath, fmt.Sprintf("*.%s", ext)))
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		err = os.Remove(file)
		if err != nil {
			panic(err)
		}
	}
}

func createDirIfNotExist(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
}

func retryOnError(attempts int, delay time.Duration, f func() error) error {
	var err error
	retryCount := 0
	for {
		if retryCount == attempts {
			break
		}
		err = f()
		if err == nil {
			break
		}
		retryCount += 1
		log.Printf("[Retry] Action failed (attempt #%d): %s\n", retryCount, err)
		time.Sleep(delay)
	}
	return err
}

func pushSemaphoreAndDefer(wg *sync.WaitGroup, semaphore chan struct{}) func() {
	semaphore <- struct{}{}
	return func() {
		wg.Done()
		<-semaphore
	}
}

func waitErrors(wg *sync.WaitGroup, readChannel chan error) {
	quit := make(chan bool)

	go func() {
		wg.Wait()
		quit <- true
	}()

	for {
		select {
		case err := <-readChannel:
			log.Println(err)
		case <-quit:
			return
		}
	}
}
