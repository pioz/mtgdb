package mtgdb_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pioz/mtgdb"
	"github.com/stretchr/testify/assert"
)

type cardJsonStruct struct {
	Name string `json:"name"`
}

var expectedCardNames = []string{
	"Acclaimed Contender",
	"Acclaimed Contender",
	"Acclaimed Contender",
	"Acclaimed Contender",
	"Garruk, Cursed Huntsman",
	"Acclaimed Contender",
	"Acclaimed Contender",
	"Acclaimed Contender",
	"Acclaimed Contender",
	"Acclaimed Contender",
	"Acclaimed Contender",
	"Acclaimed Contender",
	"Acclaimed Contender",
	"Acclaimed Contender",
	"Acclaimed Contender",
	"Garruk, Cursed Huntsman Emblem",
	"\"Rumors of My Death . . .\"",
	"Daybreak Ranger // Nightfall Predator",
	"Daybreak Ranger // Nightfall Predator",
	"Daybreak Ranger // Nightfall Predator",
	"Daybreak Ranger // Nightfall Predator",
	"Daybreak Ranger // Nightfall Predator",
	"Daybreak Ranger // Nightfall Predator",
	"Daybreak Ranger // Nightfall Predator",
	"Daybreak Ranger // Nightfall Predator",
	"Daybreak Ranger // Nightfall Predator",
	"Daybreak Ranger // Nightfall Predator",
	"Daybreak Ranger // Nightfall Predator",
	"Garruk, Cursed Huntsman",
	"Nissa, Who Shakes the World",
	"Birds of Paradise // Birds of Paradise",
}

func TestJsonStreamerJsonl(t *testing.T) {
	streamer, err := mtgdb.NewJsonStreamer(filepath.Join("./testdata", "data", "all_cards.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	var cardJson cardJsonStruct
	i := 0
	for streamer.Next() {
		err := streamer.Get(&cardJson)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, expectedCardNames[i], cardJson.Name, i)
		i++
	}
	assert.Equal(t, len(expectedCardNames), i)
}

func TestJsonStreamerArray(t *testing.T) {
	names := []string{"Acclaimed Contender", "Garruk, Cursed Huntsman", "Daybreak Ranger // Nightfall Predator"}
	arrayFilePath := filepath.Join(t.TempDir(), "cards.json")
	err := os.WriteFile(arrayFilePath, []byte(`[
		{"name": "Acclaimed Contender"},
		{"name": "Garruk, Cursed Huntsman"},
		{"name": "Daybreak Ranger // Nightfall Predator"}
	]`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	streamer, err := mtgdb.NewJsonStreamer(arrayFilePath)
	if err != nil {
		t.Fatal(err)
	}
	var cardJson cardJsonStruct
	i := 0
	for streamer.Next() {
		err := streamer.Get(&cardJson)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, names[i], cardJson.Name, i)
		i++
	}
	assert.Equal(t, len(names), i)
}

func TestJsonStreamerObject(t *testing.T) {
	streamer, err := mtgdb.NewJsonStreamer(filepath.Join("./testdata", "data", "stream_me.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cardJson cardJsonStruct
	i := 0
	for streamer.Next() {
		err := streamer.Get(&cardJson)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "Daybreak Ranger // Nightfall Predator", cardJson.Name)
		i++
	}
	assert.Equal(t, 1, i)
}

func TestJsonStreamerFileNotFound(t *testing.T) {
	_, err := mtgdb.NewJsonStreamer(filepath.Join("./testdata", "data", "not_exist.json"))
	assert.Error(t, err)
}
