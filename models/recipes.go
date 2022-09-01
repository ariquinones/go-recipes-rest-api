package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type RecipesService struct {
	recipesCollection *mongo.Collection
}

func NewRecipesService(rc *mongo.Collection) *RecipesService {
	return &RecipesService{
		recipesCollection: rc,
	}
}

type Recipe struct {
	Id           string       `json:"id,omitempty" bson:"_id,omitempty"`
	Name         string       `json:"name,omitempty" bson:"name,omitempty"`
	Price        float64      `json:"price,omitempty" bson:"price,omitempty"`
	Description  string       `json:"description,omitempty" bson:"description,omitempty"`
	Image        string       `json:"image,omitempty" son:"image,omitempty"`
	Yield        string       `json:"yield,omitempty" bson:"yield,omitempty"`
	Instructions []string     `json:"instructions,omitempty" bson:"instructions,omitempty"`
	Ingredients  []Ingredient `json:"ingredients,omitempty" bson:"ingredients,omitempty"`
	User         string       `json:"user,omitempty" bson:"user,omitempty"`
}

type Ingredient struct {
	Name        string `json:"name,omitempty" bson:"name,omitempty"`
	Preparation string `json:"preparation,omitempty" bson:"preparation,omitempty"`
	Cost        string `json:"cost,omitempty" bson:"cost,omitempty"`
	Amount      string `json:"amount,omitempty" bson:"amount,omitempty"`
}

func (rs *RecipesService) GetRecipe(id string) (*Recipe, error) {
	if id != "" {
		// For MongoDb we need to create an ObjectId out of the given ID
		objectId, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, errors.New("Invalid id")
		}
		recipe := Recipe{}
		err = rs.recipesCollection.FindOne(context.Background(), bson.M{"_id": objectId}).Decode((&recipe))
		if err != nil {
			return nil, errors.New("Error retrieving recipe by ID")
		}
		return &recipe, nil
	} else {
		return nil, errors.New("Did not provide necessary information")
	}
}

func (rs *RecipesService) CreateRecipe(userId string, requestBody io.ReadCloser) (*Recipe, error) {
	recipe := Recipe{}
	json.NewDecoder(requestBody).Decode(&recipe)
	recipe.User = userId
	insertResult, insertErr := rs.recipesCollection.InsertOne(context.Background(), recipe)
	if insertErr != nil {
		return nil, insertErr
	} else {
		objectId, _ := insertResult.InsertedID.(primitive.ObjectID)
		recipe.Id = objectId.Hex()
		return &recipe, nil
	}
}

func (rs *RecipesService) UpdateRecipe(userId, recipeId string, requestBody io.ReadCloser) (*Recipe, error) {
	recipe := Recipe{}
	json.NewDecoder(requestBody).Decode(&recipe)
	objectId, err := primitive.ObjectIDFromHex(recipeId)
	recipe.User = userId
	// We get rid of any possible ID the body recipe may contain as mongoDb will error if the id is passed in
	recipe.Id = ""
	if err != nil {
		return nil, errors.New("Recipe has invalid id")
	}
	singleResult := rs.recipesCollection.FindOneAndReplace(context.Background(), bson.M{"_id": objectId}, recipe)
	if singleResult.Err() != nil {
		return nil, singleResult.Err()
	}
	recipe.Id = recipeId
	return &recipe, nil
}

func (rs *RecipesService) DeleteRecipe(recipeId string) error {
	objectId, err := primitive.ObjectIDFromHex(recipeId)
	if err != nil {
		return errors.New("Invalid recipe id")
	}
	result := rs.recipesCollection.FindOneAndDelete(context.Background(), bson.M{"_id": objectId})
	if result.Err() != nil {
		return result.Err()
	}
	return nil
}

func (rs *RecipesService) GetAllRecipesByUserId(userId string) (*[]Recipe, error) {
	list := []Recipe{}
	cur, err := rs.recipesCollection.Find(context.Background(), bson.M{"user": userId})
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.Background())
	for cur.Next(context.Background()) {
		// To decode into a struct, use cursor.Decode()
		result := Recipe{}
		err := cur.Decode(&result)
		if err != nil {
			return nil, err
		}
		list = append(list, result)
	}
	return &list, nil
}

func (rs *RecipesService) UploadImageToRecipe(r *http.Request) error {
	recipeId := chi.URLParam(r, "recipeId")
	recipe, err := rs.GetRecipe(recipeId)
	if err != nil {
		return errors.New("Unable to find recipe")
	}
	// if we find the recipe we upload image to server file-system and then update recipe's image field with name of image
	r.ParseMultipartForm(10 << 20) // 10 MB
	file, fileHandler, err := r.FormFile("file")
	if err != nil {
		fmt.Println(err)
		return errors.New("Unable to parse form to get image file")
	}
	defer file.Close()
	// create directory if necessary
	path := "./recipe-images/"
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			fmt.Println(err)
		}
	}
	f, err := os.OpenFile("./recipe-images/"+fileHandler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		return errors.New("Unable to get files directory")
	}
	io.Copy(f, file)
	recipe.Image = fileHandler.Filename
	objectId, err := primitive.ObjectIDFromHex(recipeId)
	// We get rid of any possible ID the body recipe may contain as mongoDb will error if the id is passed in
	recipe.Id = ""
	if err != nil {
		return errors.New("Unable to update recipe with its uploaded image")
	}
	singleResult := rs.recipesCollection.FindOneAndReplace(context.Background(), bson.M{"_id": objectId}, recipe)
	if singleResult.Err() != nil {
		return singleResult.Err()
	}
	return nil
}
