package streamingjson

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"
)

type testStruct struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func TestLengthPrefixReadNext(t *testing.T) {
	type testCase struct {
		Name                string
		Data                []byte
		ExpectedObjects     []testStruct
		ExpectedTotalLength int
		ExpectedError       error
	}
	testCases := []testCase{
		{
			Name: "Single Object",
			Data: []byte("24{\"name\":\"a\",\"value\":\"b\"}"),
			ExpectedObjects: []testStruct{
				{
					Name:  "a",
					Value: "b",
				},
			},
			ExpectedTotalLength: 26,
			ExpectedError:       nil,
		},
		{
			Name: "Multiple Objects",
			Data: []byte("24{\"name\":\"a\",\"value\":\"b\"}24{\"name\":\"c\",\"value\":\"d\"}26{\"name\":\"ef\",\"value\":\"gh\"}"),
			ExpectedObjects: []testStruct{
				{
					Name:  "a",
					Value: "b",
				},
				{
					Name:  "c",
					Value: "d",
				},
				{
					Name:  "ef",
					Value: "gh",
				},
			},
			ExpectedTotalLength: 26 + 26 + 28,
			ExpectedError:       nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ExpectedObjectCount := len(tc.ExpectedObjects)
			streamHandler := NewLengthPrefixStreamJSONHandler[testStruct]()
			buffer := bytes.NewBuffer([]byte{})
			buffer.Write(tc.Data)
			totalLength := 0
			numObjects := 0
			for {
				n, obj, err := streamHandler.ReadNext(context.TODO(), buffer)
				totalLength += n
				if err != nil {
					if err == io.EOF && n == 0 {
						// we hit the end of the buffer and we are done!
						break
					}
					t.Errorf("unexpectedError Encountered: %v", err)
				}
				if tc.ExpectedObjects[numObjects].Name != obj.Name {
					t.Errorf("object name not expected value: got - %s - expected - %s", obj.Name, tc.ExpectedObjects[numObjects].Name)
				}
				if tc.ExpectedObjects[numObjects].Value != obj.Value {
					t.Errorf("object value not expected value: got - %s - expected - %s", obj.Value, tc.ExpectedObjects[numObjects].Value)
				}
				numObjects++
			}
			if numObjects != ExpectedObjectCount {
				t.Errorf("did not get the right number of objects: got - %d - expected - %d", numObjects, ExpectedObjectCount)
			}
			if totalLength != tc.ExpectedTotalLength {
				t.Errorf("total length written value not expected value: got - %d - expected - %d", totalLength, tc.ExpectedTotalLength)
			}
		})
	}
}

func BenchmarkLengthPrefixReader(b *testing.B) {
	stringData := []byte("24{\"name\":\"a\",\"value\":\"b\"}")
	streamHandler := NewLengthPrefixStreamJSONHandler[testStruct]()
	// var bytesRead int
	// var obj testStruct
	var err error
	for i := 0; i < b.N; i++ {
		buffer := bytes.NewBuffer(stringData)
		_, _, err = streamHandler.ReadNext(context.TODO(), buffer)
		if err != nil {
			b.Errorf("unexpected error encountered: %v", err)
		}
	}
}

func TestLengthPrefixReadAll(t *testing.T) {
	type testCase struct {
		Name                string
		Data                []byte
		ExpectedObjects     []testStruct
		ExpectedTotalLength int
		ExpectedError       error
	}
	testCases := []testCase{
		{
			Name: "Single Object",
			Data: []byte("24{\"name\":\"a\",\"value\":\"b\"}"),
			ExpectedObjects: []testStruct{
				{
					Name:  "a",
					Value: "b",
				},
			},
			ExpectedTotalLength: 26,
			ExpectedError:       nil,
		},
		{
			Name: "Multiple Objects",
			Data: []byte("24{\"name\":\"a\",\"value\":\"b\"}24{\"name\":\"c\",\"value\":\"d\"}26{\"name\":\"ef\",\"value\":\"gh\"}"),
			ExpectedObjects: []testStruct{
				{
					Name:  "a",
					Value: "b",
				},
				{
					Name:  "c",
					Value: "d",
				},
				{
					Name:  "ef",
					Value: "gh",
				},
			},
			ExpectedTotalLength: 26 + 26 + 28,
			ExpectedError:       nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			expectedObjectCount := len(tc.ExpectedObjects)
			streamHandler := NewLengthPrefixStreamJSONHandler[testStruct]()
			buffer := bytes.NewBuffer([]byte{})
			buffer.Write(tc.Data)
			n, objects, err := streamHandler.ReadAll(context.TODO(), buffer)
			numObjects := len(objects)
			if numObjects != expectedObjectCount {
				t.Errorf("did not get the right number of objects: got - %d - expected - %d", numObjects, expectedObjectCount)
			}
			if n != tc.ExpectedTotalLength {
				t.Errorf("total length written value not expected value: got - %d - expected - %d", n, tc.ExpectedTotalLength)
			}
			if err != tc.ExpectedError {
				t.Errorf("got unexpected error: %v", err)
			}
			for i := range tc.ExpectedObjects {
				if tc.ExpectedObjects[i].Name != objects[i].Name {
					t.Errorf("object name not expected value: got - %s - expected - %s", objects[i].Name, tc.ExpectedObjects[i].Name)
				}
				if tc.ExpectedObjects[i].Value != objects[i].Value {
					t.Errorf("object value not expected value: got - %s - expected - %s", objects[i].Value, tc.ExpectedObjects[i].Value)
				}
			}
		})
	}
}

