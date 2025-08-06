package internal

import (
	"strings"
	"testing"
)

func TestNewGreeter(t *testing.T) {
	name := "World"
	greeter := NewGreeter(name)
	
	if greeter.GetName() != name {
		t.Errorf("Expected name %s, got %s", name, greeter.GetName())
	}
}

func TestGreet(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"World", "Hello, World! Welcome to Distributed Systems Patterns in Go."},
		{"Alice", "Hello, Alice! Welcome to Distributed Systems Patterns in Go."},
		{"", "Hello, ! Welcome to Distributed Systems Patterns in Go."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			greeter := NewGreeter(tt.name)
			result := greeter.Greet()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGreetWithTime(t *testing.T) {
	greeter := NewGreeter("World")
	result := greeter.GreetWithTime()
	
	if !strings.Contains(result, "Hello, World!") {
		t.Error("Expected greeting to contain 'Hello, World!'")
	}
	
	if !strings.Contains(result, "The time is") {
		t.Error("Expected greeting to contain time information")
	}
}

func TestSetName(t *testing.T) {
	greeter := NewGreeter("Initial")
	newName := "Updated"
	
	greeter.SetName(newName)
	
	if greeter.GetName() != newName {
		t.Errorf("Expected name %s, got %s", newName, greeter.GetName())
	}
}
