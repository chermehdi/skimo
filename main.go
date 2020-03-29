package main

import (
	"flag"
	"fmt"
	"os"
	"path"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "~/"
	}
	includeDir := flag.String("include_dir", path.Join(home, "include"), "The path to your library")
	verbose := flag.Bool("verbose", false, "Print debug log")
	flag.Parse()

	inliner, err := NewInliner(*includeDir, *verbose, []string{""})
	if err != nil {
		handleError(err)
	}
	content, err := inliner.Inline(os.Stdin)
	if err != nil {
		handleError(err)
	}
	fmt.Print(content)
}

func handleError(err error) {
	formattedError := fmt.Errorf("Error produced when creating inliner %s\n", err.Error())
	fmt.Print(formattedError)
	os.Exit(1)
}