func TestLengthPrefixWriter(t *testing.T) {
	type testCase struct {
		Name                 string
		ObjToWrite           testStruct
		ExpectedOutput       []byte
		ExpectedBytesWritten int
		ExpectedError        error
	}
	testCases := []testCase{
		{
			Name: "Simple test",
			ObjToWrite: testStruct{
				Name:  "a",
				Value: "b",
			},
			ExpectedOutput:       []byte("24{\"name\":\"a\",\"value\":\"b\"}"),
			ExpectedBytesWritten: 26,
			ExpectedError:        nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			buffer := bytes.NewBuffer([]byte{})
			streamHandler := NewLengthPrefixStreamJSONHandler[testStruct]()
			n, err := streamHandler.WriteObject(context.TODO(), tc.ObjToWrite, buffer)
			if n != tc.ExpectedBytesWritten {
				t.Errorf("total length written value not expected value: got - %d - expected - %d", n, tc.ExpectedBytesWritten)
			}
			if err != tc.ExpectedError {
				t.Errorf("got unexpected error: %v", err)
			}
			writtenString := buffer.String()
			if writtenString != string(tc.ExpectedOutput) {
				t.Errorf("written data is not expected value: got - %s - expected - %s", writtenString, tc.ExpectedOutput)
			}
		})
	}
}

func TestDelimitedStreamingJSONRead(t *testing.T) {
	type testCase struct {
		Name                string
		Delimiter           []byte
		Data                []byte
		ExpectedObjects     []testStruct
		ExpectedTotalLength int
		ExpectedError       error
	}
	testCases := []testCase{
		{
			Name:      "Single Object",
			Delimiter: []byte{0x30},
			Data:      []byte("{\"name\":\"a\",\"value\":\"b\"}\x30"),
			ExpectedObjects: []testStruct{
				{
					Name:  "a",
					Value: "b",
				},
			},
			ExpectedTotalLength: 24 + 1,
			ExpectedError:       nil,
		},
		{
			Name:      "Multiple Objects",
			Delimiter: []byte{0x30},
			Data:      []byte("{\"name\":\"a\",\"value\":\"b\"}\x30{\"name\":\"c\",\"value\":\"d\"}\x30{\"name\":\"ef\",\"value\":\"gh\"}\x30"),
			ExpectedObjects: []testStruct{
				{
					Name:  "a",
					Value: "b",
				},
				{
					Name:  "c",
					Value: "d",
				},
				{
					Name:  "ef",
					Value: "gh",
				},
			},
			ExpectedTotalLength: 24 + 1 + 24 + 1 + 26 + 1,
			ExpectedError:       nil,
		},
		{
			Name:                "No Object Before Delimiter",
			Delimiter:           []byte{0x30},
			Data:                []byte{0x30},
			ExpectedObjects:     nil,
			ExpectedTotalLength: 1,
			ExpectedError:       ErrNoObjectBeforeDelimiter,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ExpectedObjectCount := len(tc.ExpectedObjects)
			streamHandler, err := NewDelimitedStreamJSONHandler[testStruct](tc.Delimiter)
			if err != nil {
				t.Errorf("failed to construct delimited JSON streaming handler")
				return
			}
			buffer := bytes.NewBuffer([]byte{})
			buffer.Write(tc.Data)
			totalLength := 0
			numObjects := 0
			for {
				n, obj, err := streamHandler.ReadNext(context.TODO(), buffer)
				totalLength += n
				if err != nil {
					if err == io.EOF && n == 0 {
						// we hit the end of the buffer and we are done!
						break
					} else if err == tc.ExpectedError {
						return
					}
					t.Errorf("unexpectedError Encountered: %v", err)
					return
				}
				if tc.ExpectedObjects[numObjects].Name != obj.Name {
					t.Errorf("object name not expected value: got - %s - expected - %s", obj.Name, tc.ExpectedObjects[numObjects].Name)
				}
				if tc.ExpectedObjects[numObjects].Value != obj.Value {
					t.Errorf("object value not expected value: got - %s - expected - %s", obj.Value, tc.ExpectedObjects[numObjects].Value)
				}
				numObjects++
			}
			if numObjects != ExpectedObjectCount {
				t.Errorf("did not get the right number of objects: got - %d - expected - %d", numObjects, ExpectedObjectCount)
			}
			if totalLength != tc.ExpectedTotalLength {
				t.Errorf("total length written value not expected value: got - %d - expected - %d", totalLength, tc.ExpectedTotalLength)
			}
		})
	}
}

