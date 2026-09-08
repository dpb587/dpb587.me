package main

import (
	"strings"
	"testing"

	"github.com/dpb587/tacitkb/ext/linkmirror"
)

func TestSnapshotLinkType(t *testing.T) {
	linkType, err := snapshotLinkType(&linkmirror.Snapshot{
		Referrer: "content/post/example",
		Target:   "https://example.test/article",
		Metadata: &linkmirror.Metadata{
			Name: "Example",
			FeaturedImageThumbnails: []linkmirror.Thumbnail{{
				URL: "/~/mirror-blob-iiif-image-v3/digest/full/240%2C120/0/default.jpg", Width: 240, Height: 120,
			}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if linkType.Metadata == nil || len(linkType.Metadata.FeaturedImageThumbnails) != 1 {
		t.Fatalf("link metadata = %#v", linkType.Metadata)
	}
	if !strings.Contains(linkType.Metadata.FeaturedImageThumbnails[0].URL, "/~/mirror-blob-iiif-image-v3/") {
		t.Fatalf("thumbnail URL = %q", linkType.Metadata.FeaturedImageThumbnails[0].URL)
	}
}
