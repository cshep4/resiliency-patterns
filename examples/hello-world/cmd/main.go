package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/cshep4/resiliency-patterns/examples/hello-world/internal"
)

func main() {
	var name string
	var withTime bool

	flag.StringVar(&name, "name", "World", "Name to greet")
	flag.BoolVar(&withTime, "time", false, "Include current time in greeting")
	flag.Parse()

	greeter := internal.NewGreeter(name)

	var greeting string
	if withTime {
		greeting = greeter.GreetWithTime()
	} else {
		greeting = greeter.Greet()
	}

	fmt.Println(greeting)

	if len(os.Args) > 1 && os.Args[1] == "demo" {
		fmt.Println("\n--- Demo Mode ---")
		fmt.Println("Demonstrating different greetings:")
		
		names := []string{"Alice", "Bob", "Charlie"}
		for _, n := range names {
			greeter.SetName(n)
			fmt.Printf("- %s\n", greeter.Greet())
		}
		
		fmt.Println("\nWith time:")
		greeter.SetName("Developer")
		fmt.Printf("- %s\n", greeter.GreetWithTime())
	}
}
