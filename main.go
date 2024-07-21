package main

import (
	"errors"
	"log"

	"github.com/Drakmyth/golang-zmachine/zmachine"
	"github.com/spf13/cobra"
)

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

		err := zmachine.Load(args[0])
		if err != nil {
			panic(err)
		}
	},
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
