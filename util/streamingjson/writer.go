package streamingjson

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
)

var (
	ErrContextDone             = errors.New("operation cancelled by context")
	ErrNoObjectBeforeDelimiter = errors.New("no object was present before delimiter")
)

// https://en.wikipedia.org/wiki/JSON_streaming

// This works with JSON objects specifically... no "raw" values...

func (j *jsonStreamer[T]) lengthPrefixStreamingJSONWrite(ctx context.Context, obj T, writer io.Writer) (int, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal object: %w", err)
	}
	dataLength := len(data)
	lengthString := strconv.FormatInt(int64(dataLength), 10)
	totalWritten := 0
	w1, err := writer.Write([]byte(lengthString))
	totalWritten += w1
	if err != nil {
		return totalWritten, fmt.Errorf("failed to write length to writer: %w", err)
	}
	w2, err := writer.Write(data)
	totalWritten += w2
	if err != nil {
		return totalWritten, fmt.Errorf("failed to write object to writer: %w", err)
	}
	return totalWritten, nil
}

func (j *jsonStreamer[T]) delimitedStreamingJSONWrite(ctx context.Context, obj T, writer io.Writer) (int, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal object: %w", err)
	}
	totalWritten := 0
	w1, err := writer.Write(data)
	totalWritten += w1
	if err != nil {
		return totalWritten, fmt.Errorf("failed to write data to writer: %w", err)
	}
	w2, err := writer.Write(j.Delimiter)
	totalWritten += w2
	if err != nil {
		return totalWritten, fmt.Errorf("failed to write delimiter to writer: %w", err)
	}
	return totalWritten, nil

}

func (j *jsonStreamer[T]) lengthPrefixStreamingJSONRead(ctx context.Context, reader io.Reader) (int, T, error) {
	var result T
	readBuffer := make([]byte, 1)
	readLengthBytes := 0
	lengthBuffer := make([]byte, 0, 16)
	doneReadingLength := false
	for {
		if doneReadingLength {
			break
		}
		select {
		case <-ctx.Done():
			// context was cancelled... let bail?
			return readLengthBytes, result, ErrContextDone
		default:
			n, err := reader.Read(readBuffer)
			readLengthBytes += n
			if err != nil {
				if err == io.EOF {
					if readLengthBytes == 0 {
						return readLengthBytes, result, io.EOF
					}
					return readLengthBytes, result, fmt.Errorf("hit EOF on reader before encountering a '{': %w", err)
				}
				return readLengthBytes, result, fmt.Errorf("reader encountered an error: %w", err)
			}
			if readBuffer[0] == '{' {
				readLengthBytes--
				doneReadingLength = true
				break
			}
			lengthBuffer = append(lengthBuffer, readBuffer[0])
		}
	}
	if len(lengthBuffer) == 0 {
		return 1, result, errors.New("no length was found before the JSON object")
	}
	length, err := strconv.ParseInt(string(lengthBuffer[:readLengthBytes]), 10, 64)
	if err != nil {
		return 0, result, err
	}
	jsonBuffer := make([]byte, length)
	jsonBuffer[0] = '{'
	totalLength := length + int64(readLengthBytes)
	readBodyBytes := 1
	doneReadingBody := false
	// TODO: need to add a buffer for reading the JSON data. Could not all be read in one pull, so need to keep track of how much is read?
	// Can we resize a slice to make it "right sized" for subsequent pulls?terraform
	for {
		if doneReadingBody {
			break
		}
		select {
		case <-ctx.Done():
			// context was cancelled... let bail?
			return readLengthBytes + readBodyBytes, result, ErrContextDone
		default:
			remainingBuffer := jsonBuffer[readBodyBytes:]
			n, err := reader.Read(remainingBuffer)
			readBodyBytes += n
			if int64(readBodyBytes+readLengthBytes) >= totalLength {
				// We have read the whole object so we are good to bail!
				doneReadingBody = true
				break
			}
			if err != nil {
				if err == io.EOF {
					return readLengthBytes, result, err
				}
				return readLengthBytes, result, fmt.Errorf("reader encountered an error: %w", err)
			}
		}
	}
	if err = json.Unmarshal(jsonBuffer, &result); err != nil {
		return readLengthBytes, result, err
	}
	return readLengthBytes + readBodyBytes, result, nil
}

