package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/ricci2511/riccis-homelab-utils/dupescout"
)

func main() {
	cfg := dupescout.Cfg{}

	flag.StringVar(&cfg.Path, "path", "", "path to search for duplicates")
	flag.BoolVar(&cfg.IgnoreHidden, "ignore-hidden", true, "ignore hidden files and directories")
	flag.Var(&cfg.ExtInclude, "include", "extensions to include")
	flag.Var(&cfg.ExtExclude, "exclude", "extensions to exclude")
	flag.Parse()

	// cfg.KeyGenerator = dupescout.FullHashKeyGenerator

	dupes := dupescout.Start(cfg)

	if len(dupes) == 0 {
		fmt.Printf("No duplicates found with the provided configuration: %s\n", cfg.String())
		os.Exit(0)
	}

	selectedDupes := []string{}

	prompt := &survey.MultiSelect{
		Message: "Select duplicates to delete:",
		Options: dupes,
	}

	err := survey.AskOne(prompt, &selectedDupes)
	if err != nil {
		log.Fatal(err)
	}

	for _, path := range selectedDupes {
		fmt.Printf("Deleting: %s\n", path)
		err := os.Remove(path)
		if err != nil {
			log.Fatal(err)
		}
	}
}
