package model

import (
	"testing"
)

func TestJson(t *testing.T) {
	type User struct {
		Name string `json:"name"`
	}
	response := NewTwockerResponse(200, []byte(`{"name":"John"}`), nil)
	user, err := TwockerJson[User](response)
	if err != nil {
		t.Errorf("Error unmarshalling JSON: %v", err)
	}
	if user.Name != "John" {
		t.Errorf("Expected name John, got %s", user.Name)
	}
}

func TestSelect(t *testing.T) {
	response := NewTwockerResponse(200, []byte(`<html><body><div class="test">Hello</div></body></html>`), nil)
	selection, err := response.Select(".test")
	if err != nil {
		t.Errorf("Error selecting element: %v", err)
	}
	if selection.Text() != "Hello" {
		t.Errorf("Expected text Hello, got %s", selection.Text())
	}
}
