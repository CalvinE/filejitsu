package util

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type FlagModifier struct {
	Hide         bool
	UsagePrefix  string
	UsagePostFix string
}

type FlagModArgs map[string]FlagModifier

func HideGlobalFlags(subCommand *cobra.Command, globalFlagsToHide FlagModArgs) {
	subCommand.SetHelpFunc(func(c *cobra.Command, s []string) {
		c.Root().PersistentFlags().VisitAll(func(f *pflag.Flag) {
			m, ok := globalFlagsToHide[f.Name]
			if ok {
				if m.Hide {
					f.Hidden = true
					return
				}
				if len(m.UsagePrefix) > 0 {
					f.Usage = fmt.Sprintf("%s - %s", m.UsagePrefix, f.Usage)
				}
				if len(m.UsagePostFix) > 0 {
					f.Usage = fmt.Sprintf("%s - %s", f.Usage, m.UsagePostFix)
				}
			}
		})
		c.Root().HelpFunc()(c, s)
	})
}
