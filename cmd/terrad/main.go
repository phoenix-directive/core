package main

import (
	"fmt"
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	terraapp "github.com/terra-money/core/v2/app"
)

func main() {
	rootCmd, _ := NewRootCmd()

	if err := svrcmd.Execute(rootCmd, "", terraapp.DefaultNodeHome); err != nil {
		fmt.Fprintln(rootCmd.OutOrStdout(), err)
		os.Exit(1)
	}
}
