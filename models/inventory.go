package models

type InventoryItem struct {
	Variety       string  `json:"variety"`
	TotalQuantity int     `json:"total_quantity"`
	TotalWeight   float64 `json:"total_weight"`
	AveragePrice  float64 `json:"average_price"`
}

type InventorySummary struct {
	TotalPotatoes int              `json:"total_potatoes"`
	TotalWeight   float64          `json:"total_weight"`
	TotalValue    float64          `json:"total_value"`
	ByVariety     []InventoryItem  `json:"by_variety"`
}

type PotatoAnalytics struct {
	MostPopularVariety string  `json:"most_popular_variety"`
	AverageWeight      float64 `json:"average_weight"`
	PremiumPercentage  float64 `json:"premium_percentage"`
	TotalValue         float64 `json:"total_value"`
}

