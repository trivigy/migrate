package cmd

import (
	"bytes"

	"github.com/spf13/cobra"
)

func executeCommand(cmd *cobra.Command, args ...string) (output string, err error) {
	_, output, err = executeCommandC(cmd, args...)
	return output, err
}
func executeCommandC(cmd *cobra.Command, args ...string) (c *cobra.Command, output string, err error) {
	buf := new(bytes.Buffer)
	cmd.SetOutput(buf)
	cmd.SetArgs(args)
	c, err = cmd.ExecuteC()
	return c, buf.String(), err
}
