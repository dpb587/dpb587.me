package storybuilder

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/dpb587/dpb587.me/tools/content"
	"github.com/dpb587/dpb587.me/tools/content/frontmatterparams"
	"github.com/dpb587/dpb587.me/tools/content/hugoutil"
	"github.com/dpb587/tacitkb/catalog"
	"github.com/dpb587/tacitkb/catalog/catalogutil"
	"github.com/dpb587/tacitkb/ext/blob"
	"github.com/dpb587/tacitkb/ext/exportgeojson"
	"github.com/dpb587/tacitkb/util/ptrutil"
	"github.com/tkrajina/gpxgo/gpx"
)

var routeString = "route"

func (b *Service) buildRouteType(ctx context.Context, blobNode catalog.Node, blobProfile *blob.ProfileResourceData) (*content.Document, error) {
	templateData := &frontmatterparams.RouteType{
		// CatalogNodeUID: ptrutil.Value(string(blobNode.UID())),
	}

	doc := &content.Document{
		Frontmatter: &content.Content_Frontmatter{
			Params: &content.Content_Frontmatter_Params{
				Nav: &frontmatterparams.Nav{
					Type: &frontmatterparams.Nav_Type{
						"route": true,
					},
				},
				RouteType: templateData,
			},
			Type: &routeString,
		},
	}

	//

	blobContentReader, err := catalogutil.NewNodeResourceReader(ctx, b.repository, blobNode.UID(), blob.ContentResourceDescriptor{})
	if err != nil {
		return nil, fmt.Errorf("open content: %v", err)
	}

	defer blobContentReader.Close()

	gpxBytes, err := io.ReadAll(blobContentReader)
	if err != nil {
		return nil, fmt.Errorf("read blob: %v", err)
	}

	gpxFile, err := gpx.ParseBytes(gpxBytes)
	if err != nil {
		return nil, fmt.Errorf("parse gpx: %v", err)
	}

	//

	startPoint := gpxFile.Tracks[0].Segments[0].Points[0]
	endPoint := gpxFile.Tracks[len(gpxFile.Tracks)-1].Segments[len(gpxFile.Tracks[len(gpxFile.Tracks)-1].Segments)-1].Points[len(gpxFile.Tracks[len(gpxFile.Tracks)-1].Segments[len(gpxFile.Tracks[len(gpxFile.Tracks)-1].Segments)-1].Points)-1]

	uphillDownhill := gpxFile.UphillDownhill()
	elevationBounds := gpxFile.ElevationBounds()

	{
		timeBounds := gpxFile.TimeBounds()

		doc.Frontmatter.Date = ptrutil.Value(hugoutil.NewFrontmatterTime(time.RFC3339, timeBounds.StartTime))
		doc.Frontmatter.Params.TimeRange = &frontmatterparams.TimeRange{
			From: doc.Frontmatter.Date,
			Thru: ptrutil.Value(hugoutil.NewFrontmatterTime(time.RFC3339, timeBounds.EndTime)),
		}
	}

	templateData.Duration = &frontmatterparams.RouteType_Quantity{
		Value: endPoint.Timestamp.Sub(startPoint.Timestamp).Seconds(),
		Unit:  "SEC",
	}

	templateData.Distance = &frontmatterparams.RouteType_Quantity{
		Value: gpxFile.Length3D(),
		Unit:  "MTR",
	}

	templateData.QuantityStats = []frontmatterparams.RouteType_QuantityStat{
		{
			Name: "Elevation Gain",
			Quantity: frontmatterparams.RouteType_Quantity{
				Value: uphillDownhill.Uphill,
				Unit:  "MTR",
			},
		},
		{
			Name: "Elevation Loss",
			Quantity: frontmatterparams.RouteType_Quantity{
				Value: uphillDownhill.Downhill,
				Unit:  "MTR",
			},
		},
		{
			Name: "Lowest Elevation",
			Quantity: frontmatterparams.RouteType_Quantity{
				Value: elevationBounds.MinElevation,
				Unit:  "MTR",
			},
		},
		{
			Name: "Highest Elevation",
			Quantity: frontmatterparams.RouteType_Quantity{
				Value: elevationBounds.MaxElevation,
				Unit:  "MTR",
			},
		},
	}

	{
		templateData.StartTime = &frontmatterparams.RouteType_Time{
			Time: hugoutil.NewFrontmatterTime(time.RFC3339, startPoint.Timestamp),
		}

		templateData.StartLocation = &frontmatterparams.RouteType_Location{
			GeoCoordinates: &frontmatterparams.RouteType_Location_GeoCoordinates{
				Latitude:  &startPoint.Latitude,
				Longitude: &startPoint.Longitude,
			},
		}

		if !startPoint.Elevation.Null() {
			templateData.StartLocation.GeoCoordinates.Elevation = ptrutil.Value(startPoint.Elevation.Value())
		}
	}

	{
		templateData.EndTime = &frontmatterparams.RouteType_Time{
			Time: hugoutil.NewFrontmatterTime(time.RFC3339, endPoint.Timestamp),
		}

		templateData.EndLocation = &frontmatterparams.RouteType_Location{
			GeoCoordinates: &frontmatterparams.RouteType_Location_GeoCoordinates{
				Latitude:  &endPoint.Latitude,
				Longitude: &endPoint.Longitude,
			},
		}

		if !endPoint.Elevation.Null() {
			templateData.EndLocation.GeoCoordinates.Elevation = ptrutil.Value(endPoint.Elevation.Value())
		}
	}

	//

	if len(gpxFile.Tracks[0].Name) > 0 {
		doc.Frontmatter.Title = ptrutil.Value(gpxFile.Tracks[0].Name)
	} else if len(gpxFile.Tracks[0].Type) > 0 {
		doc.Frontmatter.Title = ptrutil.Value(strings.ToTitle(gpxFile.Tracks[0].Type[0:1]) + gpxFile.Tracks[0].Type[1:])
	}

	err = func() error {
		artifactsResource, err := b.repository.GetResource(ctx, blobNode.UID(), exportgeojson.ArtifactsResourceDescriptor{
			Profile: "default",
		}, catalog.RepositoryGetResourceConfig{
			Generate: true,
		})
		if err != nil {
			return fmt.Errorf("get: %w", err)
		}

		artifactsData, err := exportgeojson.UnmarshalArtifactsResource(ctx, artifactsResource)
		if err != nil {
			return fmt.Errorf("unmarshal: %w", err)
		}

		templateData.GeoJSON = &frontmatterparams.RouteType_GeoJSON{
			URL:             artifactsData.URL,
			ExtraProperties: artifactsData.ExtraProperties,
		}

		return nil
	}()
	if err != nil {
		return nil, fmt.Errorf("exportgeojson: %v", err)
	}

	return doc, nil
}
