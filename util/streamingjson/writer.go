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

const (
	defaultMinBufferSize = 2048
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
	totalBytesRead := 0
	doneReadingLength := false
	for {
		if doneReadingLength {
			break
		}
		select {
		case <-ctx.Done():
			// context was cancelled... let bail?
			return totalBytesRead, result, ErrContextDone
		default:
			upperLimit := totalBytesRead + 1
			j.growInternalBufferIfNeeded(upperLimit)
			readBuffer := j.DataBuffer[totalBytesRead:upperLimit]
			n, err := reader.Read(readBuffer)
			totalBytesRead += n
			if err != nil {
				if err == io.EOF {
					if totalBytesRead == 0 {
						return totalBytesRead, result, io.EOF
					}
					return totalBytesRead, result, fmt.Errorf("hit EOF on reader before encountering a '{': %w", err)
				}
				return totalBytesRead, result, fmt.Errorf("reader encountered an error: %w", err)
			}
			if readBuffer[0] == '{' {
				doneReadingLength = true
				break
			}
		}
	}
	readLengthBytes := totalBytesRead - 1
	if readLengthBytes == 0 {
		return 1, result, errors.New("no length was found before the JSON object")
	}
	length, err := strconv.ParseInt(string(j.DataBuffer[:readLengthBytes]), 10, 64)
	if err != nil {
		return 0, result, err
	}
	totalLength := length + int64(readLengthBytes)
	doneReadingBody := false
	upperLimit := totalBytesRead + int(length-1)
	j.growInternalBufferIfNeeded(upperLimit)
	// TODO: need to add a buffer for reading the JSON data. Could not all be read in one pull, so need to keep track of how much is read?
	// Can we resize a slice to make it "right sized" for subsequent pulls?terraform
	for {
		if doneReadingBody {
			break
		}
		select {
		case <-ctx.Done():
			// context was cancelled... let bail?
			return totalBytesRead, result, ErrContextDone
		default:
			remainingBuffer := j.DataBuffer[totalBytesRead:upperLimit]
			n, err := reader.Read(remainingBuffer)
			totalBytesRead += n
			if int64(totalBytesRead) >= totalLength {
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
	if err = json.Unmarshal(j.DataBuffer[readLengthBytes:upperLimit], &result); err != nil {
		return readLengthBytes, result, err
	}
	return totalBytesRead, result, nil
}

func (j *jsonStreamer[T]) delimitedStreamingJSONRead(ctx context.Context, reader io.Reader) (int, T, error) {
	var result T
	bytesRead := 0
	delimiterLength := len(j.Delimiter)
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
			upperLimit := bytesRead + delimiterLength - delimiterBytesMatch
			j.growInternalBufferIfNeeded(upperLimit)
			readBuffer := j.DataBuffer[bytesRead:upperLimit]
			n, err := reader.Read(readBuffer[delimiterBytesMatch:])
			bytesRead += n
			if n > 0 {
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
		err := json.Unmarshal(j.DataBuffer[:dataLength], &result)
		if err != nil {
			return bytesRead, result, fmt.Errorf("failed to unmarshal data: %w", err)
		}
		return bytesRead, result, nil
	}
	// If we are here we only read a delimiter? So return empty object?
	return bytesRead, result, ErrNoObjectBeforeDelimiter
}

func (j *jsonStreamer[T]) growInternalBufferIfNeeded(index int) {
	if cap(j.DataBuffer) < index {
		additionalBuffer := make([]byte, 512)
		j.DataBuffer = append(j.DataBuffer, additionalBuffer...)
	}
}

type StreamingJSONHandler[T any] interface {
	StreamingJSONReader[T]
	StreamingJSONWriter[T]
}

type StreamJSONWriteObjectFunc[T any] func(ctx context.Context, obj T, output io.Writer) (int, error)

type StreamingJSONWriter[T any] interface {
	WriteObject(ctx context.Context, obj T, output io.Writer) (int, error)
}

type StreamJSONReadNextFunc[T any] func(ctx context.Context, input io.Reader) (int, T, error)

type StreamingJSONReader[T any] interface {
	ReadNext(ctx context.Context, input io.Reader) (int, T, error)
	ReadAll(ctx context.Context, input io.Reader) (int, []T, error)
}

type jsonStreamer[T any] struct {
	DataBuffer    []byte
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

func NewLengthPrefixStreamJSONHandler[T any]() StreamingJSONHandler[T] {
	jsonStreamer := &jsonStreamer[T]{
		DataBuffer: make([]byte, defaultMinBufferSize),
	}
	jsonStreamer.readFunction = jsonStreamer.lengthPrefixStreamingJSONRead
	jsonStreamer.writeFunction = jsonStreamer.lengthPrefixStreamingJSONWrite
	// If delimiter provided, set funcs to delimiter version

	return jsonStreamer
}

func NewDelimitedStreamJSONHandler[T any](delimiter []byte) (StreamingJSONHandler[T], error) {
	jsonStreamer := &jsonStreamer[T]{
		DataBuffer: make([]byte, defaultMinBufferSize),
		Delimiter:  delimiter,
	}
	if len(delimiter) == 0 {
		return jsonStreamer, errors.New("delimiter is required")
	}
	jsonStreamer.readFunction = jsonStreamer.delimitedStreamingJSONRead
	jsonStreamer.writeFunction = jsonStreamer.delimitedStreamingJSONWrite
	// If delimiter provided, set funcs to delimiter version

	return jsonStreamer, nil
}
