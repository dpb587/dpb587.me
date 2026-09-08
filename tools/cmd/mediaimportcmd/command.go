package mediaimportcmd

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"bytes"

	"github.com/dpb587/dpb587.me/tools/cmd/cmdflags"
	"github.com/dpb587/dpb587.me/tools/content/storybuilder"
	"github.com/dpb587/dpb587.me/tools/util/nanoidutil"
	"github.com/dpb587/tacitkb/catalog"
	"github.com/dpb587/tacitkb/ext/blob"
	"github.com/spf13/cobra"
)

// New creates the "media-import" command for embedding one-off files (e.g. a photo or video
// for a single post) as media-*.md sub-content of an existing content directory, as opposed to
// "story-import" which mirrors an entire source tree into content/story.
func New(cGlobal *cmdflags.Global) *cobra.Command {
	var fContentDir string

	cmd := &cobra.Command{
		Use:  "media-import FILE...",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if fContentDir == "" {
				return errors.New("--content-dir is required")
			}

			ctx := cmd.Context()
			builder := storybuilder.NewService(cGlobal.Log, cGlobal.Repository, cGlobal.ReverseGeo)

			for _, arg := range args {
				err := func() error {
					blobContentLocation, err := cGlobal.FileService.NewResourceContent(arg)
					if err != nil {
						return fmt.Errorf("get blob content: %v", err)
					}

					blobProfile, err := cGlobal.BlobService.BuildNodeProfile(ctx, blobContentLocation)
					if err != nil {
						return fmt.Errorf("build blob profile: %v", err)
					}

					//

					blobNode, err := cGlobal.Repository.EvaluateNode(ctx, blobProfile.Descriptor, catalog.RepositoryEvaluateNodeConfig{
						Labels: blobProfile.Labels,
					})
					if err != nil {
						return fmt.Errorf("evaluate blob: %v", err)
					}

					//

					err = cGlobal.Repository.PutResource(ctx, blobNode.UID(), blob.ContentResourceDescriptor{}, blobContentLocation)
					if err != nil {
						return fmt.Errorf("put blob content: %w", err)
					}

					//

					generatedContent, err := builder.Build(ctx, blobNode)
					if err != nil {
						return fmt.Errorf("build: %w", err)
					} else if generatedContent == nil {
						return errors.New("no content generated")
					}

					//

					decodedSha256, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(blobProfile.Descriptor.(blob.ObjectNodeDescriptor).Key, "sha-256="))
					if err != nil {
						return fmt.Errorf("decode sha256: %w", err)
					}

					contentBaseName := nanoidutil.DeterministicSimplex(bytes.NewReader(decodedSha256)) + ".md"

					switch *generatedContent.Frontmatter.Type {
					case "media":
						contentBaseName = "media-" + contentBaseName
					case "route":
						contentBaseName = "activity-" + contentBaseName
					default:
						return fmt.Errorf("unknown layout: %q", *generatedContent.Frontmatter.Layout)
					}

					contentPath := filepath.Join(fContentDir, contentBaseName)

					err = os.MkdirAll(filepath.Dir(contentPath), 0700)
					if err != nil {
						return fmt.Errorf("create content directory %q: %w", filepath.Dir(contentPath), err)
					}

					fh, err := os.OpenFile(contentPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
					if err != nil {
						return fmt.Errorf("open content file %q: %w", contentPath, err)
					}

					defer fh.Close()

					_, err = generatedContent.WriteTo(fh)
					if err != nil {
						return fmt.Errorf("write to stdout: %w", err)
					}

					cGlobal.Log.Info(
						"wrote content",
						"content", contentPath,
					)

					return nil
				}()
				if err != nil {
					cGlobal.Log.Error(
						"import failed",
						"blob", arg,
						"error", err,
					)

					continue
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&fContentDir, "content-dir", fContentDir, "")

	return cmd
}
