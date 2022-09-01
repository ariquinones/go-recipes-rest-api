package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/ariquinones/go-recipes-rest-api/models"
	"github.com/go-chi/chi/v5"
)

func NewRecipesController(rs *models.RecipesService) *RecipesController {
	return &RecipesController{
		RecipesService: rs,
	}
}

type RecipesController struct {
	RecipesService *models.RecipesService
}

func (rc *RecipesController) RecipesHandler(w http.ResponseWriter, r *http.Request) {
	recipes, err := rc.RecipesService.GetAllRecipesByUserId(chi.URLParam(r, "userId"))
	if err != nil {
		http.Error(w, "There was a problem getting all the user's recipes", http.StatusInternalServerError)
		return
	}
	jsonRecipes, err := json.Marshal(recipes)
	if err != nil {
		http.Error(w, "Error marshalling recipes", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(jsonRecipes))
}

func (rc *RecipesController) RecipeHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "recipeId")
	recipe, err := rc.RecipesService.GetRecipe(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonRecipe, err := json.Marshal(recipe)
	if err != nil {
		http.Error(w, "Error marshalling recipe", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(jsonRecipe))
}

func (rc *RecipesController) CreateRecipeHandler(w http.ResponseWriter, r *http.Request) {
	recipe, err := rc.RecipesService.CreateRecipe(chi.URLParam(r, "userId"), r.Body)
	if err != nil {
		http.Error(w, "Something went wrong, couldn't create recipe", http.StatusInternalServerError)
		return
	}
	jsonRecipe, err := json.Marshal(recipe)
	if err != nil {
		http.Error(w, "Error marshalling recipe", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(jsonRecipe))
}

func (rc *RecipesController) UpdateRecipeHandler(w http.ResponseWriter, r *http.Request) {
	recipe, err := rc.RecipesService.UpdateRecipe(chi.URLParam(r, "userId"), chi.URLParam(r, "recipeId"), r.Body)
	if err != nil {
		http.Error(w, "Something went wrong, couldn't update recipe", http.StatusInternalServerError)
		return
	}
	jsonRecipe, err := json.Marshal(recipe)
	if err != nil {
		http.Error(w, "Error marshalling recipe", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(jsonRecipe))
}

func (rc *RecipesController) DeleteRecipeHandler(w http.ResponseWriter, r *http.Request) {
	err := rc.RecipesService.DeleteRecipe(chi.URLParam(r, "recipeId"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(nil))
}

func (rc *RecipesController) UploadImageToRecipeHandler(w http.ResponseWriter, r *http.Request) {
	err := rc.RecipesService.UploadImageToRecipe(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(nil))
}
