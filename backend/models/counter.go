package models

type Counter struct {
	ID       string `bson:"_id" json:"_id"`
	Sequence int    `bson:"sequence" json:"sequence"`
}
