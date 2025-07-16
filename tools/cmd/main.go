package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/dpb587/dpb587.me/tools/cmd/cmdflags"
	"github.com/dpb587/dpb587.me/tools/cmd/storyimportcmd"
	"github.com/dpb587/dpb587.me/tools/cmd/storyindexcmd"
	"github.com/dpb587/go-iiif-image-api-v3/imagerequest"
	"github.com/dpb587/tacitkb/catalog"
	"github.com/dpb587/tacitkb/catalog/catalogmeta/embed"
	"github.com/dpb587/tacitkb/catalog/storage/ndjson"
	storageos "github.com/dpb587/tacitkb/catalog/storage/os"
	"github.com/dpb587/tacitkb/ext/blob"
	blobprofilebuilder "github.com/dpb587/tacitkb/ext/blob/profilebuilder"
	"github.com/dpb587/tacitkb/ext/blobmetaexif"
	blobmetaexifprofilebuilder "github.com/dpb587/tacitkb/ext/blobmetaexif/profilebuilder"
	"github.com/dpb587/tacitkb/ext/blobtypeimage"
	blobtypeimageprofilebuilder "github.com/dpb587/tacitkb/ext/blobtypeimage/profilebuilder"
	"github.com/dpb587/tacitkb/ext/exportgeojson"
	exportgeojsonartifactsbuilder "github.com/dpb587/tacitkb/ext/exportgeojson/artifactsbuilder"
	"github.com/dpb587/tacitkb/ext/exportiiifimage3"
	exportiiifimage3artifactsbuilder "github.com/dpb587/tacitkb/ext/exportiiifimage3/artifactsbuilder"
	"github.com/dpb587/tacitkb/tools/googlemapsreversegeocode"
	"github.com/dpb587/tacitkb/tools/kvcache/sqlite3"
	"github.com/dpb587/tacitkb/tools/locationmask"
	"github.com/spf13/cobra"
	"googlemaps.github.io/maps"
)

func main() {
	if err := mainErr(); err != nil {
		panic(err)
	}
}

func mainErr() error {
	cGlobal := &cmdflags.Global{
		Log: slog.Default(),
	}

	{
		fileService := storageos.NewService(storageos.ServiceOptions{
			MountsByVirtual: map[string]string{
				"file:///mnt/depot/": "/workspaces/dpb587.me/tmp/mnt/",
			},
		})

		cGlobal.FileService = fileService
	}

	cGlobal.BlobService = blob.NewService()

	{
		mapsClient, err := maps.NewClient(maps.WithAPIKey(os.Getenv("GOOGLE_API_KEY")), maps.WithRateLimit(2))
		if err != nil {
			panic(fmt.Errorf("client: %v", err))
		}

		reversegeocodeCache, err := sqlite3.OpenService[googlemapsreversegeocode.LookupLocationInput, *googlemapsreversegeocode.LookupLocationOutput](
			"/workspaces/dpb587.me/tmp/googlemapsreversegeocode-cache.sqlite3",
		)
		if err != nil {
			panic(fmt.Errorf("cache: %v", err))
		}

		cGlobal.ReverseGeo = googlemapsreversegeocode.NewService(cGlobal.Log, mapsClient, reversegeocodeCache)
	}

	//

	{
		profiles, err := locationmask.LoadGeoJSON("/workspaces/dpb587.me/private/locationmask/*.geojson")
		if err != nil {
			panic(fmt.Errorf("load location mask profiles: %w", err))
		}

		cGlobal.Log.Info("loaded location mask profiles", "count", len(profiles))

		locationmaskService := locationmask.NewService(profiles)

		//

		blobIdentifierNamer := func(ctx context.Context, repository catalog.Repository, node catalog.Node) (string, error) {
			raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(node.Descriptor().(blob.ObjectNodeDescriptor).Key, "sha-256="))
			if err != nil {
				return "", fmt.Errorf("decode key: %w", err)
			}

			return fmt.Sprintf("%x", raw), nil
		}

		repository := ndjson.NewRepository(ndjson.RepositoryOptions{
			BaseDir:  "/workspaces/dpb587.me/private/data/",
			Compress: true,
			NodeDescriptorFactory: catalog.NodeDescriptorFactoryList{
				cGlobal.BlobService,
			},
			NodeLabelFactory: catalog.NodeLabelFactoryList{
				cGlobal.BlobService,
			},
			ResourceDescriptorFactory: catalog.ResourceDescriptorFactoryList{
				cGlobal.BlobService,
				&blobmetaexif.Factory{},
				&blobtypeimage.Factory{},
				&exportiiifimage3.Factory{},
				&exportgeojson.Factory{},
			},
			ResourceContentFactory: catalog.ResourceContentFactoryList{
				cGlobal.FileService,
				&embed.Factory{},
			},
			ResourceBuilder: catalog.ResourceBuilderList{
				blobmetaexifprofilebuilder.NewBuilder(cGlobal.Log, locationmaskService),
				blobprofilebuilder.NewBuilder(cGlobal.Log),
				blobtypeimageprofilebuilder.NewBuilder(cGlobal.Log),
				exportiiifimage3artifactsbuilder.NewBuilder(cGlobal.Log, []exportiiifimage3artifactsbuilder.BuilderProfile{
					{
						Name:            "default",
						BaseURL:         "/~/blob-iiif-image-v3/",
						OutputDir:       "/workspaces/dpb587.me/tmp/tilde/blob-iiif-image-v3/",
						IdentifierNamer: blobIdentifierNamer,
						FilterExiftool: []string{
							"-IPTC:CopyrightNotice=Copyright " + time.Now().Format("2006") + " Daniel Berger",
							"-IPTC:Credit=Daniel Berger",
						},
						PreferredSizeRequests: []imagerequest.RawParams{
							{"full", "!240,240", "0", "default.jpg"},
							{"full", "!480,480", "0", "default.jpg"},
							{"full", "!720,720", "0", "default.jpg"},
							{"full", "!1280,1280", "0", "default.jpg"},
							{"full", "!1920,1920", "0", "default.jpg"},
						},
					},
				}),
				exportgeojsonartifactsbuilder.NewBuilder(cGlobal.Log, locationmaskService, []exportgeojsonartifactsbuilder.BuilderProfile{
					{
						Name:            "default",
						BaseURL:         "/~/blob-geojson/",
						OutputDir:       "/workspaces/dpb587.me/tmp/tilde/blob-geojson/",
						IdentifierNamer: blobIdentifierNamer,
					},
				}),
			},
		})

		defer func() {
			err := repository.Close()
			if err != nil {
				cGlobal.Log.Error(
					"failed to close repository",
					"error", err,
				)
			}
		}()

		cGlobal.Repository = repository
	}

	//

	cmd := &cobra.Command{
		Use: "dpb587",
	}

	cmd.AddCommand(
		storyimportcmd.New(cGlobal),
		storyindexcmd.New(cGlobal),
	)

	return cmd.Execute()
}
