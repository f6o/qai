package cmd

import (
	"github.com/f6o/qai/i18n"
	"github.com/f6o/qai/internal/skill"
	"github.com/spf13/cobra"
)

var skillsCmd = &cobra.Command{
	Use:   "skills",
	Short: i18n.T("cmd.skills.short"),
	Long: `Print the qai skill: the usage contract for AI agents, as markdown.
This is the same document installed as a skill file; the YAML front matter is omitted.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Print(skill.Body())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(skillsCmd)
}
