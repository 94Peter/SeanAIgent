package train

import (
	"context"
	"errors"
	"fmt"
	"time"

	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/infra/db/mongo/core"

	"github.com/94peter/vulpes/db/mgo"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	TrainDateCollectionName = "training_date"

	transformIDFailMsg = "transform id fail: %w"
)

var trainingDateCollection = mgo.NewCollectDef(TrainDateCollectionName, func() []mongo.IndexModel {
	return []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "start_date", Value: 1}, {Key: "end_date", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}
})

type trainingDateOpt func(*trainDate) error

func withDomainTrainDate(training *entity.TrainDate) trainingDateOpt {
	return func(td *trainDate) error {
		oid, err := bson.ObjectIDFromHex(training.ID())
		if err != nil {
			return fmt.Errorf(transformIDFailMsg, err)
		}
		td.ID = oid
		td.UserID = training.UserID()
		td.Date = training.Period().Start().Format("2006-01-02")
		td.Location = training.Location()
		td.Capacity = training.MaxCapacity()
		td.AvailableCapacity = training.AvailableCapacity()
		td.StartDate = training.Period().Start()
		td.EndDate = training.Period().End()
		td.Timezone = training.Period().Start().Location().String()
		td.Status = string(training.Status())
		td.CreatedAt = training.CreatedAt()
		td.UpdatedAt = training.UpdatedAt()
		return nil
	}
}

func withTrainDateID(id string) trainingDateOpt {
	return func(td *trainDate) error {
		oid, err := bson.ObjectIDFromHex(id)
		if err != nil {
			return fmt.Errorf(transformIDFailMsg, err)
		}
		td.ID = oid
		return nil
	}
}

func newTrainDate(opts ...trainingDateOpt) (*trainDate, error) {
	td := &trainDate{
		Index: trainingDateCollection,
	}
	var err error
	for _, opt := range opts {
		err = opt(td)
		if err != nil {
			return nil, fmt.Errorf("new train date fail: %w", err)
		}
	}
	if td.CreatedAt.IsZero() {
		td.CreatedAt = time.Now()
	}
	if td.UpdatedAt.IsZero() {
		td.UpdatedAt = td.CreatedAt
	}
	td.Migration.Status = mgo.MigrateStatusSuccess
	td.Migration.Version = 2
	td.Migration.LastRun = time.Now()
	td.Migration.Error = ""
	return td, nil
}

type trainDate struct {
	StartDate         time.Time `bson:"start_date"`
	UpdatedAt         time.Time `bson:"updated_at"`
	CreatedAt         time.Time `bson:"created_at"`
	EndDate           time.Time `bson:"end_date"`
	mgo.Index         `bson:"-"`
	Migration         mgo.MigrationInfo `bson:"_migration"`
	Timezone          string            `bson:"timezone"`
	Location          string            `bson:"location"`
	Date              string            `bson:"date"`
	Status            string            `bson:"status"`
	UserID            string            `bson:"user_id"`
	AvailableCapacity int               `bson:"available_capacity"`
	Capacity          int               `bson:"capacity"`
	ID                bson.ObjectID     `bson:"_id"`
}

func (s *trainDate) toDomain() (*entity.TrainDate, error) {
	timeRange, err := entity.NewTimeRange(s.StartDate, s.EndDate)
	if err != nil {
		return nil, err
	}
	trainDate, err := entity.NewTrainDate(
		entity.WithTrainDateID(s.ID.Hex()),
		entity.WithTrainDateUserID(s.UserID),
		entity.WithTrainDateLocation(s.Location),
		entity.WithTrainDateMaxCapacity(s.Capacity),
		entity.WithTrainDateAvailableCapacity(s.AvailableCapacity),
		entity.WithTrainDatePeriod(timeRange),
		entity.WithTrainDateCreatedAt(s.CreatedAt),
		entity.WithTrainDateUpdatedAt(s.UpdatedAt),
		entity.WithTrainDateTimezone(s.Timezone),
	)
	if err != nil {
		return nil, err
	}
	return trainDate, nil
}

func (s *trainDate) GetId() any {
	if s.ID.IsZero() {
		return nil
	}
	return s.ID
}

