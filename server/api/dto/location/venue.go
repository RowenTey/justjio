package model_location

type Venue struct {
	Name              string `json:"name"`
	Address           string `json:"address"`
	GoogleMapsPlaceID string `json:"googleMapsPlaceId"`
}
