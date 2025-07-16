package frontmatterparams

import "github.com/dpb587/dpb587.me/tools/content/hugoutil"

type MediaType struct {
	CatalogNodeUID *string `json:"catalogNodeUid,omitempty"`

	Width  int `json:"width"`
	Height int `json:"height"`

	Thumbnails   MediaType_ThumbnailList `json:"thumbnails,omitempty"`
	ImageService *MediaType_ImageService `json:"imageService,omitempty"`

	CaptureTime    *MediaType_CaptureTime    `json:"captureTime,omitempty"`
	GeoCoordinates *MediaType_GeoCoordinates `json:"geoCoordinates,omitempty"`
	PlacesProfile  *MediaType_PlacesProfile  `json:"placesProfile,omitempty"`
	ExifProfile    *MediaType_ExifProfile    `json:"exifProfile,omitempty"`
}

type MediaType_Thumbnail struct {
	URL    string `json:"url,omitempty"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

type MediaType_ThumbnailList []MediaType_Thumbnail

type MediaType_ImageService struct {
	InfoURL string `json:"infoUrl,omitempty"`
}

type MediaType_CaptureTime struct {
	Time hugoutil.FrontmatterTime `json:"time"`
	// TimeZoneName string
}

type MediaType_ExifProfile struct {
	Make              *MediaType_ExifProfile_Value `json:"make,omitempty"`
	Model             *MediaType_ExifProfile_Value `json:"model,omitempty"`
	LensMake          *MediaType_ExifProfile_Value `json:"lensMake,omitempty"`
	LensModel         *MediaType_ExifProfile_Value `json:"lensModel,omitempty"`
	ISO               *MediaType_ExifProfile_Value `json:"iso,omitempty"`
	FocalLength       *MediaType_ExifProfile_Value `json:"focalLength,omitempty"`
	ApertureValue     *MediaType_ExifProfile_Value `json:"apertureValue,omitempty"`
	ShutterSpeedValue *MediaType_ExifProfile_Value `json:"shutterSpeedValue,omitempty"`
}

type MediaType_ExifProfile_Value struct {
	String   *string                               `json:"string,omitempty"`
	Number   *float64                              `json:"number,omitempty"`
	Rational *MediaType_ExifProfile_Value_Rational `json:"rational,omitempty"`
}

type MediaType_ExifProfile_Value_Rational struct {
	Numerator   int64 `json:"numerator,omitempty"`
	Denominator int64 `json:"denominator,omitempty"`
}

type MediaType_GeoCoordinates struct {
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
	Elevation *float64 `json:"elevation,omitempty"`
}

type MediaType_PlacesProfile struct {
	Country       *MediaType_PlacesProfile_Country       `json:"country,omitempty"`
	CountryRegion *MediaType_PlacesProfile_CountryRegion `json:"countryRegion,omitempty"`
	Places        []MediaType_PlacesProfile_Place        `json:"places,omitempty"`
}

type MediaType_PlacesProfile_Country struct {
	Code string `json:"code,omitempty"`
	Name string `json:"name,omitempty"`
}

type MediaType_PlacesProfile_CountryRegion struct {
	Code string `json:"code,omitempty"`
	Name string `json:"name,omitempty"`
}

type MediaType_PlacesProfile_Place struct {
	Name string `json:"name,omitempty"`
	Kind string `json:"kind,omitempty"`
}
