package main

import (
	"bufio"
	"os"
)

type lineStream struct {
	token
	Text string `json:"text,omitempty"`

	buffer  *tokenBuffer
	scanner *bufio.Scanner
}

func newLineStream() *lineStream {
	return &lineStream{
		scanner: bufio.NewScanner(os.Stdin),
		buffer:  &tokenBuffer{},
	}
}

func (l *lineStream) Token() *token {
	return &l.token
}

func (l *lineStream) CopyCurrent() tokenStream {
	return &lineStream{
		token: l.token,
		Text:  l.Text,
	}
}

func (l *lineStream) AppendCurrentToBuffer() {
	*l.buffer = append(*l.buffer, l.CopyCurrent())
}

func (l *lineStream) FlushBuffer() {
	l.buffer.flush(l)
}

func (l *lineStream) CurrentMatchText() string {
	if matchTemplate != nil {
		return matchTemplate.execute(l.Text)
	}
	return l.Text
}

func (l *lineStream) Err() error {
	return l.scanner.Err()
}

func (l *lineStream) Scan() bool {
	l.Text = ""
	ok := l.scanner.Scan()
	if ok {
		l.Text = l.scanner.Text()
	}
	return ok
}
