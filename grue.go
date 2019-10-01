package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/c-14/grue/config"
)

const version = "0.2.0"

func usage() string {
	return `usage: grue [--help] {add|delete|fetch|import|init_cfg|list|rename} ...

Subcommands:
	add <name> <url>
	delete <name>
	fetch [-init] [name]
	import <config>
	init_cfg
	list [name] [--full]
	rename <old> <new>`
}

func add(args []string, conf *config.GrueConfig) error {
	if len(args) != 2 {
		return errors.New("usage: grue add <name> <url>")
	}
	var name string = args[0]
	var uri string = args[1]
	return conf.AddAccount(name, uri)
}

func del(args []string, conf *config.GrueConfig) error {
	if len(args) != 1 {
		return errors.New("usage: grue delete <name>")
	}
	name := args[0]
	if err := conf.DeleteAccount(name); err != nil {
		return err
	}
	return DeleteHistory(name)
}

func fetch(args []string, conf *config.GrueConfig) error {
	var initFlag bool
	fetchCmd := flag.NewFlagSet("fetch", flag.ContinueOnError)
	fetchCmd.BoolVar(&initFlag, "init", false, "Don't send emails, only initialize database of read entries")
	if err := fetchCmd.Parse(os.Args[2:]); err != nil {
		return err
	}
	if len(fetchCmd.Args()) == 0 {
		return fetchFeeds(conf, initFlag)
	}
	return fetchName(conf, fetchCmd.Arg(0), initFlag)
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
		var keys []string
		for k, _ := range conf.Accounts {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		if full {
			for _, k := range keys {
				fmt.Printf(fmtFull, k, conf.Accounts[k])
			}
			return nil
		}
		for _, k := range keys {
			fmt.Printf(fmtShort, k, conf.Accounts[k].URI)
		}
		return nil
	}

	name := listCmd.Args()[0]
	if cfg, ok := conf.Accounts[name]; ok {
		if full {
			fmt.Printf(fmtFull, name, cfg)
		} else {
			fmt.Printf(fmtShort, name, cfg.URI)
		}
	}
	return nil
}

func rename(args []string, conf *config.GrueConfig) error {
	if len(args) != 2 {
		return errors.New("usage: grue rename <old> <new>")
	}
	old := args[0]
	new := args[1]
	if err := conf.RenameAccount(old, new); err != nil {
		return err
	}
	return RenameHistory(old, new)
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
	case "delete":
		err = del(os.Args[2:], conf)
	case "fetch":
		err = fetch(os.Args[2:], conf)
	case "import":
		err = config.ImportCfg(os.Args[2:])
	case "init_cfg":
		break
	case "list":
		err = list(os.Args[2:], conf)
		break
	case "rename":
		err = rename(os.Args[2:], conf)
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