func (j *jsonStreamer[T]) delimitedStreamingJSONRead(ctx context.Context, reader io.Reader) (int, T, error) {
	var result T
	bytesRead := 0
	delimiterLength := len(j.Delimiter)
	readBuffer := make([]byte, delimiterLength)
	jsonBuffer := make([]byte, 0, 1024)
	delimiterBytesMatch := 0
	done := false
	// TODO: need a buffer for JSON data. after delimiter we can chop off the delimiter and JSON unmarshal the thing.
	for {
		if done {
			break
		}
		select {
		case <-ctx.Done():
			return bytesRead, result, ErrContextDone
		default:
			n, err := reader.Read(readBuffer[delimiterBytesMatch:])
			bytesRead += n
			if n > 0 {
				jsonBuffer = append(jsonBuffer, readBuffer[:n]...)
				for i := 0; i < n; i++ {
					if readBuffer[i] != j.Delimiter[delimiterBytesMatch] {
						delimiterBytesMatch = 0
					} else {
						delimiterBytesMatch++
						if delimiterBytesMatch == delimiterLength {
							// We are done, bail!
							done = true
							break
						}
					}
				}
			}
			// TODO: have a check for error function for length prefix and delimited?
			// Or am I over thinking it. The delimited works by having an object then a delimiter and repeat. Starting with the delimiter is dumb and should result in a error?
			if err != nil {
				if err == io.EOF {
					return bytesRead, result, err
				}
				return bytesRead, result, fmt.Errorf("reader encountered an error: %w", err)
			}
		}
	}
	dataLength := bytesRead - delimiterLength
	if dataLength > 0 {
		err := json.Unmarshal(jsonBuffer[:dataLength], &result)
		if err != nil {
			return bytesRead, result, fmt.Errorf("failed to unmarshal data: %w", err)
		}
		return bytesRead, result, nil
	}
	// If we are here we only read a delimiter? So return empty object?
	return bytesRead, result, ErrNoObjectBeforeDelimiter
}

type StreamingJSONHandler[T comparable] interface {
	StreamingJSONReader[T]
	StreamingJSONWriter[T]
}

type StreamJSONWriteObjectFunc[T comparable] func(ctx context.Context, obj T, output io.Writer) (int, error)

type StreamingJSONWriter[T comparable] interface {
	WriteObject(ctx context.Context, obj T, output io.Writer) (int, error)
}

type StreamJSONReadNextFunc[T comparable] func(ctx context.Context, input io.Reader) (int, T, error)

type StreamingJSONReader[T comparable] interface {
	ReadNext(ctx context.Context, input io.Reader) (int, T, error)
	ReadAll(ctx context.Context, input io.Reader) (int, []T, error)
}

type jsonStreamer[T comparable] struct {
	// Buffer        *bytes.Buffer
	Delimiter     []byte
	readFunction  StreamJSONReadNextFunc[T]
	writeFunction StreamJSONWriteObjectFunc[T]
}

func (j *jsonStreamer[T]) WriteObject(ctx context.Context, input T, output io.Writer) (int, error) {
	return j.writeFunction(ctx, input, output)
}

func (j *jsonStreamer[T]) ReadNext(ctx context.Context, input io.Reader) (int, T, error) {
	return j.readFunction(ctx, input)
}

func (j *jsonStreamer[T]) ReadAll(ctx context.Context, reader io.Reader) (int, []T, error) {
	results := make([]T, 0, 5)
	totalBytesRead := 0
	for {
		n, obj, err := j.readFunction(ctx, reader)
		totalBytesRead += n
		if err != nil {
			if err == io.EOF && n == 0 {
				break
			}
			return totalBytesRead, results, err
		}
		results = append(results, obj)
	}
	return totalBytesRead, results, nil
}

func NewLengthPrefixStreamJSONHandler[T comparable]() StreamingJSONHandler[T] {
	jsonStreamer := &jsonStreamer[T]{}
	jsonStreamer.readFunction = jsonStreamer.lengthPrefixStreamingJSONRead
	jsonStreamer.writeFunction = jsonStreamer.lengthPrefixStreamingJSONWrite
	// If delimiter provided, set funcs to delimiter version

	return jsonStreamer
}

func NewDelimitedStreamJSONHandler[T comparable](delimiter []byte) (StreamingJSONHandler[T], error) {
	jsonStreamer := &jsonStreamer[T]{
		Delimiter: delimiter,
	}
	if len(delimiter) == 0 {
		return jsonStreamer, errors.New("delimiter is required")
	}
	jsonStreamer.readFunction = jsonStreamer.delimitedStreamingJSONRead
	jsonStreamer.writeFunction = jsonStreamer.delimitedStreamingJSONWrite
	// If delimiter provided, set funcs to delimiter version

	return jsonStreamer, nil
}
