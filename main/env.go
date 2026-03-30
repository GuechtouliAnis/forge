package main

import (
	"fmt"

	"github.com/GuechtouliAnis/forge/internal"
)

func main() {
	line, err := internal.ParseEnv(".env")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}

	err = internal.WriteEnvExample(".env.example", line)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}

}
