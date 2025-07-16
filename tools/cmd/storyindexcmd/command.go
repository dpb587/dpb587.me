package storyindexcmd

import (
	"fmt"
	"path/filepath"

	"github.com/dpb587/dpb587.me/tools/cmd/cmdflags"
	"github.com/dpb587/dpb587.me/tools/content/storybuilder"
	"github.com/spf13/cobra"
)

func New(cGlobal *cmdflags.Global) *cobra.Command {
	var fContentStoryDir string

	cmd := &cobra.Command{
		Use: "story-index [DIR...]",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			builder := storybuilder.NewService(cGlobal.Log, cGlobal.Repository, cGlobal.ReverseGeo)

			if len(args) == 0 {
				args = []string{""}
			}

			for _, arg := range args {
				err := builder.FillSections(ctx, filepath.Join(fContentStoryDir, arg))
				if err != nil {
					return fmt.Errorf("fill sections: %w", err)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&fContentStoryDir, "content-story-dir", fContentStoryDir, "")

	return cmd
}
