package models

import "time"

type Potato struct {
	ID          string    `json:"id"`
	Variety     string    `json:"variety"`
	Origin      string    `json:"origin"`
	Weight      float64   `json:"weight"`
	Quality     string    `json:"quality"`
	HarvestDate time.Time `json:"harvest_date"`
	Price       float64   `json:"price"`
}

type PotatoVariety string

const (
	Russet       PotatoVariety = "Russet"
	Yukon        PotatoVariety = "Yukon Gold"
	RedPotato    PotatoVariety = "Red Potato"
	Fingerling   PotatoVariety = "Fingerling"
	SweetPotato  PotatoVariety = "Sweet Potato"
	PurplePotato PotatoVariety = "Purple Potato"
)

type Quality string

const (
	Premium  Quality = "Premium"
	Standard Quality = "Standard"
	Economy  Quality = "Economy"
)

