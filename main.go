package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	mcmaker "github.com/lack/mcmaker/pkg"
)

type cmdParser func([]string, mcmaker.McMaker) []string

func addFile(args []string, m mcmaker.McMaker) []string {
	c := flag.NewFlagSet("add", flag.ContinueOnError)
	fsource := c.String("source", "", "Local file source")
	fmode := c.Int("mode", 0644, "mode to create")
	fpath := c.String("path", "", "Path ans filename on destination host")
	c.Parse(args)
	m.AddFile(*fsource, *fpath, *fmode)
	return c.Args()
}

func addUnit(args []string, m mcmaker.McMaker) []string {
	c := flag.NewFlagSet("unit", flag.ContinueOnError)
	source := c.String("source", "", "The file source of the unit to add")
	name := c.String("name", "", "Unit name (defaults to basename of source)")
	enable := c.Bool("enable", true, "Should it be enabled")
	c.Parse(args)
	m.AddUnit(*source, *name, *enable)
	return c.Args()
}

func main() {
	name := flag.String("name", "", "The name of the object to create")
	stdout := flag.Bool("stdout", false, "If set, dump the object to stdout.  If not, creates a file called 'name.yaml' based on '--name'")
	flag.Parse()

	commands := map[string]cmdParser{
		"file": addFile,
		"unit": addUnit,
	}

	if *name == "" {
		fmt.Println("Name was not specified")
		flag.PrintDefaults()
		os.Exit(1)
	}

	var output io.Writer
	var err error
	if *stdout {
		output = os.Stdout
	} else {
		output, err = os.Create(fmt.Sprintf("%s.yaml", *name))
		if err != nil {
			panic(err)
		}
	}

	m := mcmaker.New(*name)
	remaining := flag.Args()
	for len(remaining) > 0 {
		handler, ok := commands[remaining[0]]
		if ok {
			remaining = handler(remaining[1:], m)
		} else {
			fmt.Printf("Supported commands:\n")
			for k := range commands {
				fmt.Println("    " + k)
			}
			os.Exit(1)
		}
	}
	m.WriteTo(output)
}
