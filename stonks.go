package main

import (
	"log"

	"github.com/awoodbeck/faang-stonks/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
