package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type User struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID           int                `bson:"user_id" json:"user_id"`
	Email            string             `bson:"email" json:"email"`
	Password         string             `bson:"password" json:"password"`
	UserSongLanguage string             `bson:"user_song_language" json:"user_song_language"`
	UserName         string             `bson:"user_name" json:"user_name"`
	UserAge          int                `bson:"user_age" json:"user_age"`
	UserGender       string             `bson:"user_gender" json:"user_gender"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time          `bson:"updated_at" json:"updated_at"`
}

type UserActivity struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID             int                `bson:"user_id" json:"user_id"`
	Tracks             []int           `bson:"tracks" json:"tracks"`
	Mood_Energy        float64                `bson:"mood_energy" json:"mood_energy"`
	Mood_Valence       float64                `bson:"mood_valence" json:"mood_valence"`
	Preferred_Genre    string             `bson:"preferred_genre" json:"preferred_genre"`
	Preferred_Language string             `bson:"preferred_language" json:"preferred_language"`
	Age                int                `bson:"user_age" json:"user_age"`
	Language           string             `bson:"user_langugage" json:"user_language"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time          `bson:"updated_at" json:"updated_at"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	UserSongLanguage string `json:"user_song_language"`
	UserGenre		 string `json:"user_preferred_genre"`
	UserName         string `json:"user_name"`
	UserAge          int    `json:"user_age"`
	UserGender       string `json:"user_gender"`
}
