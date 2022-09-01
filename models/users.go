package models

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/mail"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var passwordSalt string = os.Getenv("GO_RECIPES_PWD_SALT")

type UsersService struct {
	usersCollection *mongo.Collection
}

func NewUsersService(uc *mongo.Collection) *UsersService {
	return &UsersService{
		usersCollection: uc,
	}
}

type User struct {
	Id           string `bson:"_id,omitempty"`
	Email        string `bson:"email,omitempty"`
	PasswordHash string `bson:"password_hash,omitempty"`
}

func (us *UsersService) CreateUser(requestBody io.ReadCloser) (*User, error) {
	// We create a temp user struct so we never accidentally store their password
	tempUser := struct {
		Email    string
		Password string
	}{}
	json.NewDecoder(requestBody).Decode(&tempUser)
	if tempUser.Email != "" && tempUser.Password != "" {
		// Ensure email is actually a valid email
		if !valid(tempUser.Email) {
			return nil, errors.New("Invalid email")
		}
		// We need to double check the email provided isn't already in use
		dbUserFound := User{}
		e := us.usersCollection.FindOne(context.Background(), bson.M{"email": tempUser.Email}).Decode((&dbUserFound))
		if e == nil && dbUserFound.Email == tempUser.Email {
			return nil, errors.New("Email already exists")
		}
		pwHash, err := bcrypt.GenerateFromPassword([]byte(tempUser.Password+passwordSalt), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		user := User{
			Email:        tempUser.Email,
			PasswordHash: string(pwHash),
		}
		userCreatedResult, userErr := us.usersCollection.InsertOne(context.Background(), user)
		if userErr != nil {
			return nil, errors.New("Something went wrong, unable to create User")
		} else {
			objectId, _ := userCreatedResult.InsertedID.(primitive.ObjectID)
			return &User{
				Email: user.Email,
				Id:    objectId.Hex(),
			}, nil
		}
	} else {
		return nil, errors.New("You did not provide the necessary information to create a User")
	}
}

func (us *UsersService) LoginUser(requestBody io.ReadCloser) (*User, error) {
	// We create a temp user struct so we never accidentally store their password
	tempUser := struct {
		Email    string
		Password string
	}{}
	json.NewDecoder(requestBody).Decode(&tempUser)
	if tempUser.Email != "" && tempUser.Password != "" {
		dbUserFound := User{}
		e := us.usersCollection.FindOne(context.Background(), bson.M{"email": tempUser.Email}).Decode((&dbUserFound))
		if e != nil {
			return nil, errors.New("Something went wrong, unable to find User requested")
		} else {
			hashErr := bcrypt.CompareHashAndPassword([]byte(dbUserFound.PasswordHash), []byte(tempUser.Password+passwordSalt))
			if hashErr != nil {
				// Passwords did not match
				return nil, errors.New("Incorrect password provided")
			}
			dbUserFound.PasswordHash = ""
			return &dbUserFound, nil
		}
	} else {
		// Request did not have either a user email or password
		return nil, errors.New("You did not provide all necessary information for User")
	}
}

func (us *UsersService) GetUser(userId string) (*User, error) {
	if userId != "" {
		// For MongoDb we need to create an ObjectId out of the given ID
		objectId, err := primitive.ObjectIDFromHex(userId)
		if err != nil {
			return nil, errors.New("Invalid id")
		}
		user := User{}
		err = us.usersCollection.FindOne(context.Background(), bson.M{"_id": objectId}).Decode((&user))
		if err != nil {
			return nil, errors.New("Error retrieving user by ID")
		}
		user.PasswordHash = ""
		return &user, nil
	} else {
		return nil, errors.New("Did not provide necessary information")
	}
}

func valid(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
