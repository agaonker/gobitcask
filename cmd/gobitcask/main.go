package main

import (
	"fmt"
	"os"

	"github.com/ashish/gobitcask/cmd/gobitcask/commands"
)

func main() {
	rootCmd := commands.NewRootCommand()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
