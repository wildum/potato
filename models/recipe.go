package models

type Recipe struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Variety      string   `json:"variety"`
	CookingTime  int      `json:"cooking_time"`
	Difficulty   string   `json:"difficulty"`
	Ingredients  []string `json:"ingredients"`
	Instructions []string `json:"instructions"`
	Servings     int      `json:"servings"`
}

type CookingMethod string

const (
	Baked  CookingMethod = "Baked"
	Fried  CookingMethod = "Fried"
	Mashed CookingMethod = "Mashed"
	Boiled CookingMethod = "Boiled"
	Roasted CookingMethod = "Roasted"
)

