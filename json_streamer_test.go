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

func TestJsonStreamer(t *testing.T) {
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
	streamer, err := mtgdb.NewJsonStreamer(filepath.Join("./fixtures", "data", "all_cards.json"))
	if err != nil {
		panic(err)
	}
	var cardJson cardJsonStruct
	i := 0
	for streamer.Next() {
		streamer.Get(&cardJson)
		assert.Equal(t, names[i], cardJson.Name, i)
		i++
	}
}
