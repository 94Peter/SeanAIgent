package mongodb

import (
	"context"
	"errors"
	"seanAIgent/internal/db"
	"seanAIgent/internal/db/model"
	"time"

	"github.com/94peter/vulpes/db/mgo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func NewTrainingDateStore() db.TrainingDateStore {
	return &trainingDateStore{}
}

type trainingDateStore struct{}

func (es *trainingDateStore) Find(ctx context.Context, q bson.M) ([]*model.TrainingDate, error) {
	return mgo.Find(ctx, model.NewTrainingDate(), q)
}

func (es *trainingDateStore) Add(ctx context.Context, trainingDate *model.TrainingDate) error {
	_, err := mgo.Save(ctx, trainingDate)
	return err
}

func (es *trainingDateStore) AddMany(ctx context.Context, trainingDates []*model.TrainingDate) error {
	bulk, err := mgo.NewBulkOperation(model.TrainingDateCollectionName)
	if err != nil {
		return err
	}
	for _, trainingDate := range trainingDates {
		bulk = bulk.InsertOne(trainingDate)
	}
	_, err = bulk.Execute(ctx)
	return err
}

func (es *trainingDateStore) Delete(ctx context.Context, id string) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	trainingDate := model.NewTrainingDate()
	trainingDate.ID = oid
	_, err = mgo.DeleteById(ctx, trainingDate)
	return err
}

func (es *trainingDateStore) QueryTrainingDateAppointmentState(ctx context.Context, id string, q bson.M) ([]*model.AggrTrainingDateAppointState, error) {
	return mgo.PipeFind(ctx, model.NewAggrTrainingDateAppointState(id), q)
}

func (es *trainingDateStore) QueryTrainingDataHasAppointment(ctx context.Context, q bson.M) ([]*model.AggrTrainingDateHasAppoint, error) {
	return mgo.PipeFind(ctx, model.NewAggrTrainingDateHasAppoint(), q)
}

func (es *trainingDateStore) QueryTrainingDateHasCheckinList(ctx context.Context, now time.Time) (*model.AggrTrainingdateHasCheckinItems, error) {
	q := bson.M{"start_date": bson.D{{Key: "$lte", Value: now}}, "end_date": bson.D{{Key: "$gte", Value: now}}}
	results := model.NewAggrTrainingdateHasCheckinItems()
	err := mgo.PipeFindOne(ctx, results, q)
	if err != nil {
		switch {
		case errors.Is(err, mongo.ErrNoDocuments):
			return nil, errorWrap(db.ErrNotFound, err)
		}
		return nil, errorWrap(db.ErrReadFailed, err)
	}
	return results, nil
}
