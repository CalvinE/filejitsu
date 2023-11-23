package util

import (
	"slices"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func HideGlobalFlags(subCommand *cobra.Command, globalFlagsToHide []string) {
	subCommand.SetHelpFunc(func(c *cobra.Command, s []string) {
		c.Root().PersistentFlags().VisitAll(func(f *pflag.Flag) {
			if slices.Contains(globalFlagsToHide, f.Name) {
				f.Hidden = true
			}
		})
		c.Root().HelpFunc()(c, s)
	})
}
