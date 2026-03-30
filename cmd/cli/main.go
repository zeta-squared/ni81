package main

import (
	"fmt"
	"log"
	"ni81/project"
	"os"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] == "--help" {
		logUsage()
		os.Exit(0)
	}

	// TODO: use `flags` package
	switch os.Args[1] {
	case "init":
		err := project.Initialise()
		if err != nil {
			log.Fatalln(err)
		}
	case "cache":
		proj, err := project.NewProject("ni81.toml")
		if err != nil {
			log.Fatalln(err)
		}

		err = proj.CreateCache()
		if err != nil {
			log.Fatalln(err)
		}
	case "translate":
		proj, err := project.NewProject("ni81.toml")
		if err != nil {
			log.Fatalln(err)
		}

		err = proj.Translate()
		if err != nil {
			log.Fatalln(err)
		}
	default:
		logUsage()
		os.Exit(0)
	}
}

func logUsage() {
	fmt.Println("Manage your project's i18n by nibbling away at it")
	fmt.Println()
	fmt.Println("USAGE: ni81 <command>")
	fmt.Println()
	fmt.Println("COMMANDS")
	fmt.Println("  init\t\tInitialize a new project")
	fmt.Println("  cache\t\tCreate a cache for the project")
	fmt.Println("  translate\tTranslate files in the project")
}
