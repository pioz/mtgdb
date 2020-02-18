package mtgdb

import (
	"bytes"
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
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pioz/mtgdb/pb"
)

type Importer struct {
	DataDir             string
	ImagesDir           string
	OnlyTheseSetCodes   []string
	ForceDownloadData   bool
	DownloadAssets      bool
	ForceDownloadAssets bool
	ImageType           string

	errorsChan          chan error
	wg                  sync.WaitGroup
	downloaderSemaphore chan struct{}
	bar                 *pb.ProgressBar
}

func NewImporter(dataDir string) *Importer {
	return &Importer{
		DataDir:             dataDir,
		OnlyTheseSetCodes:   []string{},
		ImagesDir:           filepath.Join(dataDir, "images"),
		ForceDownloadData:   false,
		DownloadAssets:      true,
		ForceDownloadAssets: false,
		ImageType:           "normal",
		errorsChan:          make(chan error, 10),
		downloaderSemaphore: make(chan struct{}, 50),
	}
}

func (self *Importer) DownloadData() error {
	createDirIfNotExist(self.DataDir)

	allSetsJsonFilePath := filepath.Join(self.DataDir, "all_sets.json")
	allCardsJsonFilePath := filepath.Join(self.DataDir, "all_cards.json")
	if _, err := os.Stat(allSetsJsonFilePath); self.ForceDownloadData || os.IsNotExist(err) {
		err := downloadFile(allSetsJsonFilePath, "https://api.scryfall.com/sets")
		if err != nil {
			return err
		}
	}
	if _, err := os.Stat(allCardsJsonFilePath); self.ForceDownloadData || os.IsNotExist(err) {
		err = downloadFile(allCardsJsonFilePath, "https://archive.scryfall.com/json/scryfall-all-cards.json")
		if err != nil {
			return err
		}
	}
	return nil
}

func (self *Importer) BuildCardsFromJson() []Card {
	defer removeAllFilesByExtension(SetImagesDir(self.ImagesDir), "svg")
	if self.DownloadAssets {
		createDirIfNotExist(SetImagesDir(self.ImagesDir))
	}

	setsJson := setsJsonStruct{}
	err := loadFile(filepath.Join(self.DataDir, "all_sets.json"), &setsJson)
	if err != nil {
		panic(err)
	}

	setCodeToIconNameMap := self.getIconNamesAndDownloadSetIcons(&setsJson)

	allCardsJson := make([]cardJsonStruct, 0, 60000)
	err = loadFile(filepath.Join(self.DataDir, "all_cards.json"), &allCardsJson)
	if err != nil {
		panic(err)
	}

	if self.DownloadAssets {
		self.bar = pb.New("Download images", 0)
	}
	collection := make(map[string]*Card)
	for _, cardJson := range allCardsJson {
		if len(self.OnlyTheseSetCodes) != 0 && !contains(self.OnlyTheseSetCodes, cardJson.SetCode) {
			continue
		}
		self.buildCardAndDownloadCardImage(collection, &cardJson, setCodeToIconNameMap[cardJson.SetCode])
	}

	waitErrors(&self.wg, self.errorsChan)
	close(self.errorsChan)
	if self.DownloadAssets {
		self.bar.Finishln()
	}
	cards := make([]Card, 0, len(collection))
	for _, card := range collection {
		cards = append(cards, *card)
	}
	return cards
}

func AutoMigrate(db *gorm.DB) {
	db.AutoMigrate(&Set{})
	db.Model(&Set{}).AddUniqueIndex("idx_sets_code", "code")
	db.AutoMigrate(&Card{})
	db.Model(&Card{}).AddUniqueIndex("idx_cards_set_code_collector_number_is_token", "set_code", "collector_number", "is_token")
	db.Model(&Card{}).AddIndex("idx_cards_en_name", "en_name")
	db.Model(&Card{}).AddForeignKey("set_code", "sets(code)", "RESTRICT", "RESTRICT")
}

func BulkInsert(db *gorm.DB, cards []Card) error {
	sets := make(map[string]Set)
	for _, card := range cards {
		if _, found := sets[card.SetCode]; !found && card.SetCode != "" {
			sets[card.SetCode] = card.Set
		}
	}
	allSets := make([]Set, 0, len(sets))
	for _, set := range sets {
		allSets = append(allSets, set)
	}
	err := bulkInsert(db, Set{}, allSets, 1000)
	if err != nil {
		return err
	}
	return bulkInsert(db, Card{}, cards, 1000)
}

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
		errors := db.Exec(query, allValues...).GetErrors()
		if len(errors) > 0 {
			return errors[0]
		}
	}
	return nil
}

// PRIVATE types

type setsJsonStruct struct {
	Data []setJsonStruct `json:"data"`
}

