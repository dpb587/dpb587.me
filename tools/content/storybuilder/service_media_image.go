package storybuilder

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/dpb587/dpb587.me/tools/content"
	"github.com/dpb587/dpb587.me/tools/content/frontmatterparams"
	"github.com/dpb587/dpb587.me/tools/content/hugoutil"
	"github.com/dpb587/tacitkb/catalog"
	"github.com/dpb587/tacitkb/catalog/catalogutil"
	"github.com/dpb587/tacitkb/ext/blob"
	"github.com/dpb587/tacitkb/ext/blob/blobmeta"
	"github.com/dpb587/tacitkb/ext/blobmetaexif"
	"github.com/dpb587/tacitkb/ext/blobmetaexif/blobmetaexifutil"
	"github.com/dpb587/tacitkb/ext/blobtypeimage"
	"github.com/dpb587/tacitkb/ext/exportiiifimage3"
	"github.com/dpb587/tacitkb/tools/googlemapsreversegeocode"
	"github.com/dpb587/tacitkb/util/ptrutil"
)

var mediaString = "media"

func (b *Service) buildMediaImage(ctx context.Context, blobNode catalog.Node, blobProfile *blob.ProfileResourceData) (*content.Document, error) {
	templateData := &frontmatterparams.MediaType{
		CatalogNodeUID: ptrutil.Value(string(blobNode.UID())),
	}

	doc := &content.Document{
		Frontmatter: &content.Content_Frontmatter{
			Params: &content.Content_Frontmatter_Params{
				Nav: &frontmatterparams.Nav{
					Type: &frontmatterparams.Nav_Type{
						"media": true,
					},
				},
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

	err := func() error {
		profileResource, err := b.repository.GetResource(ctx, blobNode.UID(), blobtypeimage.ProfileResourceDescriptor{}, catalog.RepositoryGetResourceConfig{
			Generate: true,
		})
		if err != nil {
			return fmt.Errorf("get: %w", err)
		}

		profileData, err := blobtypeimage.UnmarshalProfileResource(ctx, profileResource)
		if err != nil {
			return fmt.Errorf("unmarshal: %w", err)
		}

		templateData.Width = profileData.Objects[0].Width
		templateData.Height = profileData.Objects[0].Height

		return nil
	}()
	if err != nil {
		return nil, fmt.Errorf("blobtypeimage: %v", err)
	}

	err = func() error {
		artifactsResource, err := b.repository.GetResource(ctx, blobNode.UID(), exportiiifimage3.ArtifactsResourceDescriptor{
			Profile: "default",
		}, catalog.RepositoryGetResourceConfig{
			Generate: true,
		})
		if err != nil {
			return fmt.Errorf("get: %w", err)
		}

		artifactsData, err := exportiiifimage3.UnmarshalArtifactsResource(ctx, artifactsResource)
		if err != nil {
			return fmt.Errorf("unmarshal: %w", err)
		}

		templateData.ImageService = &frontmatterparams.MediaType_ImageService{
			InfoURL: artifactsData.ServiceURL + "/info.json",
		}

		for _, size := range artifactsData.Sizes {
			templateData.Thumbnails = append(templateData.Thumbnails, frontmatterparams.MediaType_Thumbnail{
				Width:  size.Width,
				Height: size.Height,
				URL:    artifactsData.ServiceURL + fmt.Sprintf("/full/%d%%2C%d/0/default.jpg", size.Width, size.Height),
			})
		}

		return nil
	}()
	if err != nil {
		return nil, fmt.Errorf("exportiiifimage3: %v", err)
	}

	err = func() error {
		profileResource, err := b.repository.GetResource(ctx, blobNode.UID(), blobmetaexif.ProfileResourceDescriptor{}, catalog.RepositoryGetResourceConfig{
			Generate: true,
		})
		if err != nil {
			if errors.Is(err, catalog.ErrResourceNotFound) {
				// exif may not be present
				return nil
			}

			return fmt.Errorf("get: %w", err)
		}

		profileData, err := blobmetaexif.UnmarshalProfileResource(ctx, profileResource)
		if err != nil {
			return fmt.Errorf("unmarshal: %w", err)
		}

		setExif := func(f func(*frontmatterparams.MediaType_ExifProfile)) {
			if templateData.ExifProfile == nil {
				templateData.ExifProfile = &frontmatterparams.MediaType_ExifProfile{}
			}

			f(templateData.ExifProfile)
		}

		for _, collection := range profileData.Scopes[0].Fieldsets {
			if collection.Schema != "Exif" {
				continue
			}

			switch collection.Name {
			case "IFD0":
				for _, field := range collection.Fields {
					switch field.Name {
					case "0x10f":
						setExif(func(exif *frontmatterparams.MediaType_ExifProfile) {
							exif.Make = &frontmatterparams.MediaType_ExifProfile_Value{
								String: ptrutil.Value(field.Values[0].(blobmeta.Profile_Scope_Fieldset_Field_ValueString).String),
							}
						})
					case "0x110":
						setExif(func(exif *frontmatterparams.MediaType_ExifProfile) {
							exif.Model = &frontmatterparams.MediaType_ExifProfile_Value{
								String: ptrutil.Value(field.Values[0].(blobmeta.Profile_Scope_Fieldset_Field_ValueString).String),
							}
						})
					}
				}
			case "IFD0/Exif":
				var dateTimeOriginal string
				var dateTimeZone string

				for _, field := range collection.Fields {
					switch field.Name {
					case "0x829d":
						setExif(func(exif *frontmatterparams.MediaType_ExifProfile) {
							v := field.Values[0].(blobmeta.Profile_Scope_Fieldset_Field_ValueRationalInt32)

							exif.ApertureValue = &frontmatterparams.MediaType_ExifProfile_Value{
								Number: ptrutil.Value(float64(v.RationalInt32[0]) / float64(v.RationalInt32[1])),
							}
						})
					case "0x8827":
						setExif(func(exif *frontmatterparams.MediaType_ExifProfile) {
							exif.ISO = &frontmatterparams.MediaType_ExifProfile_Value{
								Number: ptrutil.Value(float64(field.Values[0].(blobmeta.Profile_Scope_Fieldset_Field_ValueInt32).Int32)),
							}
						})
					case "0x9003":
						dateTimeOriginal = field.Values[0].(blobmeta.Profile_Scope_Fieldset_Field_ValueString).String
					case "0x9011":
						dateTimeZone = field.Values[0].(blobmeta.Profile_Scope_Fieldset_Field_ValueString).String
					case "0x9201":
						setExif(func(exif *frontmatterparams.MediaType_ExifProfile) {
							v := field.Values[0].(blobmeta.Profile_Scope_Fieldset_Field_ValueRationalInt32)

							exif.ShutterSpeedValue = &frontmatterparams.MediaType_ExifProfile_Value{
								Number: ptrutil.Value(math.Pow(2, float64(v.RationalInt32[0])/float64(v.RationalInt32[1]))),
							}
						})
					case "0x920a":
						setExif(func(exif *frontmatterparams.MediaType_ExifProfile) {
							v := field.Values[0].(blobmeta.Profile_Scope_Fieldset_Field_ValueRationalInt32)

							exif.FocalLength = &frontmatterparams.MediaType_ExifProfile_Value{
								Number: ptrutil.Value(float64(v.RationalInt32[0]) / float64(v.RationalInt32[1])),
							}
						})
					case "0xa433":
						setExif(func(exif *frontmatterparams.MediaType_ExifProfile) {
							exif.LensMake = &frontmatterparams.MediaType_ExifProfile_Value{
								String: ptrutil.Value(field.Values[0].(blobmeta.Profile_Scope_Fieldset_Field_ValueString).String),
							}
						})
					case "0xa434":
						setExif(func(exif *frontmatterparams.MediaType_ExifProfile) {
							exif.LensModel = &frontmatterparams.MediaType_ExifProfile_Value{
								String: ptrutil.Value(field.Values[0].(blobmeta.Profile_Scope_Fieldset_Field_ValueString).String),
							}
						})
					}
				}

				if len(dateTimeOriginal) > 0 {
					err := func() error {
						effectiveInput := dateTimeOriginal
						expectedLayout := "2006:01:02 15:04:05"
						frontmatterLayout := "2006-01-02T15:04:05"

						if len(dateTimeZone) > 0 {
							effectiveInput += " " + dateTimeZone
							expectedLayout += " -07:00"
							frontmatterLayout += "Z07:00"
						}

						parsed, err := time.Parse(expectedLayout, effectiveInput)
						if err != nil {
							b.log.Info(
								"failed to parse exif datetime",
								"input", effectiveInput,
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
						return fmt.Errorf("parse datetime: %v", err)
					}
				}
			case "IFD0/GPSInfo":
				gpsCoordinates, ok := blobmetaexifutil.FindGPSInfoProfile(collection)
				if !ok {
					continue
				}

				templateData.GeoCoordinates = &frontmatterparams.MediaType_GeoCoordinates{}

				if gpsCoordinates.LatLng != nil {
					templateData.GeoCoordinates.Latitude = &(*gpsCoordinates.LatLng)[0]
					templateData.GeoCoordinates.Longitude = &(*gpsCoordinates.LatLng)[1]

					reverseGeocode, err := b.rgeo.LookupLocation(googlemapsreversegeocode.LookupLocationInput{
						Latitude:  *templateData.GeoCoordinates.Latitude,
						Longitude: *templateData.GeoCoordinates.Longitude,
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
							doc.Frontmatter.Params.SetNavPlaceArea(
								strings.ToLower(fmt.Sprintf("%s/%s", reverseGeocode.Country.PrimaryShortName, reverseGeocode.Admin1.PrimaryShortName)),
								true,
							)
						default:
							doc.Frontmatter.Params.SetNavPlaceArea(
								strings.ToLower(reverseGeocode.Country.PrimaryShortName),
								true,
							)
						}
					} else {
						doc.Frontmatter.Params.SetNavPlaceArea(
							strings.ToLower(reverseGeocode.Country.PrimaryShortName),
							true,
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

					// if reverseGeocode.Sublocality1 != nil {
					// 	templateData.PlacesProfile.Places = append(templateData.PlacesProfile.Places, frontmatterparams.MediaType_PlacesProfile_Place{
					// 		Name: reverseGeocode.Sublocality1.PrimaryLongName,
					// 		Kind: "sublocality1",
					// 	})
					// }

					// if reverseGeocode.Neighborhood != nil {
					// 	templateData.PlacesProfile.Places = append(templateData.PlacesProfile.Places, frontmatterparams.MediaType_PlacesProfile_Place{
					// 		Name: reverseGeocode.Neighborhood.PrimaryLongName,
					// 		Kind: "neighborhood",
					// 	})
					// }
				}

				if gpsCoordinates.Altitude != nil {
					templateData.GeoCoordinates.Elevation = gpsCoordinates.Altitude
				}
			}
		}

		return nil
	}()
	if err != nil {
		return nil, fmt.Errorf("exif: %v", err)
	}

	//

	return doc, nil
}
