package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/Drakmyth/golang-zmachine/zmachine"
	"github.com/spf13/cobra"
)

var debug bool

func init() {
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Print execution instructions")
}

var rootCmd = &cobra.Command{
	Use:   "zmachine <story-file-path>",
	Short: "Play a Z-Machine story file",
	Long: `Load the provided Z-Machine story file into memory and begin interpreter execution.
In other words, load and play the game!`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires positional parameter")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		interpreter, err := zmachine.Load(args[0])
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}

		interpreter.Debug = debug

		err = interpreter.Run()
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			interpreter.Screen.End()
			os.Exit(1)
		}
	},
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}
