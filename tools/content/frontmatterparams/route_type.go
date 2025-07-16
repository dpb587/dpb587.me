package frontmatterparams

import "github.com/dpb587/dpb587.me/tools/content/hugoutil"

type RouteType struct {
	StartTime     *RouteType_Time          `json:"startTime,omitempty"`
	StartLocation *RouteType_Location      `json:"startLocation,omitempty"`
	EndTime       *RouteType_Time          `json:"endTime,omitempty"`
	EndLocation   *RouteType_Location      `json:"endLocation,omitempty"`
	Duration      *RouteType_Quantity      `json:"duration,omitempty"`
	Distance      *RouteType_Quantity      `json:"distance,omitempty"`
	QuantityStats []RouteType_QuantityStat `json:"quantityStats,omitempty"`

	GeoJSON *RouteType_GeoJSON `json:"geojson,omitempty"`
}

type RouteType_GeoJSON struct {
	URL             string   `json:"url"`
	ExtraProperties []string `json:"extraProperties,omitempty"`
}

type RouteType_QuantityStat struct {
	Name     string             `json:"name"`
	Quantity RouteType_Quantity `json:"quantity"`
}

type RouteType_Quantity struct {
	Unit  string  `json:"unit"`
	Value float64 `json:"value"`
}

type RouteType_Location struct {
	GeoCoordinates *RouteType_Location_GeoCoordinates `json:"geoCoordinates,omitempty"`
	PlacesProfile  *RouteType_Location_PlacesProfile  `json:"placesProfile,omitempty"`
}

type RouteType_Time struct {
	Time hugoutil.FrontmatterTime `json:"time"`
	// TimeZoneName string
}

type RouteType_Location_GeoCoordinates struct {
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
	Elevation *float64 `json:"elevation,omitempty"`
}

type RouteType_Location_PlacesProfile struct {
	Country       *RouteType_Location_PlacesProfile_Country       `json:"country,omitempty"`
	CountryRegion *RouteType_Location_PlacesProfile_CountryRegion `json:"countryRegion,omitempty"`
	Places        []RouteType_Location_PlacesProfile_Place        `json:"places,omitempty"`
}

type RouteType_Location_PlacesProfile_Country struct {
	Code string `json:"code,omitempty"`
	Name string `json:"name,omitempty"`
}

type RouteType_Location_PlacesProfile_CountryRegion struct {
	Code string `json:"code,omitempty"`
	Name string `json:"name,omitempty"`
}

type RouteType_Location_PlacesProfile_Place struct {
	Name string `json:"name,omitempty"`
	Kind string `json:"kind,omitempty"`
}
