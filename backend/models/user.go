package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID           int                `bson:"user_id" json:"user_id"`
	Email            string             `bson:"email" json:"email"`
	Password         string             `bson:"password" json:"password"`
	UserSongLanguage string             `bson:"user_song_language" json:"user_song_language"`
	UserName         string             `bson:"user_name" json:"user_name"`
	UserAge          int                `bson:"user_age" json:"user_age"`
	UserGender       string             `bson:"user_gender" json:"user_gender"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	UserSongLanguage string `json:"user_song_language"`
	UserName         string `json:"user_name"`
	UserAge          int    `json:"user_age"`
	UserGender       string `json:"user_gender"`
}
