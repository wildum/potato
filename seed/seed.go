package seed

import (
	"time"

	"github.com/williamdumont/potato-demo/models"
	"github.com/williamdumont/potato-demo/storage"
)

func LoadSampleData(store storage.Storage) {
	potatoes := []models.Potato{
		{
			ID:          "p001",
			Variety:     "Russet",
			Origin:      "Idaho",
			Weight:      0.45,
			Quality:     string(models.Premium),
			HarvestDate: time.Now().AddDate(0, 0, -5),
			Price:       2.99,
		},
		{
			ID:          "p002",
			Variety:     "Yukon Gold",
			Origin:      "Canada",
			Weight:      0.38,
			Quality:     string(models.Premium),
			HarvestDate: time.Now().AddDate(0, 0, -3),
			Price:       3.49,
		},
		{
			ID:          "p003",
			Variety:     "Red Potato",
			Origin:      "Maine",
			Weight:      0.32,
			Quality:     string(models.Standard),
			HarvestDate: time.Now().AddDate(0, 0, -10),
			Price:       2.49,
		},
		{
			ID:          "p004",
			Variety:     "Fingerling",
			Origin:      "California",
			Weight:      0.25,
			Quality:     string(models.Premium),
			HarvestDate: time.Now().AddDate(0, 0, -2),
			Price:       4.99,
		},
		{
			ID:          "p005",
			Variety:     "Sweet Potato",
			Origin:      "North Carolina",
			Weight:      0.50,
			Quality:     string(models.Standard),
			HarvestDate: time.Now().AddDate(0, 0, -7),
			Price:       3.29,
		},
		{
			ID:          "p006",
			Variety:     "Purple Potato",
			Origin:      "Peru",
			Weight:      0.28,
			Quality:     string(models.Premium),
			HarvestDate: time.Now().AddDate(0, 0, -4),
			Price:       5.49,
		},
		{
			ID:          "p007",
			Variety:     "Russet",
			Origin:      "Washington",
			Weight:      0.52,
			Quality:     string(models.Standard),
			HarvestDate: time.Now().AddDate(0, 0, -15),
			Price:       2.79,
		},
		{
			ID:          "p008",
			Variety:     "Yukon Gold",
			Origin:      "Quebec",
			Weight:      0.41,
			Quality:     string(models.Economy),
			HarvestDate: time.Now().AddDate(0, 0, -20),
			Price:       1.99,
		},
	}

	for _, potato := range potatoes {
		store.AddPotato(potato)
	}

	recipes := []models.Recipe{
		{
			ID:          "r001",
			Name:        "Classic Baked Potato",
			Variety:     "Russet",
			CookingTime: 60,
			Difficulty:  "Easy",
			Ingredients: []string{
				"1 large Russet potato",
				"2 tbsp butter",
				"Salt and pepper",
				"Sour cream",
				"Chives",
			},
			Instructions: []string{
				"Preheat oven to 400°F (200°C)",
				"Wash and dry potato thoroughly",
				"Pierce potato several times with a fork",
				"Rub with oil and sprinkle with salt",
				"Bake for 50-60 minutes until tender",
				"Cut open and add butter, salt, and toppings",
			},
			Servings: 1,
		},
		{
			ID:          "r002",
			Name:        "Garlic Yukon Gold Mash",
			Variety:     "Yukon Gold",
			CookingTime: 30,
			Difficulty:  "Easy",
			Ingredients: []string{
				"2 lbs Yukon Gold potatoes",
				"4 cloves garlic",
				"1/2 cup milk",
				"4 tbsp butter",
				"Salt and pepper",
			},
			Instructions: []string{
				"Peel and cube potatoes",
				"Boil potatoes with garlic cloves for 20 minutes",
				"Drain and return to pot",
				"Add butter and milk",
				"Mash until smooth",
				"Season with salt and pepper",
			},
			Servings: 4,
		},
		{
			ID:          "r003",
			Name:        "Roasted Red Potatoes",
			Variety:     "Red Potato",
			CookingTime: 45,
			Difficulty:  "Easy",
			Ingredients: []string{
				"2 lbs Red potatoes",
				"3 tbsp olive oil",
				"2 tsp rosemary",
				"1 tsp thyme",
				"Salt and pepper",
			},
			Instructions: []string{
				"Preheat oven to 425°F (220°C)",
				"Cut potatoes into quarters",
				"Toss with oil and herbs",
				"Spread on baking sheet",
				"Roast for 40-45 minutes, turning once",
				"Serve hot",
			},
			Servings: 6,
		},
		{
			ID:          "r004",
			Name:        "Fancy Fingerling Medley",
			Variety:     "Fingerling",
			CookingTime: 35,
			Difficulty:  "Medium",
			Ingredients: []string{
				"1.5 lbs Fingerling potatoes",
				"3 tbsp butter",
				"2 cloves garlic minced",
				"Fresh thyme",
				"Lemon zest",
				"Sea salt",
			},
			Instructions: []string{
				"Halve fingerlings lengthwise",
				"Boil in salted water for 10 minutes",
				"Drain and pat dry",
				"Sauté in butter with garlic",
				"Add thyme and lemon zest",
				"Cook until golden brown",
			},
			Servings: 4,
		},
		{
			ID:          "r005",
			Name:        "Sweet Potato Fries",
			Variety:     "Sweet Potato",
			CookingTime: 30,
			Difficulty:  "Easy",
			Ingredients: []string{
				"2 large Sweet potatoes",
				"2 tbsp olive oil",
				"1 tsp paprika",
				"1/2 tsp garlic powder",
				"Salt",
			},
			Instructions: []string{
				"Preheat oven to 425°F (220°C)",
				"Cut potatoes into fry shapes",
				"Toss with oil and seasonings",
				"Arrange in single layer on baking sheet",
				"Bake for 25-30 minutes, flipping halfway",
				"Serve immediately",
			},
			Servings: 3,
		},
		{
			ID:          "r006",
			Name:        "Purple Potato Salad",
			Variety:     "Purple Potato",
			CookingTime: 25,
			Difficulty:  "Medium",
			Ingredients: []string{
				"2 lbs Purple potatoes",
				"1/4 cup olive oil",
				"2 tbsp white wine vinegar",
				"1 tbsp Dijon mustard",
				"Red onion",
				"Fresh dill",
			},
			Instructions: []string{
				"Boil whole potatoes until tender",
				"Cool and cut into bite-sized pieces",
				"Whisk together oil, vinegar, and mustard",
				"Toss potatoes with dressing",
				"Add chopped onion and dill",
				"Refrigerate for 1 hour before serving",
			},
			Servings: 6,
		},
	}

	for _, recipe := range recipes {
		store.AddRecipe(recipe)
	}
}

