package gzip

import (
	"bytes"
	"compress/gzip"
	"log/slog"
	"os"
	"slices"
	"testing"
	"time"
)

func TestGzip(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug.Level(),
	}))
	testString := "This is a test string"
	input := bytes.NewBuffer([]byte(testString))
	output := bytes.NewBuffer([]byte{})
	gzipOut, err := NewGZIPWriter(logger, output, gzip.DefaultCompression, gzip.Header{})
	if err != nil {
		t.Error(err)
		return
	}
	if err := Compress(logger, input, gzipOut); err != nil {
		t.Error(err)
	}
	outputData := slices.Clone(output.Bytes())
	outputString := string(outputData)
	t.Log(outputString)
	// if testString != outputString {
	// 	t.Error("test string does not equal output")
	// 	return
	// }
	output2 := bytes.NewBuffer([]byte{})
	gzipIn, header, err := NewGZIPReader(logger, output)
	err = Decompress(logger, gzipIn, output2)
	if err != nil {
		t.Error("failed to decompress gzip")
		return
	}
	t.Log(header)
	output2Data := slices.Clone(output2.Bytes())
	output2String := string(output2Data)
	if testString != output2String {
		t.Error("test string does not equal output")
		return
	}
}

func TestGZipWithHeader(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug.Level(),
	}))
	comment := "this is a comment about the content"
	extra := "This is extra data about the content"
	name := "file.txt"
	modTime := time.Now()
	inputHeader := gzip.Header{
		Comment: comment,
		Extra:   []byte(extra),
		ModTime: modTime,
		Name:    name,
		OS:      13,
	}
	testString := "This is a test string"
	input := bytes.NewBuffer([]byte(testString))
	output := bytes.NewBuffer([]byte{})
	gzipOut, err := NewGZIPWriter(logger, output, gzip.DefaultCompression, inputHeader)
	if err != nil {
		t.Error(err)
		return
	}
	if err := Compress(logger, input, gzipOut); err != nil {
		t.Error(err)
	}
	outputData := slices.Clone(output.Bytes())
	outputString := string(outputData)
	t.Log(outputString)
	// if testString != outputString {
	// 	t.Error("test string does not equal output")
	// 	return
	// }
	output2 := bytes.NewBuffer([]byte{})
	gzipIn, header, err := NewGZIPReader(logger, output)
	err = Decompress(logger, gzipIn, output2)
	if err != nil {
		t.Error("failed to decompress gzip")
		return
	}
	t.Log(header)
	if header.Comment != comment {
		t.Error("decompressed header comment does not match what was passed in")
		return
	}
	if header.Name != name {
		t.Error("decompressed header name does not match what was passed in")
		return
	}
	if string(header.Extra) != extra {
		t.Error("decompressed header extra does not match what was passed in")
		return
	}
	isAfterOneSecond := header.ModTime.After(modTime.Add(-1 * time.Second))
	isBeforeOneSecond := header.ModTime.Before(modTime.Add(1 * time.Second))
	if !isBeforeOneSecond || !isAfterOneSecond {
		t.Error("decompressed header modTime does not match what was passed in")
		return
	}
	if header.OS != inputHeader.OS {
		t.Error("decompressed header os does not match what was passed in")
		return
	}
	output2Data := slices.Clone(output2.Bytes())
	output2String := string(output2Data)
	if testString != output2String {
		t.Error("test string does not equal output")
		return
	}
}
