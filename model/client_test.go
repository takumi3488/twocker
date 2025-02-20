package model

import (
	"testing"
)

func TestNewTwockerClientGet(t *testing.T) {
	c := NewTwockerClient()
	resp, err := c.Get("https://jsonplaceholder.typicode.com/todos")
	if err != nil {
		t.Errorf("Error making GET request: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}
}
