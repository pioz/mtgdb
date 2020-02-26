package mtgdb

import (
	"encoding/json"
	"errors"
	"os"
)

type Token interface{}

type JsonStreamer struct {
	*json.Decoder
	file *os.File
}

func NewJsonStreamer(filepath string) (*JsonStreamer, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(file)
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	if delim, ok := token.(json.Delim); !ok || delim != '[' {
		return nil, errors.New("json is not an array")
	}
	return &JsonStreamer{Decoder: decoder, file: file}, nil
}

func (streamer *JsonStreamer) Next() bool {
	more := streamer.Decoder.More()
	if !more {
		_, err := streamer.Token()
		if err != nil {
			panic(err)
		}
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
