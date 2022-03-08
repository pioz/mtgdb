package mtgdb

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pioz/mtgdb/pb"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	DisplayProgressBar       bool

	cardCollection        map[string]*Card
	setCollection         map[string]*Set
	rulingsCollection     map[string]Rulings
	setIconsDownloaded    map[string]struct{}
	notEnImagesToDownload map[string]*cardJsonStruct
	downloadedImagesCount uint32

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
		DisplayProgressBar:       false,
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
	rulingsJsonFilePath := filepath.Join(importer.DataDir, "rulings.json")
	if _, err := os.Stat(allSetsJsonFilePath); importer.ForceDownloadData || os.IsNotExist(err) {
		err := downloadFile(allSetsJsonFilePath, "https://api.scryfall.com/sets")
		if err != nil {
			return err
		}
	}
	if _, err := os.Stat(allCardsJsonFilePath); importer.ForceDownloadData || os.IsNotExist(err) {
		urls, err := fetchAllCardsDataUrl()
		if err != nil {
			return err
		}
		err = downloadFile(allCardsJsonFilePath, urls["all_cards"])
		if err != nil {
			return err
		}
		err = downloadFile(rulingsJsonFilePath, urls["rulings"])
		if err != nil {
			return err
		}
	}
	return nil
}

func (importer *Importer) BuildCardsFromJson() ([]Card, uint32) {
	defer removeAllFilesByExtension(SetImagesDir(importer.ImagesDir), "svg")

	importer.downloadedImagesCount = 0
	importer.cardCollection = make(map[string]*Card)
	importer.setCollection = make(map[string]*Set)
	importer.rulingsCollection = make(map[string]Rulings)
	if importer.DownloadAssets {
		createDirIfNotExist(SetImagesDir(importer.ImagesDir))
		importer.setIconsDownloaded = make(map[string]struct{})
		importer.notEnImagesToDownload = make(map[string]*cardJsonStruct)
	}

	// Fill importer.rulingsCollection
	rulingsJson := make([]rulingsJsonStruct, 0)
	err := loadFile(filepath.Join(importer.DataDir, "rulings.json"), &rulingsJson)
	if err != nil {
		panic(err)
	}
	for _, rulingJson := range rulingsJson {
		importer.buildRuling(&rulingJson)
	}

	// Fill importer.setCollection
	setsJson := setsJsonStruct{}
	err = loadFile(filepath.Join(importer.DataDir, "all_sets.json"), &setsJson)
	if err != nil {
		panic(err)
	}
	for _, setJson := range setsJson.Data {
		if len(importer.OnlyTheseSetCodes) != 0 && !contains(importer.OnlyTheseSetCodes, setJson.Code) {
			continue
		}
		importer.buildSet(&setJson)
	}

	if importer.DownloadAssets && importer.DisplayProgressBar {
		importer.bar = pb.New("Download images", 0)
	}

	// Fill importer.cardCollection
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
			if importer.bar != nil {
				importer.bar.IncrementMax()
			}
			importer.wg.Add(1)
			go importer.downloadCardImage(*cardJson, "en")
		}
		if importer.bar != nil {
			importer.bar.Finishln()
		}
	}

	waitErrors(&importer.wg, importer.errorsChan)
	close(importer.errorsChan)
	cards := make([]Card, 0, len(importer.cardCollection))
	for _, card := range importer.cardCollection {
		cards = append(cards, *card)
	}
	return cards, importer.downloadedImagesCount
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

	scope := db.Clauses(clause.OnConflict{UpdateAll: true}).Session(&gorm.Session{CreateBatchSize: 1000})
	err := scope.Create(allSets).Error
	if err != nil {
		return err
	}
	return scope.Omit("Set").Create(cards).Error
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
	SetType       string `json:"set_type"`
}

