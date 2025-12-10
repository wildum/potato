package background

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/williamdumont/potato-demo/models"
	"github.com/williamdumont/potato-demo/storage"
	logapi "go.opentelemetry.io/otel/log"
)

// Fake user emails for demo/exercise purposes (simulating sensitive data leak)
var fakeUserEmails = []string{
	"john.smith@example.com",
	"alice.johnson@company.org",
	"bob.wilson@email.net",
	"sarah.davis@corporate.io",
	"mike.brown@startup.co",
	"emma.taylor@business.com",
	"admin@potato-warehouse.internal",
	"support@freshpotatoes.com",
}

// Logger interface for OTLP logging
type Logger interface {
	Emit(ctx context.Context, record logapi.Record)
}

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
	logger  Logger
}

func NewWorker(storage storage.Storage, logger Logger) *Worker {
	return &Worker{
		storage: storage,
		logger:  logger,
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

func (w *Worker) StartPotatoRemover(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			w.removeRandomPotatoes()
		}
	}()
}

func (w *Worker) removeRandomPotatoes() {
	potatoes := w.storage.GetAllPotatoes()
	if len(potatoes) == 0 {
		return
	}

	// Pick a random variety to simulate a sale
	targetVariety := varieties[rand.Intn(len(varieties))]

	// Filter potatoes by the target variety
	var varietyPotatoes []models.Potato
	for _, p := range potatoes {
		if p.Variety == targetVariety {
			varietyPotatoes = append(varietyPotatoes, p)
		}
	}

	if len(varietyPotatoes) == 0 {
		return
	}

	// Remove only 1 potato of this variety to simulate a single sale
	potato := varietyPotatoes[rand.Intn(len(varietyPotatoes))]
	err := w.storage.DeletePotato(potato.ID)
	if err == nil {
		// Simulate a log with sensitive data (for exercise purposes)
		userEmail := fakeUserEmails[rand.Intn(len(fakeUserEmails))]
		w.logSale(potato, userEmail)
	}
}

func (w *Worker) logSale(potato models.Potato, userEmail string) {
	if w.logger == nil {
		return
	}

	ctx := context.Background()
	record := logapi.Record{}
	record.SetTimestamp(time.Now())
	record.SetSeverity(logapi.SeverityInfo)
	record.SetSeverityText("INFO")
	record.SetBody(logapi.StringValue(fmt.Sprintf(
		"Sale completed: %s potato sold (%.2fkg, $%.2f). Customer: %s",
		potato.Variety, potato.Weight, potato.Price, userEmail,
	)))

	record.AddAttributes(
		logapi.String("event.type", "sale"),
		logapi.String("potato.id", potato.ID),
		logapi.String("potato.variety", potato.Variety),
		logapi.Float64("potato.weight_kg", potato.Weight),
		logapi.Float64("potato.price", potato.Price),
		logapi.String("potato.quality", potato.Quality),
		logapi.String("customer.email", userEmail),
		logapi.Int("transaction.id", rand.Intn(99999)),
	)

	w.logger.Emit(ctx, record)
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
