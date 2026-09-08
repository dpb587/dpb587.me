package storybuilder

import (
	"fmt"
	"math"
	"strings"
	"time"

	"context"

	"github.com/dpb587/dpb587.me/tools/content"
	"github.com/dpb587/dpb587.me/tools/content/frontmatterparams"
	"github.com/dpb587/dpb587.me/tools/content/hugoutil"
	"github.com/dpb587/tacitkb/catalog"
	"github.com/dpb587/tacitkb/catalog/catalogutil"
	"github.com/dpb587/tacitkb/ext/blob"
	"github.com/dpb587/tacitkb/ext/blobtypevideo"
	"github.com/dpb587/tacitkb/ext/exportvideostream"
	"github.com/dpb587/tacitkb/tools/googlemapsreversegeocode"
	"github.com/dpb587/tacitkb/util/ptrutil"
)

func (b *Service) buildMediaVideo(ctx context.Context, blobNode catalog.Node, blobProfile *blob.ProfileResourceData) (*content.Document, error) {
	templateData := &frontmatterparams.MediaType{
		CatalogNodeUID: ptrutil.Value(string(blobNode.UID())),
	}

	doc := &content.Document{
		Frontmatter: &content.Content_Frontmatter{
			Params: &content.Content_Frontmatter_Params{
				MediaType: templateData,
			},
			Type: &mediaString,
		},
	}

	{
		blobContentResource, err := b.repository.GetResource(ctx, blobNode.UID(), blob.ContentResourceDescriptor{})
		if err != nil {
			return nil, fmt.Errorf("get content: %w", err)
		}

		blobContentResourceContents, err := b.repository.GetResourceContentList(ctx, blobContentResource)
		if err != nil {
			return nil, fmt.Errorf("get content contents: %w", err)
		}

		if filepathName, ok := catalogutil.GetResourceContentFilepathBase(blobContentResourceContents); ok {
			doc.Frontmatter.Title = ptrutil.Value(filepathName)
		}
	}

	var profileData *blobtypevideo.ProfileResourceData

	err := func() error {
		profileResource, err := b.repository.GetResource(ctx, blobNode.UID(), blobtypevideo.ProfileResourceDescriptor{}, catalog.RepositoryGetResourceConfig{
			Generate: true,
		})
		if err != nil {
			return fmt.Errorf("get: %w", err)
		}

		data, err := blobtypevideo.UnmarshalProfileResource(ctx, profileResource)
		if err != nil {
			return fmt.Errorf("unmarshal: %w", err)
		}

		profileData = data

		templateData.Width = data.Width
		templateData.Height = data.Height

		return nil
	}()
	if err != nil {
		return nil, fmt.Errorf("blobtypevideo: %v", err)
	}

	if profileData.CreationTime != nil {
		err := func() error {
			frontmatterLayout := "2006-01-02T15:04:05Z07:00"

			parsed, err := time.Parse(time.RFC3339Nano, *profileData.CreationTime)
			if err != nil {
				b.log.Info(
					"failed to parse video creation time",
					"input", *profileData.CreationTime,
				)

				return nil
			}

			frontmatterTime := hugoutil.NewFrontmatterTime(frontmatterLayout, parsed)

			doc.Frontmatter.Date = &frontmatterTime
			templateData.CaptureTime = &frontmatterparams.MediaType_CaptureTime{
				Time: frontmatterTime,
			}

			return nil
		}()
		if err != nil {
			return nil, fmt.Errorf("parse creation time: %v", err)
		}
	}

	if profileData.GPSCoordinates != nil {
		err := func() error {
			templateData.GeoCoordinates = &frontmatterparams.MediaType_GeoCoordinates{
				Latitude:  ptrutil.Value(profileData.GPSCoordinates.Latitude),
				Longitude: ptrutil.Value(profileData.GPSCoordinates.Longitude),
				Elevation: profileData.GPSCoordinates.Altitude,
			}

			reverseGeocode, err := b.rgeo.LookupLocation(googlemapsreversegeocode.LookupLocationInput{
				Latitude:  profileData.GPSCoordinates.Latitude,
				Longitude: profileData.GPSCoordinates.Longitude,
			})
			if err != nil {
				return fmt.Errorf("reverse geocode: %v", err)
			}

			templateData.PlacesProfile = &frontmatterparams.MediaType_PlacesProfile{}

			if reverseGeocode.Country != nil {
				templateData.PlacesProfile.Country = &frontmatterparams.MediaType_PlacesProfile_Country{
					Code: reverseGeocode.Country.PrimaryShortName,
					Name: reverseGeocode.Country.PrimaryLongName,
				}
			} else {
				panic("expected country")
			}

			if reverseGeocode.Admin1 != nil {
				templateData.PlacesProfile.CountryRegion = &frontmatterparams.MediaType_PlacesProfile_CountryRegion{
					Code: reverseGeocode.Country.PrimaryShortName + "-" + reverseGeocode.Admin1.PrimaryShortName,
					Name: reverseGeocode.Admin1.PrimaryLongName,
				}

				switch reverseGeocode.Country.PrimaryShortName {
				case "CA", "US":
					doc.Frontmatter.Params.SetTopicPlace(
						strings.ToLower(fmt.Sprintf("%s/%s", reverseGeocode.Country.PrimaryShortName, reverseGeocode.Admin1.PrimaryShortName)),
						frontmatterparams.TopicParam{},
					)
				default:
					doc.Frontmatter.Params.SetTopicPlace(
						strings.ToLower(reverseGeocode.Country.PrimaryShortName),
						frontmatterparams.TopicParam{},
					)
				}
			} else {
				doc.Frontmatter.Params.SetTopicPlace(
					strings.ToLower(reverseGeocode.Country.PrimaryShortName),
					frontmatterparams.TopicParam{},
				)
			}

			if reverseGeocode.Admin2 != nil {
				templateData.PlacesProfile.Places = append(templateData.PlacesProfile.Places, frontmatterparams.MediaType_PlacesProfile_Place{
					Name: reverseGeocode.Admin2.PrimaryLongName,
					Kind: "admin2",
				})
			}

			if reverseGeocode.Locality != nil {
				templateData.PlacesProfile.Places = append(templateData.PlacesProfile.Places, frontmatterparams.MediaType_PlacesProfile_Place{
					Name: reverseGeocode.Locality.PrimaryLongName,
					Kind: "locality",
				})
			}

			return nil
		}()
		if err != nil {
			return nil, fmt.Errorf("gps: %v", err)
		}
	}

	err = func() error {
		artifactsResource, err := b.repository.GetResource(ctx, blobNode.UID(), exportvideostream.ArtifactsResourceDescriptor{
			Profile: "default",
		}, catalog.RepositoryGetResourceConfig{
			Generate: true,
		})
		if err != nil {
			return fmt.Errorf("get: %w", err)
		}

		artifactsData, err := exportvideostream.UnmarshalArtifactsResource(ctx, artifactsResource)
		if err != nil {
			return fmt.Errorf("unmarshal: %w", err)
		}

		templateData.VideoService = &frontmatterparams.MediaType_VideoService{
			PlaylistURL: artifactsData.PlaylistURL,
			VTTURL:      artifactsData.VTTURL,
			DurationSec: int(math.Round(artifactsData.DurationSec)),
		}

		if artifactsData.StartFrame != nil {
			templateData.VideoService.StartThumbnails = buildPosterFrameThumbnails(artifactsData.StartFrame)
		}

		if artifactsData.EndFrame != nil {
			templateData.VideoService.EndThumbnails = buildPosterFrameThumbnails(artifactsData.EndFrame)
		}

		return nil
	}()
	if err != nil {
		return nil, fmt.Errorf("exportvideostream: %v", err)
	}

	return doc, nil
}

func buildPosterFrameThumbnails(frame *exportvideostream.ArtifactsResourceData_PosterFrame) frontmatterparams.MediaType_ThumbnailList {
	var thumbnails frontmatterparams.MediaType_ThumbnailList

	for _, size := range frame.Sizes {
		thumbnails = append(thumbnails, frontmatterparams.MediaType_Thumbnail{
			Width:  size.Width,
			Height: size.Height,
			URL:    frame.ServiceURL + fmt.Sprintf("/full/%d%%2C%d/0/default.jpg", size.Width, size.Height),
		})
	}

	return thumbnails
}
