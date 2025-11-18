package storage

import (
	"errors"
	"sync"

	"github.com/williamdumont/potato-demo/models"
)

var (
	ErrNotFound      = errors.New("potato not found")
	ErrRecipeNotFound = errors.New("recipe not found")
)

type Storage interface {
	AddPotato(potato models.Potato) error
	GetPotato(id string) (models.Potato, error)
	GetAllPotatoes() []models.Potato
	UpdatePotato(id string, potato models.Potato) error
	DeletePotato(id string) error
	GetPotatoesByVariety(variety string) []models.Potato
	
	AddRecipe(recipe models.Recipe) error
	GetRecipe(id string) (models.Recipe, error)
	GetAllRecipes() []models.Recipe
	GetRecipesByVariety(variety string) []models.Recipe
}

type InMemoryStorage struct {
	potatoes map[string]models.Potato
	recipes  map[string]models.Recipe
	mu       sync.RWMutex
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		potatoes: make(map[string]models.Potato),
		recipes:  make(map[string]models.Recipe),
	}
}

func (s *InMemoryStorage) AddPotato(potato models.Potato) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.potatoes[potato.ID] = potato
	return nil
}

func (s *InMemoryStorage) GetPotato(id string) (models.Potato, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	potato, exists := s.potatoes[id]
	if !exists {
		return models.Potato{}, ErrNotFound
	}
	return potato, nil
}

func (s *InMemoryStorage) GetAllPotatoes() []models.Potato {
	s.mu.RLock()
	defer s.mu.RUnlock()
	potatoes := make([]models.Potato, 0, len(s.potatoes))
	for _, potato := range s.potatoes {
		potatoes = append(potatoes, potato)
	}
	return potatoes
}

func (s *InMemoryStorage) UpdatePotato(id string, potato models.Potato) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.potatoes[id]; !exists {
		return ErrNotFound
	}
	s.potatoes[id] = potato
	return nil
}

func (s *InMemoryStorage) DeletePotato(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.potatoes[id]; !exists {
		return ErrNotFound
	}
	delete(s.potatoes, id)
	return nil
}

func (s *InMemoryStorage) GetPotatoesByVariety(variety string) []models.Potato {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var potatoes []models.Potato
	for _, potato := range s.potatoes {
		if potato.Variety == variety {
			potatoes = append(potatoes, potato)
		}
	}
	return potatoes
}

func (s *InMemoryStorage) AddRecipe(recipe models.Recipe) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recipes[recipe.ID] = recipe
	return nil
}

func (s *InMemoryStorage) GetRecipe(id string) (models.Recipe, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	recipe, exists := s.recipes[id]
	if !exists {
		return models.Recipe{}, ErrRecipeNotFound
	}
	return recipe, nil
}

func (s *InMemoryStorage) GetAllRecipes() []models.Recipe {
	s.mu.RLock()
	defer s.mu.RUnlock()
	recipes := make([]models.Recipe, 0, len(s.recipes))
	for _, recipe := range s.recipes {
		recipes = append(recipes, recipe)
	}
	return recipes
}

func (s *InMemoryStorage) GetRecipesByVariety(variety string) []models.Recipe {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var recipes []models.Recipe
	for _, recipe := range s.recipes {
		if recipe.Variety == variety {
			recipes = append(recipes, recipe)
		}
	}
	return recipes
}

