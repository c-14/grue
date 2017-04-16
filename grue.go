package main

import (
	"errors"
	"fmt"
	"log"
	"os"
)

const version = "0.1-alpha"
const cfgPath = "grue.cfg"

func usage() string {
	return "usage: grue {add} ..."
}

func add(args []string, conf *config) error {
	if len(args) != 2 {
		return errors.New("usage: grue add <name> <url>")
	}
	var name string = args[0]
	var uri string = args[1]
	return conf.addAccount(name, uri)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, usage())
		os.Exit(EX_USAGE)
	}
	conf, err := readConfig(cfgPath)
	if err != nil {
		log.Fatal(err)
	}
	switch cmd := os.Args[1]; cmd {
	case "add":
		err = add(os.Args[2:], conf)
	case "import":
		err = importCfg(cfgPath, os.Args[2:])
	default:
		fmt.Fprintln(os.Stderr, usage())
		os.Exit(EX_USAGE)
	}
	if err != nil {
		log.Fatal(err)
	}
}
