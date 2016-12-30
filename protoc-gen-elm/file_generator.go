package main

import (
	"fmt"
	"io"
	"strings"
)

type FileGenerator struct {
	w io.Writer
	// Used to avoid qualifying names in the same file.
	inFileName string
	indent     uint
}

func NewFileGenerator(w io.Writer, inFileName string) *FileGenerator {
	return &FileGenerator{
		w:          w,
		inFileName: inFileName,
	}
}

func (fg *FileGenerator) In() {
	fg.indent++
}

func (fg *FileGenerator) Out() {
	fg.indent--
}

func (fg *FileGenerator) P(format string, a ...interface{}) error {
	var err error

	// If format is empty, avoid printing just whitespaces.
	if format != "" {
		_, err = fmt.Fprintf(fg.w, strings.Repeat("    ", int(fg.indent)))
		if err != nil {
			return err
		}

		_, err = fmt.Fprintf(fg.w, format, a...)
		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprintf(fg.w, "\n")
	if err != nil {
		return err
	}

	return nil
}
