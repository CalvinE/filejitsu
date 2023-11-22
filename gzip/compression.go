package gzip

import (
	"compress/gzip"
	"errors"
)

type GZipCompressionLevel string

var (
	ErrInvalidGZipLevel = errors.New("invalid GZIP level provided")
)

const (
	NoCompression      GZipCompressionLevel = "NoCompression"
	BestSpeed          GZipCompressionLevel = "BestSpeed"
	BestCompression    GZipCompressionLevel = "BestCompression"
	HuffmanOnly        GZipCompressionLevel = "HuffmanOnly"
	DefaultCompression GZipCompressionLevel = "DefaultCompression"
)

func GZipCompressionLevelToLevel(level GZipCompressionLevel) (int, error) {
	switch level {
	case NoCompression:
		return gzip.NoCompression, nil
	case BestSpeed:
		return gzip.BestSpeed, nil
	case BestCompression:
		return gzip.BestCompression, nil
	case HuffmanOnly:
		return gzip.HuffmanOnly, nil
	case DefaultCompression:
		return gzip.DefaultCompression, nil
	}
	return 0, ErrInvalidGZipLevel
}
