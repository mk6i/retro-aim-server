// This program generates env config scripts from config.Config struct tags for
// unix and windows platforms.
// Usage: go run ./cmd/config_generator [platform] [filename]
// Example: go run ./cmd/config_generator unix settings.env
package main

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/mk6i/retro-aim-server/config"

	"github.com/mitchellh/go-wordwrap"
)

var platformKeywords = map[string]struct {
	comment    string
	assignment string
}{
	"windows": {
		comment:    "rem ",
		assignment: "set ",
	},
	"unix": {
		comment:    "# ",
		assignment: "export ",
	},
}

func main() {
	args := os.Args[1:]
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run cmd/config_generator [platform] [filename]")
		os.Exit(1)
	}

	keywords, ok := platformKeywords[args[0]]
	if !ok {
		fmt.Fprintf(os.Stderr, "unable to find platform `%s`\n", os.Args[1])
		os.Exit(1)
	}
	fmt.Println("writing to", args[1])
	f, err := os.Create(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating file: %s\n", err.Error())
		os.Exit(1)
	}
	defer f.Close()

	configType := reflect.TypeOf(config.Config{})
	for i := 0; i < configType.NumField(); i++ {
		field := configType.Field(i)
		comment := field.Tag.Get("description")
		if err := writeComment(f, comment, 80, keywords.comment); err != nil {
			fmt.Fprintf(os.Stderr, "error writing to file: %s\n", err.Error())
			os.Exit(1)
		}

		varName := field.Tag.Get("envconfig")
		val := field.Tag.Get("val")
		if err := writeAssignment(f, keywords.assignment, varName, val); err != nil {
			fmt.Fprintf(os.Stderr, "error writing to file: %s\n", err.Error())
			os.Exit(1)
		}
	}
}

func writeComment(w io.Writer, comment string, width uint, keyword string) error {
	// adjust wrapping threshold to accommodate comment keyword length
	width = width - uint(len(keyword))
	comment = wordwrap.WrapString(comment, width)
	// prepend lines with comment keyword
	comment = strings.ReplaceAll(comment, "\n", fmt.Sprintf("\n%s", keyword))
	_, err := fmt.Fprintf(w, "%s%s\n", keyword, comment)
	return err
}

func writeAssignment(w io.Writer, keyword string, varName string, val string) error {
	_, err := fmt.Fprintf(w, "%s%s=%s\n\n", keyword, varName, val)
	return err
}
