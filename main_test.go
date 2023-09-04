package main

import (
	"fmt"
	"regexp"
	"testing"
)

func TestReplace(t *testing.T) {
	key := "videos/demo.mp4"
	re := regexp.MustCompile("\\.[^.]+$")
	resultKey := re.ReplaceAllString(key, fmt.Sprintf(".%s", "jpg"))
	t.Log(resultKey)
}

func TestAssembleResultKey(t *testing.T) {
	Extension = "jpg"
	tests := []string{
		"u/ac/b2c1841b-8d56-4cba-8543-ba0e01226f3f.mp4",
		"u/d3/9f438a00-d37c-4af9-a816-d0fd7c744605.jpg",
	}
	for _, key := range tests {
		resultKey := assembleResultKey(key)
		t.Log(key, resultKey)
	}
}
