package main

// EncDec defines a encoder and decoder
type EncDec struct {
	Encoder string
	Decoder string
}

// CustomType defines a custom type
type CustomType struct {
	JSON    EncDec
	Default string
}

// FieldOptions overrides the data type of a field
type FieldOptions struct {
	Type     string
	Required bool
}

// FileOptions contains file-specific options
type FileOptions struct {
	Imports []string
	Fields  map[string]FieldOptions
}

// Options contains the code generator options
type Options struct {
	Types map[string]CustomType
	Files map[string]FileOptions
}
