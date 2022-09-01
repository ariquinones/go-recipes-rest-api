package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ariquinones/go-recipes-rest-api/controllers"
	"github.com/ariquinones/go-recipes-rest-api/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var secret string = os.Getenv("GO_RECIPES_SECRET")
var dbName string = os.Getenv("GO_DB_NAME")
var dbUrl string = os.Getenv("GO_RECIPES_DB_URL")
var port string = os.Getenv("GO_RECIPES_PORT")

func jwtMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header["Token"] != nil {
			// checking jwt here to make sure user is authorized
			token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("There was an error")
				}
				return []byte(secret), nil
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if token.Valid {
				// If we want to check the claims on the token:
				claims := token.Claims.(jwt.MapClaims)
				routeUserId := chi.URLParam(r, "userId")
				jwtUserId := claims["user_id"]
				if routeUserId != jwtUserId {
					http.Error(w, "User not authorized through token", http.StatusBadRequest)
					return
				}
				next.ServeHTTP(w, r)
				return
			}
		} else {
			http.Error(w, "No token on request, not authorized", http.StatusBadRequest)
			return
		}
	})
}

func main() {
	client, err := mongo.NewClient(options.Client().ApplyURI(dbUrl))
	if err != nil {
		panic(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(ctx)

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		panic(err)
	}
	myDatabase := client.Database(dbName)
	r := chi.NewRouter()
	// CORS - For dev use only
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
		AllowedHeaders: []string{"Token", "Access-Control-Allow-Origin", "Content-Type", "Authorization"},
		AllowedMethods: []string{"GET", "OPTION", "PUT", "POST", "DELETE", "HEAD"},
		ExposedHeaders: []string{"Token", "Content-Type", "Authorization"},
	}))

	r.Get("/", homeHandler)
	imagesHandler := http.FileServer(http.Dir("./recipe-images"))
	r.Handle("/recipe-images/*", http.StripPrefix("/recipe-images/", imagesHandler))
	usersService := models.NewUsersService(myDatabase.Collection("users"))
	usersController := controllers.NewUsersController(usersService)

	recipesService := models.NewRecipesService(myDatabase.Collection("recipes"))
	recipesController := controllers.NewRecipesController(recipesService)

	r.Route("/users", func(r chi.Router) {
		r.Post("/signup", usersController.SignUpHandler)
		r.Post("/login", usersController.LoginHandler)
		r.Mount("/{userId}", recipesRoutes(usersController, recipesController))
	})
	r.NotFound(notFoundHandler)
	http.ListenAndServe(port, r)
}

func recipesRoutes(uc *controllers.UsersController, rc *controllers.RecipesController) chi.Router {
	r := chi.NewRouter()
	r.Use(jwtMiddleware)
	r.Get("/", uc.UserHandler)
	r.Route("/recipes", func(r chi.Router) {
		r.Get("/", rc.RecipesHandler)
		r.Post("/", rc.CreateRecipeHandler)
		r.Get("/{recipeId}", rc.RecipeHandler)
		r.Delete("/{recipeId}", rc.DeleteRecipeHandler)
		r.Put("/{recipeId}", rc.UpdateRecipeHandler)
		r.Post("/{recipeId}/images", rc.UploadImageToRecipeHandler)
	})
	return r
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "HELLO WORLD")
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Request URL not found", http.StatusNotFound)
}
