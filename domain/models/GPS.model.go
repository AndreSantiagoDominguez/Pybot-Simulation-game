package models

type GPS struct {
	Prototype_id string `json:"prototype_id"`
	Lat          string `json:"lat"`
	Lon          string `json:"lon"`
	Spd          string `json:"spd"`
	Date         string `json:"date"`
	UTC          string `json:"UTC"`
	Alt          string `json:"alt"`
	Sats         string `json:"sats"`
}