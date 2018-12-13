package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"strings"
)

const (
	versionHelp = `print contacts version.`
	unreleased  = "unreleased"
)

var version string

func newVersionCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: versionHelp,
		Long:  versionHelp,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(out, ver())
		},
	}
	return cmd
}

func ver() string {
	if v := strings.TrimSpace(version); v != "" {
		return v
	} else {
		return unreleased
	}
}