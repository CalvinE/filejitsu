package streamingjson

import (
	"bytes"
	"context"
	"testing"
	"time"
)

func Test_lengthPrefixStreamingJSONWrite(t *testing.T) {
	type testStruct struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}
	jsonStreamer := NewLengthPrefixStreamJSONHandler[testStruct]()
	testPayload := testStruct{
		Name:  "a",
		Value: "b",
	}
	data := make([]byte, 0, 0)
	dataBuffer := bytes.NewBuffer(data)
	writeContext, _ := context.WithTimeout(context.Background(), time.Millisecond*1)
	_, err := jsonStreamer.WriteObject(writeContext, testPayload, dataBuffer)
	if err != nil {
		t.Error(err)
	}
	if len(dataBuffer.Bytes()) == 0 {
		t.Error("uhh.")
	}
	readContext, _ := context.WithTimeout(context.TODO(), time.Millisecond*1)
	bytesConsumed, obj, err := jsonStreamer.ReadNext(readContext, dataBuffer)
	if err != nil {
		t.Error("buhh..", err)
	}
	if bytesConsumed != 26 {
		t.Error("damn")
	}
	if obj.Name != "a" && obj.Value != "b" {
		t.Error(":-(")
	}
}

func Test_delimitedStreamingJSONWrite(t *testing.T) {
	type testStruct struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}
	delimiter := []byte{0x30, 0x30}
	jsonStreamer, _ := NewDelimitedStreamJSONHandler[testStruct](delimiter)
	testPayload := testStruct{
		Name:  "a",
		Value: "b",
	}
	data := make([]byte, 0, 0)
	dataBuffer := bytes.NewBuffer(data)
	writeContext, _ := context.WithTimeout(context.Background(), time.Millisecond*1)
	_, err := jsonStreamer.WriteObject(writeContext, testPayload, dataBuffer)
	if err != nil {
		t.Error(err)
	}
	if len(dataBuffer.Bytes()) == 0 {
		t.Error("uhh.")
	}
	readContext, _ := context.WithTimeout(context.TODO(), time.Hour*1)
	bytesConsumed, obj, err := jsonStreamer.ReadNext(readContext, dataBuffer)
	if err != nil {
		t.Error("buhh..", err)
	}
	if bytesConsumed != 26 {
		t.Error("damn")
	}
	if obj.Name != "a" && obj.Value != "b" {
		t.Error(":-(")
	}
}
