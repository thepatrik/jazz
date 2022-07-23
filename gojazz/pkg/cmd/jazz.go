package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thepatrik/jazz/gojazz/pkg/jazz"
	"github.com/thepatrik/strcolor"
)

const version = "0.0.1"

var jazzCmd = &cobra.Command{
	Use:   "jazz",
	Short: "jazz is a gas",
	Run: func(cmd *cobra.Command, _ []string) {
		file, err := cmd.Flags().GetString("file")
		if err != nil {
			fmt.Printf("could not read file flag %s\n", err)
			os.Exit(1)
		}

		if file != "" {
			info, err := os.Stat(file)
			if err != nil {
				fmt.Printf("could not read file %s\n", err)
				os.Exit(1)
			}

			if info.IsDir() {
				runFilesInDir(file)
			} else {
				runFile(file)
			}
		} else {
			repl()
		}
	},
}

func init() {
	jazzCmd.PersistentFlags().StringP("file", "f", "", "a file or a directory to parse.")
}

func Execute() {
	if err := jazzCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(interpreter *jazz.Interpreter, source string) error {
	scanner := jazz.NewScanner(source)
	tokens, err := scanner.ScanTokens()
	if err != nil {
		return err
	}

	parser := jazz.NewParser(tokens)
	stmts, err := parser.Parse()
	if err != nil || parser.HasErrors() {
		return err
	}

	resolver := jazz.NewResolver(interpreter)
	err = resolver.Resolve(stmts)
	if err != nil {
		return err
	}

	interpreter.Interpret(stmts)

	return nil
}

func runFile(file string) {
	b, err := os.ReadFile(file)
	if err != nil {
		fmt.Printf("could not read line %s", err)
		os.Exit(1)
	}

	interpreter := jazz.NewInterpreter()
	err = run(interpreter, string(b))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runFilesInDir(dir string) {
	files, err := filepath.Glob(filepath.Join(dir, "*.jz"))
	if err != nil {
		fmt.Printf("could not read files in %s\n", dir)
		os.Exit(1)
	}

	for _, file := range files {

		runFile(file)
	}
}

func repl() {
	interpreter := jazz.NewInterpreter(jazz.WithRepl(true))
	reader := bufio.NewReader(os.Stdin)

	fmt.Println(strcolor.Cyan(fmt.Sprintf("Welcome to Jazz %s\n", version)))

	for {
		fmt.Printf("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("could not read line %s\n", err)
			os.Exit(1)
		}

		line = strings.TrimSuffix(line, "\n")
		if line == ".exit" {
			break
		}

		err = run(interpreter, line)
		if err != nil {
			fmt.Println(err)
		}
	}
}
