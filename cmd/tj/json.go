package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
)

const jsonStreamScratchBufferBytes = 4096

type jsonStream struct {
	token
	Text   string      `json:"-"` // the original text that Object was parsed from
	Object interface{} `json:"object,omitempty"`

	textBuffer    *bytes.Buffer // intercepts bytes read by decoder
	scratchBuffer []byte        // determines size of decoder.Buffered()
	buffer        *tokenBuffer
	decoder       *json.Decoder
	decodeError   error
	done          bool
}

func newJSONStream() *jsonStream {
	textBuffer := bytes.NewBuffer(nil)
	tee := io.TeeReader(os.Stdin, textBuffer)
	return &jsonStream{
		decoder:       json.NewDecoder(tee),
		textBuffer:    textBuffer,
		scratchBuffer: make([]byte, jsonStreamScratchBufferBytes),
		buffer:        &tokenBuffer{},
	}
}

func (j *jsonStream) Token() *token {
	return &j.token
}

func (j *jsonStream) CopyCurrent() tokenStream {
	return &jsonStream{
		token:  j.token,
		Object: j.Object,
	}
}

func (j *jsonStream) AppendCurrentToBuffer() {
	*j.buffer = append(*j.buffer, j.CopyCurrent())
}

func (j *jsonStream) FlushBuffer() {
	j.buffer.flush(j)
}

func (j *jsonStream) CurrentMatchText() string {
	if matchTemplate != nil {
		return matchTemplate.execute(j.Object)
	}
	return j.Text
}

func (j *jsonStream) Err() error {
	if j.decodeError == io.EOF {
		return nil
	}
	return j.decodeError
}

func (j *jsonStream) readerSize(r io.Reader) int {
	total := 0
	var err error
	var n int
	for err == nil {
		n, err = r.Read(j.scratchBuffer)
		total += n
	}
	return total
}

func (j *jsonStream) Scan() bool {
	j.Object = new(interface{})
	err := j.decoder.Decode(&j.Object)
	numBytesNotParsedByJSON := j.readerSize(j.decoder.Buffered()) // "{..} XYZ" -> len("XYZ")
	bytesUnreadByUs := j.textBuffer.Bytes()                       // "{..} XYZ" -> "{..} XYZ"
	numBytesUnreadByUs := len(bytesUnreadByUs)
	numBytesParsedByJSON := numBytesUnreadByUs - numBytesNotParsedByJSON // len("{..}")
	bytesReadByJSON := bytesUnreadByUs[:numBytesParsedByJSON]            // "{..} XYZ" -> "{..}"
	j.Text = strings.TrimSpace(string(bytesReadByJSON))
	j.textBuffer.Next(numBytesParsedByJSON) // "*{..} XYZ" -> "*XYZ"
	if err != nil {
		if j.decodeError == nil || j.decodeError == io.EOF {
			j.decodeError = err
		}
		return false
	}
	return true
}