type rulingsJsonStruct struct {
	OracleId    string `json:"oracle_id"`
	PublishedAt string `json:"published_at"`
	Comment     string `json:"comment"`
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
	Id              string                 `json:"id"`
	OracleId        string                 `json:"oracle_id"`
	Name            string                 `json:"name"`
	PrintedName     string                 `json:"printed_name"`
	Lang            string                 `json:"lang"`
	ReleasedAt      string                 `json:"released_at"`
	ImageUris       imagesCardJsonStruct   `json:"image_uris"`
	CardFaces       []cardFaceStruct       `json:"card_faces"`
	SetCode         string                 `json:"set"`
	SetName         string                 `json:"set_name"`
	SetType         string                 `json:"set_type"`
	CollectorNumber string                 `json:"collector_number"`
	Foil            bool                   `json:"foil"`
	NonFoil         bool                   `json:"nonfoil"`
	MtgoID          uint64                 `json:"mtgo_id"`
	ArenaID         uint64                 `json:"arena_id"`
	TcgplayerID     uint64                 `json:"tcgplayer_id"`
	CardmarketID    uint64                 `json:"cardmarket_id"`
	Layout          string                 `json:"layout"`
	ManaCost        string                 `json:"mana_cost"`
	CMC             float32                `json:"cmc"`
	TypeLine        string                 `json:"type_line"`
	OracleText      string                 `json:"oracle_text"`
	Power           string                 `json:"power"`
	Toughness       string                 `json:"toughness"`
	Colors          []string               `json:"colors"`
	ColorIdentity   []string               `json:"color_identity"`
	Keywords        []string               `json:"keywords"`
	ProducedMana    []string               `json:"produced_mana"`
	Legalities      map[string]interface{} `json:"legalities"`
	Games           []string               `json:"games"`
	Oversized       bool                   `json:"oversized"`
	Promo           bool                   `json:"promo"`
	Reprint         bool                   `json:"reprint"`
	Variation       bool                   `json:"variation"`
	Digital         bool                   `json:"digital"`
	Rarity          string                 `json:"rarity"`
	Watermark       string                 `json:"watermark"`
	Artist          string                 `json:"artist"`
	BorderColor     string                 `json:"border_color"`
	Frame           string                 `json:"frame"`
	FrameEffects    []string               `json:"frame_effects"`
	SecurityStamp   string                 `json:"security_stamp"`
	FullArt         bool                   `json:"full_art"`
	Textless        bool                   `json:"textless"`
	Booster         bool                   `json:"booster"`
	StorySpotlight  bool                   `json:"story_spotlight"`
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

func fetchAllCardsDataUrl() (map[string]string, error) {
	resp, err := http.Get("https://api.scryfall.com/bulk-data")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var bulkDataArray bulkDataArrayJsonStruct
	err = json.Unmarshal(body, &bulkDataArray)
	if err != nil {
		return nil, err
	}
	urls := make(map[string]string)
	for _, bulkData := range bulkDataArray.Data {
		urls[bulkData.Type] = bulkData.DownloadUri
	}
	return urls, nil
}

func (importer *Importer) buildRuling(rulingJson *rulingsJsonStruct) {
	publishedAt, err := time.Parse("2006-01-02 15:04:05", rulingJson.PublishedAt+" 00:00:00")
	if err != nil {
		return
	}
	ruling := Ruling{PublishedAt: publishedAt, Comment: rulingJson.Comment}
	if _, found := importer.rulingsCollection[rulingJson.OracleId]; !found {
		importer.rulingsCollection[rulingJson.OracleId] = make(Rulings, 0)
	}
	importer.rulingsCollection[rulingJson.OracleId] = append(importer.rulingsCollection[rulingJson.OracleId], ruling)
}

func (importer *Importer) buildSet(setJson *setJsonStruct) {
	iconName := setJson.getIconName()
	if _, found := importer.setCollection[setJson.Code]; !found {
		set := &Set{
			Name:       setJson.Name,
			Code:       setJson.Code,
			ParentCode: setJson.getParentCode(),
			Typology:   setJson.SetType,
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
	key := fmt.Sprintf("%s-%s", cardJson.SetCode, cardJson.CollectorNumber)
	card, found := importer.cardCollection[key]
	if !found {
		card = &Card{
			EnName:          cardJson.Name,
			SetCode:         cardJson.SetCode,
			CollectorNumber: cardJson.CollectorNumber,
			HasBackSide:     hasBackSide(cardJson),
			Set:             importer.setCollection[cardJson.SetCode],
			Foil:            cardJson.Foil,
			NonFoil:         cardJson.NonFoil,
			ScryfallId:      cardJson.Id,
			OracleId:        cardJson.OracleId,
			MtgoID:          cardJson.MtgoID,
			ArenaID:         cardJson.ArenaID,
			TcgplayerID:     cardJson.TcgplayerID,
			CardmarketID:    cardJson.CardmarketID,
			Layout:          cardJson.Layout,
			ManaCost:        cardJson.ManaCost,
			CMC:             cardJson.CMC,
			TypeLine:        cardJson.TypeLine,
			OracleText:      cardJson.OracleText,
			Power:           cardJson.Power,
			Toughness:       cardJson.Toughness,
			Colors:          cardJson.Colors,
			ColorIdentity:   cardJson.ColorIdentity,
			Keywords:        cardJson.Keywords,
			ProducedMana:    cardJson.ProducedMana,
			Legalities:      cardJson.Legalities,
			Games:           cardJson.Games,
			Oversized:       cardJson.Oversized,
			Promo:           cardJson.Promo,
			Reprint:         cardJson.Reprint,
			Variation:       cardJson.Variation,
			Digital:         cardJson.Digital,
			Rarity:          cardJson.Rarity,
			Watermark:       cardJson.Watermark,
			Artist:          cardJson.Artist,
			BorderColor:     cardJson.BorderColor,
			Frame:           cardJson.Frame,
			FrameEffects:    cardJson.FrameEffects,
			SecurityStamp:   cardJson.SecurityStamp,
			FullArt:         cardJson.FullArt,
			Textless:        cardJson.Textless,
			Booster:         cardJson.Booster,
			StorySpotlight:  cardJson.StorySpotlight,
			Rulings:         importer.rulingsCollection[cardJson.OracleId],
		}
		importer.cardCollection[key] = card
		if importer.DownloadAssets {
			importer.notEnImagesToDownload[key] = cardJson
		}
	} else if cardJson.Lang == "en" {
		card.ScryfallId = cardJson.Id
	}
	if importer.DownloadAssets && (!importer.DownloadOnlyEnAssets || cardJson.Lang == "en") {
		if importer.bar != nil {
			importer.bar.IncrementMax()
		}
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
		if imageUrl != "" {
			filePath := CardImagePath(importer.ImagesDir, cardJson.SetCode, cardJson.CollectorNumber, saveAsLang, false)
			importer.downloadImage(imageUrl, filePath)
		}
		imageUrl = cardJson.CardFaces[1].ImageUris.GetImageByTypeName(importer.ImageType)
		if imageUrl != "" {
			filePath := CardImagePath(importer.ImagesDir, cardJson.SetCode, cardJson.CollectorNumber, saveAsLang, true)
			importer.downloadImage(imageUrl, filePath)
		}
	} else {
		imageUrl := cardJson.ImageUris.GetImageByTypeName(importer.ImageType)
		if imageUrl != "" {
			filePath := CardImagePath(importer.ImagesDir, cardJson.SetCode, cardJson.CollectorNumber, saveAsLang, false)
			importer.downloadImage(imageUrl, filePath)
		}
	}
	if importer.bar != nil {
		importer.bar.Increment()
	}
}

func (importer *Importer) downloadImage(imageUrl, filePath string) {
	var downloadErr error
	if importer.ForceDownloadAssets {
		downloadErr = downloadFile(filePath, imageUrl)
		if downloadErr != nil {
			importer.errorsChan <- downloadErr
		} else {
			atomic.AddUint32(&importer.downloadedImagesCount, 1)
		}
		return
	}

	stat, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		downloadErr = downloadFile(filePath, imageUrl)
		if downloadErr != nil {
			importer.errorsChan <- downloadErr
		} else {
			atomic.AddUint32(&importer.downloadedImagesCount, 1)
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
		downloaded, downloadErr := downloadFileWhenChanged(filePath, imageUrl, stat, sha1)
		if downloadErr != nil {
			importer.errorsChan <- downloadErr
		} else if downloaded {
			atomic.AddUint32(&importer.downloadedImagesCount, 1)
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

func downloadFileWhenChanged(filepath, url string, stat os.FileInfo, sha1 string) (bool, error) {
	header, err := getResponseHeader(url)
	if err != nil {
		return false, err
	}

	reDownloadReasons := []string{}
	if stat != nil && remoteFileIsNewer(header, stat.ModTime()) {
		reDownloadReasons = append(reDownloadReasons, "remote timestamp is newer")
	}

	if sha1 != "" && remoteFileHasDiffSha1(header, sha1) {
		reDownloadReasons = append(reDownloadReasons, "remote sha1 is changed")
	}

	if len(reDownloadReasons) == 0 {
		return false, nil
	}

	log.Printf("Force re-download of image file '%s': %s\n", filepath, strings.Join(reDownloadReasons, " and "))
	return true, downloadFile(filepath, url)
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
	return fmt.Errorf("download file `%s` failed with status code %d", url, statusCode)
}

func runCmd(arg string, args ...string) error {
	var stderr bytes.Buffer
	cmd := exec.Command(arg, args...)
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("command `%s` fail: %s\n%s", cmd.String(), err, stderr.String())
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
