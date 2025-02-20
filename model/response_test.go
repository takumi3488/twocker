package model

import (
	"testing"
)

func TestJson(t *testing.T) {
	type User struct {
		Name string `json:"name"`
	}
	response := NewTwockerResponse(200, []byte(`{"name":"John"}`))
	user, err := TwockerJson[User](response)
	if err != nil {
		t.Errorf("Error unmarshalling JSON: %v", err)
	}
	if user.Name != "John" {
		t.Errorf("Expected name John, got %s", user.Name)
	}
}
