package service

import (
	"errors"

	"github.com/williamdumont/potato-demo/models"
	"github.com/williamdumont/potato-demo/storage"
)

var (
	ErrInvalidRecipe = errors.New("invalid recipe data")
)

type RecipeService struct {
	storage storage.Storage
}

func NewRecipeService(storage storage.Storage) *RecipeService {
	return &RecipeService{
		storage: storage,
	}
}

func (s *RecipeService) CreateRecipe(recipe models.Recipe) (models.Recipe, error) {
	if err := s.validateRecipe(recipe); err != nil {
		return models.Recipe{}, err
	}
	
	if err := s.storage.AddRecipe(recipe); err != nil {
		return models.Recipe{}, err
	}
	
	return recipe, nil
}

func (s *RecipeService) GetRecipe(id string) (models.Recipe, error) {
	return s.storage.GetRecipe(id)
}

func (s *RecipeService) GetAllRecipes() []models.Recipe {
	return s.storage.GetAllRecipes()
}

func (s *RecipeService) GetRecipesByVariety(variety string) []models.Recipe {
	return s.storage.GetRecipesByVariety(variety)
}

func (s *RecipeService) RecommendRecipe(variety string, difficulty string) (models.Recipe, error) {
	recipes := s.storage.GetRecipesByVariety(variety)
	
	for _, recipe := range recipes {
		if difficulty == "" || recipe.Difficulty == difficulty {
			return recipe, nil
		}
	}
	
	if len(recipes) > 0 {
		return recipes[0], nil
	}
	
	return models.Recipe{}, errors.New("no recipes found for variety")
}

func (s *RecipeService) validateRecipe(recipe models.Recipe) error {
	if recipe.ID == "" || recipe.Name == "" || recipe.Variety == "" {
		return ErrInvalidRecipe
	}
	
	if recipe.CookingTime <= 0 {
		return errors.New("cooking time must be positive")
	}
	
	return nil
}