func (s *trainDate) SetId(id any) {
	oid, ok := id.(bson.ObjectID)
	if !ok {
		return
	}
	s.ID = oid
}

func (p *trainDate) Validate() error {
	return nil
}

func (*trainRepoImpl) SaveTrainDate(
	ctx context.Context, training *entity.TrainDate,
) repository.RepoError {
	const op = "save_train_date"
	modelTraining, err := newTrainDate(withDomainTrainDate(training))
	if err != nil {
		return newInternalError(op, err)
	}
	_, err = mgo.Save(ctx, modelTraining)
	if err != nil {
		return newInternalError(op, err)
	}
	return nil
}

func (*trainRepoImpl) SaveManyTrainDates(
	ctx context.Context, trainings []*entity.TrainDate,
) repository.RepoError {
	const op = "save_many_train_dates"
	bulkOpts, err := mgo.NewBulkOperation(TrainDateCollectionName)
	if err != nil {
		return newInternalError(op, fmt.Errorf("new bulk operation fail: %w", err))
	}
	for _, training := range trainings {
		modelTraining, err := newTrainDate(withDomainTrainDate(training))
		if err != nil {
			return newInternalError(op, err)
		}
		bulkOpts = bulkOpts.InsertOne(modelTraining)
	}
	_, err = bulkOpts.Execute(ctx)
	if err != nil {
		return newInternalError(op, fmt.Errorf("execute bulk operation fail: %w", err))
	}
	return nil
}

func (*trainRepoImpl) DeleteTrainingDate(
	ctx context.Context, training *entity.TrainDate,
) repository.RepoError {
	const op = "delete_training_date"
	modelTraining, err := newTrainDate(withDomainTrainDate(training))
	if err != nil {
		return newInternalError(op, err)
	}
	_, err = mgo.DeleteById(ctx, modelTraining)
	if err != nil {
		return newInternalError(op, err)
	}
	return nil
}

func (*trainRepoImpl) FindTrainDateByID(
	ctx context.Context, id string,
) (*entity.TrainDate, repository.RepoError) {
	const op = "find_train_date_by_id"
	doc, err := newTrainDate(withTrainDateID(id))
	if err != nil {
		return nil, newInternalError(op, err)
	}
	err = mgo.FindById(ctx, doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, newNotFoundError(op, err)
		}
		return nil, newInternalError(op, fmt.Errorf("mgo find fail: %w", err))
	}
	trainDate, err := doc.toDomain()
	if err != nil {
		return nil, newInternalError(op, fmt.Errorf("trans to domain fail: %w", err))
	}
	return trainDate, nil
}

func (*trainRepoImpl) FindTrainDates(
	ctx context.Context,
	filter repository.FilterTrainDate,
) ([]*entity.TrainDate, repository.RepoError) {
	const op = "find_train_dates"
	q, repoErr := getQueryByFilterTrainDate(filter)
	if repoErr != nil {
		return nil, repoErr
	}
	modelTrainDates, err := newTrainDate()
	if err != nil {
		return nil, newInternalError(op, err)
	}
	docs, err := mgo.Find(ctx, modelTrainDates, q, core.DefaultLimit)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, newNotFoundError(op, err)
		}
		return nil, newInternalError(op, fmt.Errorf("mgo find fail: %w", err))
	}
	trainings := make([]*entity.TrainDate, 0, len(docs))
	for _, doc := range docs {
		training, err := doc.toDomain()
		if err != nil {
			return nil, newInternalError(op, fmt.Errorf("trans to domain fail: %w", err))
		}
		trainings = append(trainings, training)
	}
	return trainings, nil
}

func (*trainRepoImpl) CheckOverlap(
	ctx context.Context, coachID string, tr entity.TimeRange,
) (bool, repository.RepoError) {
	const op = "check_overlap"
	filter := bson.M{
		"user_id":    coachID,
		"start_date": bson.M{"$lt": tr.End()},
		"end_date":   bson.M{"$gt": tr.Start()},
	}
	count, err := mgo.CountDocument(ctx, TrainDateCollectionName, filter)
	if err != nil {
		return false, newInternalError(op, fmt.Errorf("count document fail: %w", err))
	}
	return count > 0, nil
}

