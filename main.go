package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	mcmaker "github.com/lack/mcmaker/pkg"
)

type cmdParser func([]string, mcmaker.McMaker) ([]string, error)

func addFile(args []string, m mcmaker.McMaker) ([]string, error) {
	c := flag.NewFlagSet("file", flag.ExitOnError)
	c.Usage = func() {
		o := flag.CommandLine.Output()
		fmt.Fprintf(o, "Adds a file to the MachineConfig object\n\nUsage:\n  %s ... file [options] ...\n\nOptions:\n", os.Args[0])
		c.PrintDefaults()
	}
	source := c.String("source", "", "The local file containing the file data")
	path := c.String("path", "", "Path and filename to create on the destination host")
	mode := c.Int("mode", 0644, "mode to create")
	err := c.Parse(args)
	if err != nil {
		return nil, err
	}
	err = m.AddFile(*source, *path, *mode)
	if err != nil {
		return nil, err
	}
	return c.Args(), nil
}

func addUnit(args []string, m mcmaker.McMaker) ([]string, error) {
	c := flag.NewFlagSet("unit", flag.ExitOnError)
	c.Usage = func() {
		o := flag.CommandLine.Output()
		fmt.Fprintf(o, "Adds a systemd unit to the MachineConfig object\n\nUsage:\n  %s ... unit [options] ...\n\nOptions:\n", os.Args[0])
		c.PrintDefaults()
	}
	source := c.String("source", "", "The local file containing the unit definition")
	name := c.String("name", "", "Unit name to create (defaults to basename of source)")
	enable := c.Bool("enable", true, "Should it be enabled")
	err := c.Parse(args)
	if err != nil {
		return nil, err
	}
	err = m.AddUnit(*source, *name, *enable)
	if err != nil {
		return nil, err
	}
	return c.Args(), nil
}

func main() {
	commands := map[string]cmdParser{
		"file": addFile,
		"unit": addUnit,
	}

	flag.Usage = func() {
		o := flag.CommandLine.Output()
		fmt.Fprintf(o, "Creates a MachineConfig object with custom contents\n\nUsage:\n  %s [options] [commands...]\n\nOptions:\n", os.Args[0])
		flag.CommandLine.PrintDefaults()
		fmt.Fprintf(o, "\nCommands:\n")
		for k := range commands {
			fmt.Fprintf(o, "  %s\n", k)
		}
		fmt.Fprintf(o, "\nRun %s -name foo [command] -help for details on each specific command\n", os.Args[0])
	}
	name := flag.String("name", "", "The name of the MC object to create")
	stdout := flag.Bool("stdout", false, "If set, dump the object to stdout.  If not, creates a file called 'name.yaml' based on '-name'")
	flag.Parse()

	if *name == "" {
		fmt.Fprintf(flag.CommandLine.Output(), "No -name was specified\n\n")
		flag.Usage()
		os.Exit(1)
	}

	m := mcmaker.New(*name)
	remaining := flag.Args()
	var err error
	for len(remaining) > 0 {
		handler, ok := commands[remaining[0]]
		if ok {
			remaining, err = handler(remaining[1:], m)
			if err != nil {
				panic(err)
			}
		} else {
			fmt.Fprintf(flag.CommandLine.Output(), "Unrecognized command %q\n\n", remaining[0])
			flag.Usage()
			os.Exit(1)
		}
	}

	var output io.Writer
	if *stdout {
		output = os.Stdout
	} else {
		output, err = os.Create(fmt.Sprintf("%s.yaml", *name))
		if err != nil {
			panic(err)
		}
	}
	m.WriteTo(output)
}
