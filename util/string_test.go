package util

import "testing"

func TestPadLeft(t *testing.T) {
	type testCase struct {
		name           string
		s              string
		padChar        string
		desiredLength  int
		expectedOutput string
	}
	testCases := []testCase{
		{
			name:           "pad single digit with zeros",
			s:              "1",
			padChar:        "0",
			desiredLength:  2,
			expectedOutput: "01",
		},
		{
			name:           "2 char pad char string",
			s:              "w",
			padChar:        "12",
			desiredLength:  2,
			expectedOutput: "12w",
		},
		{
			name:           "string is already desired length",
			s:              "ww",
			padChar:        "1",
			desiredLength:  2,
			expectedOutput: "ww",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := PadLeft(tc.s, tc.padChar, tc.desiredLength)
			if output != tc.expectedOutput {
				t.Errorf("output was not expected output: got - %s  -  expected - %s", output, tc.expectedOutput)
			}
		})
	}
}

func TestPadRight(t *testing.T) {
	type testCase struct {
		name           string
		s              string
		padChar        string
		desiredLength  int
		expectedOutput string
	}
	testCases := []testCase{
		{
			name:           "pad single digit with zeros",
			s:              "1",
			padChar:        "0",
			desiredLength:  2,
			expectedOutput: "10",
		},
		{
			name:           "2 char pad char string",
			s:              "w",
			padChar:        "12",
			desiredLength:  2,
			expectedOutput: "w12",
		},
		{
			name:           "string is already desired length",
			s:              "ww",
			padChar:        "1",
			desiredLength:  2,
			expectedOutput: "ww",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := PadRight(tc.s, tc.padChar, tc.desiredLength)
			if output != tc.expectedOutput {
				t.Errorf("output was not expected output: got - %s  -  expected - %s", output, tc.expectedOutput)
			}
		})
	}
}