func (*trainRepoImpl) HasAnyOverlap(
	ctx context.Context, coachID string, tr []entity.TimeRange,
) (bool, repository.RepoError) {
	if len(tr) == 0 {
		return false, nil
	}
	const op = "has_any_overlap"
	// 構建所有的時間判斷條件
	orConditions := make([]bson.M, 0, len(tr))
	for _, t := range tr {
		orConditions = append(orConditions, bson.M{
			"start_date": bson.M{"$lt": t.End()},
			"end_date":   bson.M{"$gt": t.Start()},
		})
	}

	filter := bson.M{
		"user_id": coachID,
		"$or":     orConditions,
	}

	count, err := mgo.CountDocument(ctx, TrainDateCollectionName, filter)
	if err != nil {
		return false, newInternalError(op, fmt.Errorf("count document fail: %w", err))
	}
	return count > 0, nil
}

func (*trainRepoImpl) DeductCapacity(
	ctx context.Context, trainingID string, count int,
) repository.RepoError {
	const op = "deduct_capacity"
	oid, err := bson.ObjectIDFromHex(trainingID)
	if err != nil {
		return newInvalidDocumentIDError(op, err)
	}
	filter := bson.D{
		{Key: "_id", Value: oid},
		{Key: "available_capacity", Value: bson.M{"$gte": count}},
	}
	update := bson.D{
		{Key: "$inc", Value: bson.D{{Key: "available_capacity", Value: -count}}},
		{Key: "$set", Value: bson.D{{Key: "updated_at", Value: time.Now()}}},
	}
	doc, _ := newTrainDate()
	_, err = mgo.UpdateOne(ctx, doc, filter, update)
	if err != nil {
		return newInternalError(op, err)
	}
	return nil
}

func (*trainRepoImpl) IncreaseCapacity(
	ctx context.Context, trainingID string, count int,
) repository.RepoError {
	const op = "increase_capacity"
	oid, err := bson.ObjectIDFromHex(trainingID)
	if err != nil {
		return newInvalidDocumentIDError(op, err)
	}
	filter := bson.D{
		{Key: "_id", Value: oid},
	}
	update := bson.D{
		{Key: "$inc", Value: bson.D{{Key: "available_capacity", Value: count}}},
		{Key: "$set", Value: bson.D{{Key: "updated_at", Value: time.Now()}}},
	}
	doc, _ := newTrainDate()
	_, err = mgo.UpdateOne(ctx, doc, filter, update)
	if err != nil {
		return newInternalError(op, err)
	}
	return nil
}

func (*trainRepoImpl) UpdateManyTrainDates(
	ctx context.Context, trainings []*entity.TrainDate,
) repository.RepoError {
	const op = "update_many_train_dates"
	bulkOpts, err := mgo.NewBulkOperation(TrainDateCollectionName)
	if err != nil {
		return newInternalError(op, fmt.Errorf("new bulk operation fail: %w", err))
	}
	for _, training := range trainings {
		modelTraining, err := newTrainDate(withDomainTrainDate(training))
		if err != nil {
			return newInternalError(op, err)
		}
		updateField := getUpdateFieldFromModel(modelTraining)
		bulkOpts = bulkOpts.UpdateById(modelTraining.ID, bson.D{
			{Key: "$set", Value: updateField},
		})
	}
	_, err = bulkOpts.Execute(ctx)
	if err != nil {
		return newInternalError(op, fmt.Errorf("execute bulk operation fail: %w", err))
	}
	return nil
}

func getUpdateFieldFromModel(training *trainDate) bson.M {
	updateField := bson.M{
		"user_id":            training.UserID,
		"location":           training.Location,
		"capacity":           training.Capacity,
		"available_capacity": training.AvailableCapacity,
		"start_date":         training.StartDate,
		"end_date":           training.EndDate,
		"timezone":           training.Timezone,
		"status":             training.Status,
		"created_at":         training.CreatedAt,
		"updated_at":         training.UpdatedAt,
	}
	return updateField
}
