package internal

import (
	"fmt"
	"time"
)

type Greeter struct {
	name string
}

func NewGreeter(name string) *Greeter {
	return &Greeter{name: name}
}

func (g *Greeter) Greet() string {
	return fmt.Sprintf("Hello, %s! Welcome to Distributed Systems Patterns in Go.", g.name)
}

func (g *Greeter) GreetWithTime() string {
	now := time.Now().Format("15:04:05")
	return fmt.Sprintf("Hello, %s! The time is %s. Welcome to Distributed Systems Patterns in Go.", g.name, now)
}

func (g *Greeter) SetName(name string) {
	g.name = name
}

func (g *Greeter) GetName() string {
	return g.name
}
