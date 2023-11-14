package util

import "fmt"

const (
	B  float64 = 1
	KB float64 = 1 << (10 * iota)
	MB
	GB
	TB
	PB
	EB
)

func GetPrettyBytesSize(sizeInBytes int64) string {
	sizeInBytesFloat := float64(sizeInBytes)
	var divSize float64
	var unit string
	var unitSize float64
	formatString := "%.2f %s"
	if sizeInBytesFloat >= EB {
		unitSize = EB
		unit = "EB"
	} else if sizeInBytesFloat >= PB {
		unitSize = PB
		unit = "PB"
	} else if sizeInBytesFloat >= TB {
		unitSize = TB
		unit = "TB"
	} else if sizeInBytesFloat >= GB {
		unitSize = GB
		unit = "GB"
	} else if sizeInBytesFloat >= MB {
		unitSize = MB
		unit = "MB"
	} else if sizeInBytesFloat >= KB {
		unitSize = KB
		unit = "KB"
	} else {
		formatString = "%.0f %s"
		unitSize = B
		unit = "B"
	}
	divSize = sizeInBytesFloat / unitSize
	prettySize := fmt.Sprintf(formatString, divSize, unit)
	return prettySize
}