func BenchmarkDelimitedReader(b *testing.B) {
	delimiter := []byte{0x30}
	stringData := []byte("{\"name\":\"a\",\"value\":\"b\"}\x30")
	streamHandler, err := NewDelimitedStreamJSONHandler[testStruct](delimiter)
	if err != nil {
		b.Errorf("failed to construct delimited streaming JSON handler: %v", err)
	}
	// var bytesRead int
	// var obj testStruct
	for i := 0; i < b.N; i++ {
		buffer := bytes.NewBuffer(stringData)
		_, _, err = streamHandler.ReadNext(context.TODO(), buffer)
		if err != nil {
			b.Errorf("unexpected error encountered: %v", err)
		}
	}
}

func TestDelimitedStreamingJSONWrite(t *testing.T) {
	type testCase struct {
		Name                 string
		Delimiter            []byte
		ObjToWrite           testStruct
		ExpectedBytesWritten int
		ExpectedOutput       []byte
		ExpectedError        error
	}
	testCases := []testCase{
		{
			Name:      "Simple Write",
			Delimiter: []byte{0x30},
			ObjToWrite: testStruct{
				Name:  "a",
				Value: "b",
			},
			ExpectedBytesWritten: 25,
			ExpectedOutput:       []byte("{\"name\":\"a\",\"value\":\"b\"}\x30"),
			ExpectedError:        nil,
		},
		{
			Name:      "Two Byte Delimiter Write",
			Delimiter: []byte{0x30, 0x30},
			ObjToWrite: testStruct{
				Name:  "a",
				Value: "b",
			},
			ExpectedBytesWritten: 26,
			ExpectedOutput:       []byte("{\"name\":\"a\",\"value\":\"b\"}\x30\x30"),
			ExpectedError:        nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			jsonStreamer, _ := NewDelimitedStreamJSONHandler[testStruct](tc.Delimiter)
			dataBuffer := bytes.NewBuffer([]byte{})
			writeContext, cancel := context.WithTimeout(context.Background(), time.Hour*1)
			defer cancel()
			n, err := jsonStreamer.WriteObject(writeContext, tc.ObjToWrite, dataBuffer)
			if err != tc.ExpectedError {
				t.Errorf("got unexpected error: %v", err)
			}
			if n != tc.ExpectedBytesWritten {
				t.Errorf("total length written value not expected value: got - %d - expected - %d", n, tc.ExpectedBytesWritten)
			}
			if dataBuffer.String() != string(tc.ExpectedOutput) {
				t.Errorf("output did not match expected value")
			}
			// readContext, cancel := context.WithTimeout(context.TODO(), time.Hour*1)
			// defer cancel()
			// bytesConsumed, obj, err := jsonStreamer.ReadNext(readContext, dataBuffer)
			// if err != nil {
			// 	t.Error("buhh..", err)
			// }
			// if bytesConsumed != 26 {
			// 	t.Error("damn")
			// }
			// if obj.Name != "a" && obj.Value != "b" {
			// 	t.Error(":-(")
			// }
		})
	}
}
