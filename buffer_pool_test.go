package twig

import (
	"strings"
	"testing"
)

func TestBufferPool(t *testing.T) {
	// Test getting a buffer from the pool
	buf := GetBuffer()
	if buf == nil {
		t.Fatal("GetBuffer() returned nil")
	}
	
	// Test writing to the buffer
	str := "Hello, world!"
	n, err := buf.WriteString(str)
	if err != nil {
		t.Fatalf("WriteString failed: %v", err)
	}
	if n != len(str) {
		t.Fatalf("WriteString returned wrong length: got %d, want %d", n, len(str))
	}
	
	// Test getting the string back
	if buf.String() != str {
		t.Fatalf("String() returned wrong value: got %q, want %q", buf.String(), str)
	}
	
	// Test resetting the buffer
	buf.Reset()
	if buf.Len() != 0 {
		t.Fatalf("Reset() didn't clear the buffer: length = %d", buf.Len())
	}
	
	// Test writing a different string after reset
	str2 := "Another string"
	buf.WriteString(str2)
	if buf.String() != str2 {
		t.Fatalf("String() after reset returned wrong value: got %q, want %q", buf.String(), str2)
	}
	
	// Test releasing the buffer
	buf.Release()
	
	// Getting a new buffer should not have any content
	buf2 := GetBuffer()
	if buf2.Len() != 0 {
		t.Fatalf("New buffer from pool has content: %q", buf2.String())
	}
	buf2.Release()
}

func TestWriteValue(t *testing.T) {
	buf := GetBuffer()
	defer buf.Release()
	
	tests := []struct {
		value    interface{}
		expected string
	}{
		{nil, ""},
		{"test", "test"},
		{123, "123"},
		{-456, "-456"},
		{3.14159, "3.14159"},
		{true, "true"},
		{false, "false"},
		{[]byte("bytes"), "bytes"},
	}
	
	for _, test := range tests {
		buf.Reset()
		_, err := WriteValue(buf, test.value)
		if err != nil {
			t.Errorf("WriteValue(%v) error: %v", test.value, err)
			continue
		}
		
		if buf.String() != test.expected {
			t.Errorf("WriteValue(%v) = %q, want %q", test.value, buf.String(), test.expected)
		}
	}
}

func TestWriteInt(t *testing.T) {
	buf := GetBuffer()
	defer buf.Release()
	
	tests := []struct {
		value    int
		expected string
	}{
		{0, "0"},
		{5, "5"},
		{-5, "-5"},
		{123, "123"},
		{-123, "-123"},
		{9999, "9999"},
		{-9999, "-9999"},
		{123456789, "123456789"},
		{-123456789, "-123456789"},
	}
	
	for _, test := range tests {
		buf.Reset()
		_, err := buf.WriteInt(test.value)
		if err != nil {
			t.Errorf("WriteInt(%d) error: %v", test.value, err)
			continue
		}
		
		if buf.String() != test.expected {
			t.Errorf("WriteInt(%d) = %q, want %q", test.value, buf.String(), test.expected)
		}
	}
}

func TestBufferWriteTo(t *testing.T) {
	buf := GetBuffer()
	defer buf.Release()
	
	testStr := "This is a test string"
	buf.WriteString(testStr)
	
	// Create a destination buffer to write to
	var dest strings.Builder
	
	n, err := buf.WriteTo(&dest)
	if err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}
	if n != int64(len(testStr)) {
		t.Fatalf("WriteTo returned wrong length: got %d, want %d", n, len(testStr))
	}
	
	if dest.String() != testStr {
		t.Fatalf("WriteTo output mismatch: got %q, want %q", dest.String(), testStr)
	}
}

func TestBufferGrowCapacity(t *testing.T) {
	buf := GetBuffer()
	defer buf.Release()
	
	// Start with small string
	initialStr := "small"
	buf.WriteString(initialStr)
	initialCapacity := cap(buf.buf)
	
	// Write a larger string that should cause a grow
	largeStr := strings.Repeat("abcdefghijklmnopqrstuvwxyz", 100) // 2600 bytes
	buf.WriteString(largeStr)
	
	// Verify capacity increased
	if cap(buf.buf) <= initialCapacity {
		t.Fatalf("Buffer didn't grow capacity: initial=%d, after=%d", 
			initialCapacity, cap(buf.buf))
	}
	
	// Verify content is correct
	expected := initialStr + largeStr
	if buf.String() != expected {
		t.Fatalf("Buffer content incorrect after growth")
	}
}