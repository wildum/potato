package background

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/williamdumont/potato-demo/models"
	"github.com/williamdumont/potato-demo/storage"
)

var (
	varieties = []string{"Russet", "Yukon Gold", "Red Potato", "Fingerling", "Sweet Potato", "Purple Potato"}
	origins   = []string{"Idaho", "Washington", "Maine", "California", "North Carolina", "Quebec", "Peru", "Colorado"}
	qualities = []string{string(models.Premium), string(models.Standard), string(models.Economy)}
	
	recipeNames = map[string][]string{
		"Russet":        {"Loaded Baked Potato", "Perfect French Fries", "Potato Wedges", "Russet Gratin"},
		"Yukon Gold":    {"Creamy Potato Soup", "Golden Potato Pancakes", "Yukon Scalloped Potatoes"},
		"Red Potato":    {"Red Potato Hash", "Potato Salad Deluxe", "Herbed Red Potatoes"},
		"Fingerling":    {"Crispy Fingerlings", "Fingerling Confit", "Fancy Fingerling Salad"},
		"Sweet Potato":  {"Sweet Potato Casserole", "Sweet Potato Chips", "Candied Sweet Potatoes"},
		"Purple Potato": {"Purple Potato Mash", "Colorful Potato Medley", "Purple Potato Gnocchi"},
	}
	
	difficulties = []string{"Easy", "Medium", "Hard"}
	
	counter = 1000
)

type Worker struct {
	storage storage.Storage
}

func NewWorker(storage storage.Storage) *Worker {
	return &Worker{
		storage: storage,
	}
}

func (w *Worker) StartPotatoGenerator(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			w.addRandomPotato()
		}
	}()
}

func (w *Worker) StartRecipeGenerator(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			w.addRandomRecipe()
		}
	}()
}

func (w *Worker) StartQualityDegradation(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			w.degradePotatoQuality()
		}
	}()
}

func (w *Worker) addRandomPotato() {
	counter++
	id := fmt.Sprintf("p%d", counter)
	
	variety := varieties[rand.Intn(len(varieties))]
	origin := origins[rand.Intn(len(origins))]
	quality := qualities[rand.Intn(len(qualities))]
	
	weight := 0.20 + rand.Float64()*0.40
	basePrice := 2.0
	if quality == string(models.Premium) {
		basePrice = 3.5
	} else if quality == string(models.Economy) {
		basePrice = 1.5
	}
	price := basePrice + rand.Float64()*1.5
	
	daysAgo := rand.Intn(14)
	harvestDate := time.Now().AddDate(0, 0, -daysAgo)
	
	potato := models.Potato{
		ID:          id,
		Variety:     variety,
		Origin:      origin,
		Weight:      weight,
		Quality:     quality,
		HarvestDate: harvestDate,
		Price:       price,
	}
	
	w.storage.AddPotato(potato)
}

func (w *Worker) addRandomRecipe() {
	counter++
	id := fmt.Sprintf("r%d", counter)
	
	variety := varieties[rand.Intn(len(varieties))]
	names := recipeNames[variety]
	name := names[rand.Intn(len(names))]
	
	difficulty := difficulties[rand.Intn(len(difficulties))]
	cookingTime := 20 + rand.Intn(60)
	servings := 2 + rand.Intn(6)
	
	ingredients := generateRandomIngredients(variety)
	instructions := generateRandomInstructions()
	
	recipe := models.Recipe{
		ID:           id,
		Name:         name,
		Variety:      variety,
		CookingTime:  cookingTime,
		Difficulty:   difficulty,
		Ingredients:  ingredients,
		Instructions: instructions,
		Servings:     servings,
	}
	
	w.storage.AddRecipe(recipe)
}

func (w *Worker) degradePotatoQuality() {
	potatoes := w.storage.GetAllPotatoes()
	
	for _, potato := range potatoes {
		daysSinceHarvest := int(time.Since(potato.HarvestDate).Hours() / 24)
		
		if daysSinceHarvest > 30 && potato.Quality == string(models.Premium) {
			potato.Quality = string(models.Standard)
			w.storage.UpdatePotato(potato.ID, potato)
		} else if daysSinceHarvest > 60 && potato.Quality == string(models.Standard) {
			potato.Quality = string(models.Economy)
			w.storage.UpdatePotato(potato.ID, potato)
		}
	}
}

func generateRandomIngredients(variety string) []string {
	ingredients := []string{
		fmt.Sprintf("%d lbs %s potatoes", 1+rand.Intn(3), variety),
	}
	
	extras := []string{
		"Salt and pepper",
		"Olive oil",
		"Butter",
		"Garlic cloves",
		"Fresh herbs",
		"Heavy cream",
		"Cheese",
		"Onions",
	}
	
	numExtras := 2 + rand.Intn(4)
	rand.Shuffle(len(extras), func(i, j int) {
		extras[i], extras[j] = extras[j], extras[i]
	})
	
	for i := 0; i < numExtras && i < len(extras); i++ {
		ingredients = append(ingredients, extras[i])
	}
	
	return ingredients
}

func generateRandomInstructions() []string {
	return []string{
		"Prepare all ingredients",
		"Wash and prepare potatoes",
		"Follow cooking method appropriate for the dish",
		"Season to taste",
		"Cook until golden and tender",
		"Serve hot and enjoy",
	}
}

