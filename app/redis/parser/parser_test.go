package parser

import (
	"fmt"
	"testing"
)

func TestGetAggregateLength(t *testing.T) {
	parser := &RedisParserImpl{}
	tests := []struct {
		input      []byte
		wantLength int
		wantEnd    int
		wantErr    bool
	}{
		// Valid aggregate length
		{[]byte("$5\r\nhello\r\n"), 5, 4, false},
		// Empty aggregate length
		{[]byte("$0\r\n\r\n"), 0, 4, false},
		// Valid multi digit length
		{[]byte("$15\r\nabcdefghij\r\n"), 15, 5, false},
		// Aggregate length with header only, no content
		{[]byte("$5\r\n"), 5, 4, false},
		// Invalid length (not a number)
		{[]byte("$notnumber\r\n"), -1, -1, true},
		// Negative length (unsupported)
		{[]byte("$-12\r\n"), -1, -1, true},
	}

	for _, tt := range tests {
		testName := fmt.Sprintf("input: %q", tt.input)
		t.Run(testName, func(t *testing.T) {
			gotLength, gotEnd, err := parser.GetAggregateLength(tt.input)
			gotErr := (err != nil)
			if gotErr != tt.wantErr {
				t.Fatalf("got error %v, want error %v, %v", gotErr, tt.wantErr, err)
			}
			if gotLength != tt.wantLength {
				t.Errorf("got length %d, want %d", gotLength, tt.wantLength)
			}
			if gotEnd != tt.wantEnd {
				t.Errorf("got end %d, want %d", gotEnd, tt.wantEnd)
			}
		})
	}
}

func TestParseBulkString(t *testing.T) {
	parser := &RedisParserImpl{}
	tests := []struct {
		input   []byte
		wantStr string
		wantEnd int
		wantErr bool
	}{
		// Valid bulk string
		{[]byte("$5\r\nhello\r\n"), "hello", 11, false},
		// Empty bulk string
		{[]byte("$0\r\n\r\n"), "", 6, false},
		// Valid bulk string with longer content
		{[]byte("$10\r\nabcdefghij\r\n"), "abcdefghij", 17, false},
		// Incomplete bulk string (header only)
		{[]byte("$5\r\n"), "", -1, true},
		// Incomplete bulk string (missing CRLF)
		{[]byte("$5\r\nhello"), "", -1, true},
		// Length mismatch: too short
		{[]byte("$4\r\nhel\r\n"), "", -1, true},
		// Length mismatch: too long
		{[]byte("$4\r\nhello\r\n"), "", -1, true},
		// Invalid type byte
		{[]byte("*1\r\nh\r\n"), "", -1, true},
	}

	for _, tt := range tests {
		testName := fmt.Sprintf("input: %q", tt.input)
		t.Run(testName, func(t *testing.T) {
			gotStr, gotEnd, err := parser.ParseBulkString(tt.input)
			gotErr := (err != nil)
			if gotErr != tt.wantErr {
				t.Fatalf("got error %v, want error %v, %v", gotErr, tt.wantErr, err)
			}
			if gotStr != tt.wantStr {
				t.Errorf("got string %q, want %q", gotStr, tt.wantStr)
			}
			if gotEnd != tt.wantEnd {
				t.Errorf("got end %d, want %d", gotEnd, tt.wantEnd)
			}
		})
	}
}

func TestParseArray(t *testing.T) {
	parser := &RedisParserImpl{}
	tests := []struct {
		input      []byte
		wantArr    []string
		wantLength int
		wantErr    bool
	}{
		// Valid array with multiple bulk strings
		{[]byte("*3\r\n$3\r\nfoo\r\n$3\r\nbar\r\n$3\r\nbaz\r\n"), []string{"foo", "bar", "baz"}, 31, false},
		// Empty array
		{[]byte("*0\r\n"), []string{}, 4, false},
		// Array with single element
		{[]byte("*1\r\n$5\r\nhello\r\n"), []string{"hello"}, 15, false},
		// Array with incomplete elements (fewer elements than declared)
		{[]byte("*3\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"), nil, -1, true},
		// Array with non-bulk string element (unsupported RESP type)
		{[]byte("*2\r\n$3\r\nfoo\r\n+OK\r\n"), nil, -1, true},
		// Invalid array type byte
		{[]byte("+3\r\n$3\r\nfoo\r\n"), nil, -1, true},
		// Array with mixed element sizes
		{[]byte("*3\r\n$1\r\na\r\n$8\r\nverylong\r\n$0\r\n\r\n"), []string{"a", "verylong", ""}, 31, false},
	}

	for _, tt := range tests {
		testName := fmt.Sprintf("input: %q", tt.input)
		t.Run(testName, func(t *testing.T) {
			gotArr, gotLength, err := parser.ParseArray(tt.input)
			gotErr := (err != nil)
			if gotErr != tt.wantErr {
				t.Fatalf("got error %v, want error %v, %v", gotErr, tt.wantErr, err)
			}
			if !equalStringSlices(gotArr, tt.wantArr) {
				t.Errorf("got array %v, want %v", gotArr, tt.wantArr)
			}
			if gotLength != tt.wantLength {
				t.Errorf("got length %d, want %d", gotLength, tt.wantLength)
			}
		})
	}
}

// Helper function to compare string slices
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
