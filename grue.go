package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/c-14/grue/config"
	"os"
	"text/tabwriter"
)

const version = "0.2.0-alpha"

func usage() string {
	return `usage: grue [--help] {add|fetch|import|init_cfg|list} ...

Subcommands:
	add <name> <url>
	fetch [-init]
	import <config>
	init_cfg
	list [name] [--full]`
}

func add(args []string, conf *config.GrueConfig) error {
	if len(args) != 2 {
		return errors.New("usage: grue add <name> <url>")
	}
	var name string = args[0]
	var uri string = args[1]
	return conf.AddAccount(name, uri)
}

func list(args []string, conf *config.GrueConfig) error {
	const (
		fmtShort = "%s\t%s\n"
		fmtFull  = "%s:\n%s\n"
	)
	var full bool
	var listCmd = flag.NewFlagSet("list", flag.ContinueOnError)
	listCmd.BoolVar(&full, "full", false, "Show full account info")
	if err := listCmd.Parse(args); err != nil {
		return err
	}
	if len(listCmd.Args()) == 0 {
		if full {
			for name, cfg := range conf.Accounts {
				fmt.Printf(fmtFull, name, cfg.String())
			}
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 8, 0, '\t', 0)
		for name, cfg := range conf.Accounts {
			fmt.Fprintf(w, fmtShort, name, cfg.URI)
		}
		w.Flush()
		return nil
	}

	name := listCmd.Args()[0]
	if cfg, ok := conf.Accounts[name]; ok {
		if full {
			fmt.Printf(fmtFull, name, cfg.String())
		} else {
			fmt.Printf(fmtShort, name, cfg.URI)
		}
	}
	return nil
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
	case "list":
		err = list(os.Args[2:], conf)
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
