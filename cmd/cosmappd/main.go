package main

import (
	"fmt"
	"os"

	"github.com/ciprianmuja/weight-shift/app"

	"github.com/ciprianmuja/weight-shift/cmd/cosmappd/cmd"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
)

func main() {
	rootCmd, _ := cmd.NewRootCmd()

	if err := svrcmd.Execute(rootCmd, "", app.DefaultNodeHome); err != nil {
		fmt.Fprintln(rootCmd.OutOrStderr(), err)
		os.Exit(1)
	}
}
