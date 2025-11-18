package service

import (
	"errors"
	"time"

	"github.com/williamdumont/potato-demo/models"
	"github.com/williamdumont/potato-demo/storage"
)

var (
	ErrInvalidPotato = errors.New("invalid potato data")
	ErrInvalidWeight = errors.New("weight must be positive")
	ErrInvalidPrice  = errors.New("price must be non-negative")
)

type PotatoService struct {
	storage storage.Storage
}

func NewPotatoService(storage storage.Storage) *PotatoService {
	return &PotatoService{
		storage: storage,
	}
}

func (s *PotatoService) CreatePotato(potato models.Potato) (models.Potato, error) {
	if err := s.validatePotato(potato); err != nil {
		return models.Potato{}, err
	}
	
	if potato.HarvestDate.IsZero() {
		potato.HarvestDate = time.Now()
	}
	
	if err := s.storage.AddPotato(potato); err != nil {
		return models.Potato{}, err
	}
	
	return potato, nil
}

func (s *PotatoService) GetPotato(id string) (models.Potato, error) {
	return s.storage.GetPotato(id)
}

func (s *PotatoService) GetAllPotatoes() []models.Potato {
	return s.storage.GetAllPotatoes()
}

func (s *PotatoService) UpdatePotato(id string, potato models.Potato) (models.Potato, error) {
	if err := s.validatePotato(potato); err != nil {
		return models.Potato{}, err
	}
	
	if err := s.storage.UpdatePotato(id, potato); err != nil {
		return models.Potato{}, err
	}
	
	return potato, nil
}

func (s *PotatoService) DeletePotato(id string) error {
	return s.storage.DeletePotato(id)
}

func (s *PotatoService) GetPotatoesByVariety(variety string) []models.Potato {
	return s.storage.GetPotatoesByVariety(variety)
}

func (s *PotatoService) GetInventorySummary() models.InventorySummary {
	potatoes := s.storage.GetAllPotatoes()
	
	varietyMap := make(map[string]*models.InventoryItem)
	totalWeight := 0.0
	totalValue := 0.0
	
	for _, potato := range potatoes {
		totalWeight += potato.Weight
		totalValue += potato.Price
		
		if item, exists := varietyMap[potato.Variety]; exists {
			item.TotalQuantity++
			item.TotalWeight += potato.Weight
			item.AveragePrice = (item.AveragePrice*float64(item.TotalQuantity-1) + potato.Price) / float64(item.TotalQuantity)
		} else {
			varietyMap[potato.Variety] = &models.InventoryItem{
				Variety:       potato.Variety,
				TotalQuantity: 1,
				TotalWeight:   potato.Weight,
				AveragePrice:  potato.Price,
			}
		}
	}
	
	byVariety := make([]models.InventoryItem, 0, len(varietyMap))
	for _, item := range varietyMap {
		byVariety = append(byVariety, *item)
	}
	
	return models.InventorySummary{
		TotalPotatoes: len(potatoes),
		TotalWeight:   totalWeight,
		TotalValue:    totalValue,
		ByVariety:     byVariety,
	}
}

func (s *PotatoService) GetAnalytics() models.PotatoAnalytics {
	potatoes := s.storage.GetAllPotatoes()
	
	if len(potatoes) == 0 {
		return models.PotatoAnalytics{}
	}
	
	varietyCount := make(map[string]int)
	totalWeight := 0.0
	premiumCount := 0
	totalValue := 0.0
	
	for _, potato := range potatoes {
		varietyCount[potato.Variety]++
		totalWeight += potato.Weight
		totalValue += potato.Price
		if potato.Quality == string(models.Premium) {
			premiumCount++
		}
	}
	
	mostPopular := ""
	maxCount := 0
	for variety, count := range varietyCount {
		if count > maxCount {
			maxCount = count
			mostPopular = variety
		}
	}
	
	return models.PotatoAnalytics{
		MostPopularVariety: mostPopular,
		AverageWeight:      totalWeight / float64(len(potatoes)),
		PremiumPercentage:  float64(premiumCount) / float64(len(potatoes)) * 100,
		TotalValue:         totalValue,
	}
}

func (s *PotatoService) CalculateFreshness(potato models.Potato) string {
	daysSinceHarvest := int(time.Since(potato.HarvestDate).Hours() / 24)
	
	switch {
	case daysSinceHarvest <= 7:
		return "Fresh"
	case daysSinceHarvest <= 30:
		return "Good"
	case daysSinceHarvest <= 90:
		return "Fair"
	default:
		return "Old"
	}
}

func (s *PotatoService) validatePotato(potato models.Potato) error {
	if potato.ID == "" || potato.Variety == "" {
		return ErrInvalidPotato
	}
	
	if potato.Weight <= 0 {
		return ErrInvalidWeight
	}
	
	if potato.Price < 0 {
		return ErrInvalidPrice
	}
	
	return nil
}

