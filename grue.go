package main

import (
	"log"
)

const version = "0.1-alpha"

func main() {
	conf, err := readConfig("grue.cfg")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%+v\n", conf)
}
