package service

import (
	"context"
	"errors"
	"github.com/ArtemevDenis/time-tracker/internal/app/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Service struct {
	db *mongo.Database
}

func New(db *mongo.Database) *Service {
	return &Service{
		db: db,
	}
}

func (s *Service) CreateTask(task *entity.Task) (t *entity.Task, err error) {
	var res *mongo.InsertOneResult
	res, err = s.db.Collection("tasks").InsertOne(context.TODO(), task)

	if err != nil {
		return nil, err
	}

	t = &entity.Task{
		ID:          res.InsertedID.(primitive.ObjectID),
		Title:       task.Title,
		Description: task.Description,
		Duration:    task.Duration,
		Tag:         task.Tag,
		Author:      task.Author,
		Date:        task.Date,
		AuthorId:    task.AuthorId,
	}
	return t, nil
}

func (s *Service) GetTasks(filter *entity.TaskFilter) ([]entity.Task, error) {
	bsonFilter := bson.M{}

	if filter.ID != primitive.NilObjectID {
		bsonFilter["_id"] = filter.ID
	}
	if filter.AuthorId != primitive.NilObjectID {
		bsonFilter["author_id"] = filter.AuthorId
	}
	if filter.Title != "" {
		bsonFilter["title"] = bson.M{"$regex": filter.Title, "$options": "i"}
	}
	if filter.Description != "" {
		bsonFilter["description"] = bson.M{"$regex": filter.Description, "$options": "i"}
	}
	if filter.DurationMin != 0 || filter.DurationMax != 0 {
		durationFilter := bson.M{}
		if filter.DurationMin != 0 {
			durationFilter["$gte"] = filter.DurationMin
		}
		if filter.DurationMax != 0 {
			durationFilter["$lte"] = filter.DurationMax
		}
		bsonFilter["duration"] = durationFilter
	}
	if filter.Tag != "" {
		bsonFilter["tag"] = filter.Tag
	}
	if filter.Author != "" {
		bsonFilter["author"] = filter.Author
	}

	if !filter.DateFrom.IsZero() || !filter.DateTo.IsZero() {
		dateFilter := bson.M{}
		if !filter.DateFrom.IsZero() {
			dateFilter["$gte"] = filter.DateFrom
		}
		if !filter.DateTo.IsZero() {
			dateFilter["$lte"] = filter.DateTo
		}
		bsonFilter["date"] = dateFilter
	}

	cursor, err := s.db.Collection("tasks").Find(context.TODO(), bsonFilter)

	if err != nil {
		return nil, err
	}

	defer cursor.Close(context.TODO())

	var tasks []entity.Task

	if err = cursor.All(context.TODO(), &tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s *Service) UpdateTask(task *entity.Task, authorId *primitive.ObjectID) (t *entity.Task, err error) {
	_, err = s.db.Collection("tasks").UpdateOne(context.TODO(), bson.M{"_id": task.ID, "author_id": authorId}, bson.M{"$set": task})
	if err != nil {
		return nil, err
	}

	var result entity.Task
	err = s.db.Collection("tasks").FindOne(context.TODO(), bson.D{{"_id", task.ID}}).Decode(&result)

	if err != nil {
		return nil, err
	}
	t = &entity.Task{
		ID:          result.ID,
		Title:       result.Title,
		Description: result.Description,
		Duration:    result.Duration,
		Tag:         result.Tag,
		Author:      result.Author,
		Date:        result.Date,
	}

	return t, nil
}

func (s *Service) DeleteTask(id *primitive.ObjectID, authorId *primitive.ObjectID) (err error) {
	res, err := s.db.Collection("tasks").DeleteOne(context.TODO(), bson.D{{"_id", id}, {"author_id", authorId}})
	if err != nil {
		return err
	}

	if res.DeletedCount == 0 {
		return errors.New("nothing to delete")
	}
	return nil
}

func (s *Service) GetUserByEmail(email *string) (*entity.User, error) {
	user := entity.User{}
	err := s.db.Collection("users").FindOne(context.TODO(), bson.M{"email": *email}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
