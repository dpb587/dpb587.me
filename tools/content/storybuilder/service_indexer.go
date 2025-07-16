package storybuilder

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dpb587/dpb587.me/tools/content"
	"github.com/dpb587/dpb587.me/tools/content/frontmatterparams"
	"github.com/dpb587/dpb587.me/tools/content/hugoutil"
	"github.com/dpb587/tacitkb/util/ptrutil"
)

func (s *Service) FillSections(ctx context.Context, rootStoryDir string) error {
	contentByDir := map[string]map[string]*content.Document{}
	indexByDir := map[string]*content.Document{}

	err := filepath.WalkDir(rootStoryDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		} else if strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		} else if d.IsDir() {
			return nil
		}

		fh, err := os.OpenFile(path, os.O_RDONLY, 0)
		if err != nil {
			return fmt.Errorf("open: %v", err)
		}

		defer fh.Close()

		doc := &content.Document{}

		_, err = doc.ReadFrom(fh)
		if err != nil {
			return fmt.Errorf("read content: %v", err)
		}

		if filepath.Base(path) == "_index.md" {
			indexByDir[filepath.Dir(path)] = doc
		} else {
			if contentByDir[filepath.Dir(path)] == nil {
				contentByDir[filepath.Dir(path)] = map[string]*content.Document{}
			}

			contentByDir[filepath.Dir(path)][filepath.Base(path)] = doc
		}

		return nil
	})
	if err != nil {
		return err
	}

	{
		// fill in any parents which didn't have content of their own
		baseCount := len(strings.Split(rootStoryDir, string(filepath.Separator))) + 1

		for dir := range contentByDir {
			dirSegments := strings.Split(dir, string(filepath.Separator))

			for idx := range dirSegments {
				if idx <= baseCount {
					continue
				}

				parentDir := filepath.Join(dirSegments[:idx]...)

				if _, known := contentByDir[parentDir]; !known {
					contentByDir[parentDir] = map[string]*content.Document{}
				}
			}
		}
	}

	completeDirs := map[string]struct{}{}

	for len(contentByDir) != len(completeDirs) {
		for dir := range contentByDir {
			if _, known := completeDirs[dir]; known {
				continue
			} else if _, known := indexByDir[dir]; known {
				completeDirs[dir] = struct{}{}

				continue
			}

			var delay bool

			for maybeSubdir := range contentByDir {
				if strings.HasPrefix(maybeSubdir, dir+"/") {
					if _, done := completeDirs[maybeSubdir]; !done {
						delay = true

						break
					}
				}
			}

			if delay {
				continue
			}

			err := func() error {
				content := &content.Document{
					Frontmatter: &content.Content_Frontmatter{
						Params: &content.Content_Frontmatter_Params{},
						Title:  ptrutil.Value(strings.Title(strings.ReplaceAll(regexp.MustCompile(`^20\d\d\-`).ReplaceAllString(filepath.Base(dir), ""), "-", " "))),
					},
				}

				var minTime, maxTime *hugoutil.FrontmatterTime

				for contentDir, contentByBase := range contentByDir {
					if contentDir != dir && !strings.HasPrefix(contentDir, dir+"/") {
						continue
					}

					for _, contentItem := range contentByBase {
						if contentItem.Frontmatter.Params.Nav != nil {
							if contentItem.Frontmatter.Params.Nav.Place != nil {
								for k := range *contentItem.Frontmatter.Params.Nav.Place {
									content.Frontmatter.Params.SetNavPlaceArea(k, false)
								}
							}

							if contentItem.Frontmatter.Params.Nav.PlacePark != nil {
								for k := range *contentItem.Frontmatter.Params.Nav.PlacePark {
									content.Frontmatter.Params.SetNavPlacePark(k, false)
								}
							}
						}

						if contentItem.Frontmatter.Params.TimeRange != nil {
							if f := contentItem.Frontmatter.Params.TimeRange.From; minTime == nil || f.Time().Before(minTime.Time()) {
								minTime = f
							}

							if f := contentItem.Frontmatter.Params.TimeRange.Thru; maxTime == nil || f.Time().After(maxTime.Time()) {
								maxTime = f
							}
						} else if contentItem.Frontmatter.Date != nil {
							if minTime == nil || contentItem.Frontmatter.Date.Time().Before(minTime.Time()) {
								minTime = contentItem.Frontmatter.Date
							}

							if maxTime == nil || contentItem.Frontmatter.Date.Time().After(maxTime.Time()) {
								maxTime = contentItem.Frontmatter.Date
							}
						}
					}
				}

				if minTime != nil && maxTime != nil {
					if minTime.Time().Equal(maxTime.Time()) {
						content.Frontmatter.Date = minTime
					} else {
						content.Frontmatter.Date = minTime
						content.Frontmatter.Params.TimeRange = &frontmatterparams.TimeRange{
							From: minTime,
							Thru: maxTime,
						}
					}
				}

				fh, err := os.OpenFile(filepath.Join(dir, "_index.md"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
				if err != nil {
					return fmt.Errorf("open content file: %v", err)
				}

				defer fh.Close()

				_, err = content.WriteTo(fh)
				if err != nil {
					return fmt.Errorf("write content: %v", err)
				}

				s.log.Info(
					"wrote content",
					"content", filepath.Join(dir, "_index.md"),
				)

				return nil
			}()
			if err != nil {
				return err
			}

			completeDirs[dir] = struct{}{}
		}
	}

	return nil
}
