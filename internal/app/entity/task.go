package entity

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Task struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	Duration    int                `bson:"duration" json:"duration"`
	Tag         string             `bson:"tag" json:"tag"`
	Author      string             `bson:"author" json:"author"`
	AuthorId    primitive.ObjectID `bson:"author_id" json:"author_id"`
	Date        time.Time          `bson:"date" json:"date"`
}

type TaskFilter struct {
	ID          primitive.ObjectID `json:"_id,omitempty"`
	Title       string             `json:"title,omitempty"`
	Description string             `json:"description,omitempty"`
	DurationMin int                `json:"duration_min,omitempty"`
	DurationMax int                `json:"duration_max,omitempty"`
	Tag         string             `json:"tag,omitempty"`
	Author      string             `json:"author,omitempty"`
	DateFrom    time.Time          `json:"date_from,omitempty"`
	DateTo      time.Time          `json:"date_to,omitempty"`
	AuthorId    primitive.ObjectID `json:"author_id,omitempty"`
}
