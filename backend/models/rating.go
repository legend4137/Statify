package models

type Rating struct {
	UserID  int `bson:"user_id"`
	TrackID int    `bson:"track_id"`
	Rating  int    `bson:"rating"`
}