package mtgdb_test

import (
	"path/filepath"
	"testing"

	"github.com/pioz/mtgdb"
	"github.com/stretchr/testify/assert"
)

type cardJsonStruct struct {
	Name string `json:"name"`
}

func TestJsonStreamerArray(t *testing.T) {
	names := []string{
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
	}
	streamer, _ := mtgdb.NewJsonStreamer(filepath.Join("./testdata", "data", "all_cards.json"))
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
}

func TestJsonStreamerObject(t *testing.T) {
	_, err := mtgdb.NewJsonStreamer(filepath.Join("./fixtures", "data", "stream_me.json"))
	assert.Error(t, err, "json is not an array")
}
