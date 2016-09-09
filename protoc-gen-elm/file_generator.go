package main

import (
	"fmt"
	"io"
	"strings"
)

type FileGenerator struct {
	w      io.Writer
	indent uint
}

func NewFileGenerator(w io.Writer) *FileGenerator {
	return &FileGenerator{
		w: w,
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

	_, err = fmt.Fprintf(fg.w, strings.Repeat("  ", int(fg.indent)))
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(fg.w, format, a...)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(fg.w, "\n")
	if err != nil {
		return err
	}

	return nil
}
