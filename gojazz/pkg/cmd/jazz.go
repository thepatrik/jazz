package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/peterh/liner"
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

func run(interpreter *jazz.Interpreter, source string, repl bool) error {
	scanner := jazz.NewScanner(source)
	tokens, err := scanner.ScanTokens()
	if err != nil {
		return err
	}

	parser := jazz.NewParser(tokens, repl)
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
	err = run(interpreter, string(b), false)
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

// historyPath returns the shared REPL history file (~/.jazz_history, same
// filename as the cjazz linenoise REPL) or "" if $HOME is unset.
func historyPath() string {
	home := os.Getenv("HOME")
	if home == "" {
		return ""
	}
	return filepath.Join(home, ".jazz_history")
}

func repl() {
	interpreter := jazz.NewInterpreter(jazz.WithRepl(true))

	// liner is the pure-Go linenoise port: left/right cursor, Home/End,
	// backspace, and up/down history. It also detects non-TTY (piped) stdin
	// and falls back to a plain buffered line read, returning io.EOF at end.
	state := liner.NewLiner()
	defer state.Close()
	state.SetCtrlCAborts(true)

	histPath := historyPath()
	if histPath != "" {
		if f, err := os.Open(histPath); err == nil {
			_, _ = state.ReadHistory(f)
			f.Close()
		}
	}

	fmt.Println(strcolor.BrightCyan(fmt.Sprintf("Welcome to Jazz v%s", version)))
	fmt.Println(strcolor.Cyan("Type \".exit\" to exit."))

	for {
		line, err := state.Prompt("> ")
		if err == liner.ErrPromptAborted || err == io.EOF {
			// Ctrl-C (abort) or Ctrl-D / piped EOF: exit cleanly.
			fmt.Println()
			break
		}
		if err != nil {
			fmt.Printf("could not read line %s\n", err)
			break
		}

		if line == ".exit" {
			break
		}

		if line != "" {
			state.AppendHistory(line)
			if histPath != "" {
				if f, err := os.Create(histPath); err == nil {
					_, _ = state.WriteHistory(f)
					f.Close()
				}
			}
		}

		if err := run(interpreter, line, true); err != nil {
			fmt.Println(err)
		}
	}
}
