package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/c-14/grue/config"
	"os"
)

const version = "0.2.0-alpha"

func usage() string {
	return `usage: grue [--help] {add|fetch|import|init_cfg} ...

Subcommands:
	add <name> <url>
	fetch [-init]
	import <config>
	init_cfg`
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
		fmt.Fprintln(os.Stderr, err)
		os.Exit(EX_TEMPFAIL)
	}
	defer conf.Unlock()
	switch cmd := os.Args[1]; cmd {
	case "add":
		err = add(os.Args[2:], conf)
	case "fetch":
		var fetchCmd = flag.NewFlagSet("fetch", flag.ContinueOnError)
		var initFlag = fetchCmd.Bool("init", false, "Don't send emails, only initialize database of read entries")
		err = fetchCmd.Parse(os.Args[2:])
		if err == nil {
			err = fetchFeeds(conf, *initFlag)
		}
	case "import":
		err = config.ImportCfg(os.Args[2:])
	case "init_cfg":
		break
	case "-v":
		fallthrough
	case "--version":
		fmt.Println(version)
	case "-h":
		fallthrough
	case "--help":
		fmt.Println(usage())
	default:
		fmt.Fprintln(os.Stderr, usage())
		conf.Unlock()
		os.Exit(EX_USAGE)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
