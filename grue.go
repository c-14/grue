package main

import (
	"c-14/grue/config"
	"errors"
	"fmt"
	"log"
	"os"
)

const version = "0.1-alpha"

func usage() string {
	return "usage: grue {add|fetch|import} ..."
}

func add(args []string, conf *config.GrueConfig) error {
	if len(args) != 2 {
		return errors.New("usage: grue add <name> <url>")
	}
	var name string = args[0]
	var uri string = args[1]
	return conf.AddAccount(name, uri)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, usage())
		os.Exit(EX_USAGE)
	}
	conf, err := config.ReadConfig()
	if err != nil {
		log.Fatal(err)
	}
	switch cmd := os.Args[1]; cmd {
	case "add":
		err = add(os.Args[2:], conf)
	case "fetch":
		var hasError bool = false
		ret := make(chan error)
		go fetchFeeds(ret, conf)
		for r := range ret {
			if r != nil {
				log.Println(r)
				hasError = true
			}
		}
		if hasError {
			os.Exit(EX_SOFTWARE)
		}
	case "import":
		err = config.ImportCfg(os.Args[2:])
	default:
		fmt.Fprintln(os.Stderr, usage())
		os.Exit(EX_USAGE)
	}
	if err != nil {
		log.Fatal(err)
	}
}
