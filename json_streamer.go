package mtgdb

import (
	"encoding/json"
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
	// defer file.Close()
	decoder := json.NewDecoder(file)
	_, err = decoder.Token()
	if err != nil {
		return nil, err
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
