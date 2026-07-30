package mtgdb

import (
	"bufio"
	"encoding/json"
	"os"
)

type Token interface{}

type JsonStreamer struct {
	*json.Decoder
	file *os.File
}

// NewJsonStreamer creates a streamer over a file containing a stream of
// whitespace-separated JSON values, like the JSONL files (one JSON value per
// line) used by Scryfall bulk data. Each call to Get decodes the next
// top-level value in the stream.
func NewJsonStreamer(filepath string) (*JsonStreamer, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReaderSize(file, 1024*1024)
	return &JsonStreamer{Decoder: json.NewDecoder(reader), file: file}, nil
}

func (streamer *JsonStreamer) Next() bool {
	more := streamer.More()
	if !more {
		streamer.file.Close()
	}
	return more
}

func (streamer *JsonStreamer) Get(out interface{}) error {
	err := streamer.Decode(out)
	if err != nil {
		return err
	}
	return nil
}
