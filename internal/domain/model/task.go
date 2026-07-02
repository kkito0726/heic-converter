package model

// DefaultQuality is the encoder quality used when none (or an invalid one)
// is specified.
const DefaultQuality = 90

// EncodeOptions carries encoder tuning parameters shared by all encoders.
type EncodeOptions struct {
	// Quality is the lossy-compression quality in the range 1-100.
	// Lossless formats ignore it.
	Quality int
}

// NewEncodeOptions builds EncodeOptions, falling back to DefaultQuality when
// quality is out of the 1-100 range.
func NewEncodeOptions(quality int) EncodeOptions {
	if quality < 1 || quality > 100 {
		quality = DefaultQuality
	}
	return EncodeOptions{Quality: quality}
}

// ConversionResult describes the outcome of converting one source file.
type ConversionResult struct {
	SourcePath string
	// Outputs holds the paths that were written successfully.
	Outputs []string
	// Err is non-nil if any part of the conversion failed.
	Err error
}

// Succeeded reports whether the whole conversion of the source completed
// without errors.
func (r ConversionResult) Succeeded() bool { return r.Err == nil }
