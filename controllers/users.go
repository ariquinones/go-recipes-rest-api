package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ariquinones/go-recipes-rest-api/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi/v5"
)

func NewUsersController(us *models.UsersService) *UsersController {
	return &UsersController{
		UsersService: us,
	}
}

type UsersController struct {
	UsersService *models.UsersService
}

var tokenKey string = os.Getenv("GO_RECIPES_SECRET")

func (uc *UsersController) SignUpHandler(w http.ResponseWriter, r *http.Request) {
	user, err := uc.UsersService.CreateUser(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}
	token, e := generateJWT(user.Id, user.Email)
	if e != nil {
		fmt.Fprintf(w, "Successfully created user, %v, but unable to create jwt", user)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "recipes-api-token",
		Value:    token,
		HttpOnly: false,
		Expires:  time.Now().Add(30 * time.Minute),
	})
	w.Header().Set("Token", token)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(user.Id))
}

func (uc *UsersController) LoginHandler(w http.ResponseWriter, r *http.Request) {
	user, err := uc.UsersService.LoginUser(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}
	token, e := generateJWT(user.Id, user.Email)
	if e != nil {
		fmt.Fprintf(w, "Successfully logged in user, %v, but unable to create jwt", user)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "recipes-api-token",
		Value:    token,
		HttpOnly: false,
		Expires:  time.Now().Add(30 * time.Minute),
	})
	w.Header().Set("Token", token)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(user.Id))
}

func (uc *UsersController) UserHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "userId")
	user, err := uc.UsersService.GetUser(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}
	jsonUser, err := json.Marshal(user)
	if err != nil {
		http.Error(w, "Error marshalling user", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(jsonUser))
}

func generateJWT(userId, userEmail string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user"] = userEmail
	claims["user_id"] = userId
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()
	tokenString, err := token.SignedString([]byte(tokenKey))
	if err != nil {
		fmt.Println(err.Error())
		return "", errors.New("Something went wrong creating the JWT")
	}
	return tokenString, nil
}
