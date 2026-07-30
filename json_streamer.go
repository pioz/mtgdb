package mtgdb

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
)

type Token interface{}

type JsonStreamer struct {
	*json.Decoder
	file    *os.File
	isArray bool
}

// NewJsonStreamer creates a streamer over a file containing either a JSONL
// stream (one JSON value per line, the format used by Scryfall bulk data
// files) or a single JSON array (the legacy Scryfall bulk data format).
func NewJsonStreamer(filepath string) (*JsonStreamer, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReaderSize(file, 1024*1024)
	firstByte, err := firstNonSpaceByte(reader)
	if err != nil && err != io.EOF {
		file.Close()
		return nil, err
	}
	streamer := &JsonStreamer{Decoder: json.NewDecoder(reader), file: file, isArray: firstByte == '['}
	if streamer.isArray {
		_, err := streamer.Token()
		if err != nil {
			file.Close()
			return nil, err
		}
	}
	return streamer, nil
}

func (streamer *JsonStreamer) Next() bool {
	more := streamer.More()
	if !more {
		if streamer.isArray {
			_, err := streamer.Token()
			if err != nil {
				panic(err)
			}
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

func firstNonSpaceByte(reader *bufio.Reader) (byte, error) {
	for i := 1; ; i++ {
		buf, err := reader.Peek(i)
		if len(buf) < i {
			return 0, err
		}
		c := buf[i-1]
		if c != ' ' && c != '\t' && c != '\r' && c != '\n' {
			return c, nil
		}
	}
}
