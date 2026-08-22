package main

import (
	"log"

	"github.com/hafidluqman50/maoi/src/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