type setJsonStruct struct {
	Code          string `json:"code"`
	IconSvgUri    string `json:"icon_svg_uri"`
	ParentSetCode string `json:"parent_set_code"`
}

func (self *setJsonStruct) getIconName() string {
	basename := filepath.Base(self.IconSvgUri)
	return strings.Split(basename, ".")[0]
}

func (self *setJsonStruct) getParentCode() (code string) {
	if self.ParentSetCode != "" {
		code = self.ParentSetCode
	} else {
		code = self.Code
	}
	return strings.ToLower(code)
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
}

type imagesCardJsonStruct struct {
	Png    string `json:"png"`
	Large  string `json:"large"`
	Normal string `json:"normal"`
	Small  string `json:"small"`
}

func (self *imagesCardJsonStruct) GetImageByTypeName(name string) string {
	switch name {
	case "png":
		return self.Png
	case "large":
		return self.Large
	case "normal":
		return self.Normal
	case "small":
		return self.Small
	default:
		return self.Normal
	}
}

type cardFaceStruct struct {
	PrintedName string               `json:"printed_name"`
	ImageUris   imagesCardJsonStruct `json:"image_uris"`
}

// PRIVATE methods

func (importer *Importer) getIconNamesAndDownloadSetIcons(setsJson *setsJsonStruct) map[string]string {
	setCodeToIconNameMap := make(map[string]string)
	iconNameDone := make(map[string]struct{})
	for _, setJson := range setsJson.Data {
		iconName := setJson.getIconName()
		setCodeToIconNameMap[setJson.Code] = iconName
		if importer.DownloadAssets {
			createDirIfNotExist(filepath.Join(CardImagesDir(importer.ImagesDir), setJson.Code))
			if _, found := iconNameDone[iconName]; !found {
				iconNameDone[iconName] = struct{}{}
				importer.wg.Add(1)
				go importer.downloadSetIcon(setJson)
			}
		}
	}
	return setCodeToIconNameMap
}

func (importer *Importer) buildCardAndDownloadCardImage(collection map[string]*Card, cardJson *cardJsonStruct, iconName string) {
	isToken := false
	if cardJson.SetType == "token" {
		isToken = true
	}
	key := fmt.Sprintf("%s-%s", cardJson.SetCode, cardJson.CollectorNumber)
	card, found := collection[key]
	if !found {
		card = &Card{
			ScryfallId:      cardJson.Id,
			EnName:          cardJson.Name,
			SetCode:         cardJson.SetCode,
			CollectorNumber: cardJson.CollectorNumber,
			IsToken:         isToken,
			Set: Set{
				Name:     cardJson.SetName,
				Code:     cardJson.SetCode,
				IconName: iconName,
			},
		}
		releasedAt, err := time.Parse("2006-01-02", cardJson.ReleasedAt)
		if err == nil {
			card.Set.ReleasedAt = &releasedAt
		}
		collection[key] = card
		if importer.DownloadAssets {
			importer.wg.Add(1)
			importer.bar.IncrementMax()
			go importer.downloadCardImage(*cardJson)
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
func (importer *Importer) downloadCardImage(cardJson cardJsonStruct) {
	defer pushSemaphoreAndDefer(&importer.wg, importer.downloaderSemaphore)()

	importer.bar.Increment()
	imageUrl := cardJson.ImageUris.GetImageByTypeName(importer.ImageType)
	if imageUrl == "" && len(cardJson.CardFaces) > 0 {
		imageUrl = cardJson.CardFaces[0].ImageUris.GetImageByTypeName(importer.ImageType)
	}
	filePath := CardImagePath(importer.ImagesDir, cardJson.SetCode, cardJson.CollectorNumber)
	if _, err := os.Stat(filePath); importer.ForceDownloadAssets || os.IsNotExist(err) {
		err = downloadFile(filePath, imageUrl)
		if err != nil {
			importer.errorsChan <- err
		}
	}
}

func downloadFile(filepath, url string) error {
	return retryOnError(3, 100*time.Millisecond, func() error {
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return errors.New(fmt.Sprintf("Download file `%s` filed with status code %d", url, resp.StatusCode))
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

func runCmd(arg string, args ...string) error {
	var stderr bytes.Buffer
	cmd := exec.Command(arg, args...)
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return errors.New(fmt.Sprintf("Command `%s` fail: %s\n%s", cmd.String(), err, stderr.String()))
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

func removeAllFilesByExtension(dirPath, ext string) error {
	files, err := filepath.Glob(filepath.Join(dirPath, fmt.Sprintf("*.%s", ext)))
	if err != nil {
		return err
	}
	for _, file := range files {
		err = os.Remove(file)
		if err != nil {
			return err
		}
	}
	return nil
}

func createDirIfNotExist(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, os.ModePerm)
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
