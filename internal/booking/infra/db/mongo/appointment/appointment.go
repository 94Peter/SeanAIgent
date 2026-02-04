package appointment

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
	appointmentCollectionName = "appointment"
	transformIDFailMsg        = "transform id fail: %w"
)

var appointmentCollection = mgo.NewCollectDef(appointmentCollectionName, func() []mongo.IndexModel {
	return []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "training_date_id", Value: 1}, {Key: "child_name", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "training_date_id", Value: 1}},
		},
	}
})

type apptOpt func(*appointment) error

func withApptID(id string) apptOpt {
	return func(appt *appointment) error {
		oid, err := bson.ObjectIDFromHex(id)
		if err != nil {
			return fmt.Errorf(transformIDFailMsg, err)
		}
		appt.ID = oid
		return nil
	}
}

func withDomainAppt(appt *entity.Appointment) apptOpt {
	return func(model *appointment) error {
		if appt == nil {
			return errors.New("entity is nil")
		}
		oid, err := bson.ObjectIDFromHex(appt.ID())
		if err != nil {
			return fmt.Errorf(transformIDFailMsg, err)
		}
		trainID, err := bson.ObjectIDFromHex(appt.TrainingID())
		if err != nil {
			return fmt.Errorf("transform trainID fail: %w", err)
		}
		model.ID = oid
		model.UserID = appt.User().UserID()
		model.UserName = appt.User().UserName()
		model.ChildName = appt.ChildName()
		model.Status = string(appt.Status())
		model.TrainingDateId = trainID
		model.CreatedAt = appt.CreatedAt()
		model.UpdateAt = appt.UpdateAt()
		model.IsCheckedIn = (appt.VerifiedAt() != nil)
		model.VerifyTime = appt.VerifiedAt()
		if info := appt.LeaveInfo(); !info.IsEmpty() {
			model.Leave = &leaveInfo{
				Reason:    info.Reason(),
				Status:    string(info.Status()),
				CreatedAt: info.CreatedAt(),
			}
			model.IsOnLeave = true
		} else {
			model.IsOnLeave = false
			model.Leave = nil
		}
		model.Migration.Status = mgo.MigrateStatusSuccess
		model.Migration.Version = 2
		model.Migration.LastRun = time.Now()
		model.Migration.Error = ""
		return nil
	}
}

func newModelAppt(opts ...apptOpt) (*appointment, error) {
	appt := &appointment{
		Index: appointmentCollection,
	}
	var err error
	for _, opt := range opts {
		err = opt(appt)
		if err != nil {
			return nil, fmt.Errorf("new appointment fail: %w", err)
		}
	}
	return appt, nil
}

type appointment struct {
	// v2 field
	CreatedAt time.Time `bson:"created_at"`
	mgo.Index `bson:"-"`
	Migration mgo.MigrationInfo `bson:"_migration"`
	// v2 field
	v2Fields `bson:",inline"`
	// deprecated v1 field
	v1_deprecatedFields `bson:",inline"`
	ChildName           string        `bson:"child_name,omitempty"`
	UserName            string        `bson:"user_name"`
	UserID              string        `bson:"user_id"`
	TrainingDateId      bson.ObjectID `bson:"training_date_id"`
	ID                  bson.ObjectID `bson:"_id"`
}

type v2Fields struct {
	UpdateAt   time.Time  `bson:"update_at"`
	VerifyTime *time.Time `bson:"verify_time,omitempty"`
	Leave      *leaveInfo `bson:"leave,omitempty"`
	Status     string     `bson:"status"`
}

type v1_deprecatedFields struct {
	IsCheckedIn bool `bson:"is_checked_in"`
	IsOnLeave   bool `bson:"is_on_leave"`
}

type leaveInfo struct {
	CreatedAt time.Time `bson:"created_at"`
	Reason    string    `bson:"reason"`
	Status    string    `bson:"status"`
}

func (s *appointment) toDomain() (*entity.Appointment, error) {
	user, err := entity.NewUser(s.UserID, s.UserName)
	if err != nil {
		return nil, err
	}
	status, ok := entity.AppointmentStatusFromString(s.Status)
	if !ok {
		status = entity.StatusConfirmed
	}
	leaveInfo := entity.EmptyLeaveInfo
	if s.Leave != nil {
		status, ok := entity.LeaveStatusFromString(s.Leave.Status)
		if !ok {
			return nil, fmt.Errorf("leave status is invalid: %s", s.Leave.Status)
		}
		leaveInfo = entity.NewLeaveInfo(
			s.Leave.Reason,
			status,
			s.Leave.CreatedAt)
	}

	return entity.NewAppointment(
		entity.WithApptID(s.ID.Hex()),
		entity.WithUser(user),
		entity.WithChildName(s.ChildName),
		entity.WithTrainingID(s.TrainingDateId.Hex()),
		entity.WithStatus(status),
		entity.WithCreatedAt(s.CreatedAt),
		entity.WithUpdatedAt(s.UpdateAt),
		entity.WithVerifiedAt(s.VerifyTime),
		entity.WithLeaveInfo(leaveInfo),
	)
}

func (s *appointment) GetId() any {
	if s.ID.IsZero() {
		return nil
	}
	return s.ID
}

func (s *appointment) SetId(id any) {
	oid, ok := id.(bson.ObjectID)
	if !ok {
		return
	}
	s.ID = oid
}

func (p *appointment) Validate() error {
	return nil
}

// repo impl
func (*apptRepoImpl) SaveAppointment(
	ctx context.Context, appt *entity.Appointment,
) repository.RepoError {
	const op = "save_appointment"
	modelAppt, err := newModelAppt(withDomainAppt(appt))
	if err != nil {
		return newInternalError(op, err)
	}
	_, err = mgo.Save(ctx, modelAppt)
	if err != nil {
		return newInternalError(op, err)
	}
	return nil
}

func (*apptRepoImpl) SaveManyAppointments(
	ctx context.Context, appts []*entity.Appointment,
) repository.RepoError {
	const op = "save_many_appointments"
	bulkOpts, err := mgo.NewBulkOperation(appointmentCollectionName)
	if err != nil {
		return newInternalError(op, fmt.Errorf("new bulk operation fail: %w", err))
	}
	for _, appt := range appts {
		modelAppt, err := newModelAppt(withDomainAppt(appt))
		if err != nil {
			return newInternalError(op, err)
		}

		bulkOpts = bulkOpts.InsertOne(modelAppt)
	}
	_, err = bulkOpts.Execute(ctx)
	if err != nil {
		return newInternalError(op, fmt.Errorf("execute bulk operation fail: %w", err))
	}
	return nil
}

func (*apptRepoImpl) FindApptByID(
	ctx context.Context, id string,
) (*entity.Appointment, repository.RepoError) {
	const op = "find_appt_by_id"
	modelAppt, err := newModelAppt(withApptID(id))
	if err != nil {
		return nil, newInternalError(op, err)
	}
	err = mgo.FindById(ctx, modelAppt)
	if err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			return nil, newNotFoundError(op, err)
		default:
			return nil, newInternalError(op, err)
		}
	}
	appt, err := modelAppt.toDomain()
	if err != nil {
		return nil, newInternalError(op, err)
	}
	return appt, nil
}

func (*apptRepoImpl) DeleteAppointment(
	ctx context.Context, appt *entity.Appointment,
) repository.RepoError {
	const op = "delete_appointment"
	if appt.Status() != entity.StatusCancelled {
		return newInternalError(op, entity.ErrAppointmentInvalidStatus)
	}
	modelAppt, err := newModelAppt(withDomainAppt(appt))
	if err != nil {
		return newInternalError(op, err)
	}
	_, err = mgo.DeleteById(ctx, modelAppt)
	if err != nil {
		return newInternalError(op, err)
	}
	return nil
}

func (*apptRepoImpl) FindApptsByFilter(
	ctx context.Context, filter repository.FilterAppointment,
) ([]*entity.Appointment, repository.RepoError) {
	const op = "find_appts_by_filter"

	q, repoErr := getQueryByFilterAppt(filter)
	if repoErr != nil {
		return nil, repoErr
	}
	modelAppts, _ := newModelAppt()
	results, err := mgo.Find(ctx, modelAppts, q, core.DefaultLimit)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, newNotFoundError(op, err)
		}
		return nil, newInternalError(op, err)
	}
	appts := make([]*entity.Appointment, 0, len(results))
	for _, result := range results {
		appt, err := result.toDomain()
		if err != nil {
			return nil, newInternalError(op, err)
		}
		appts = append(appts, appt)
	}
	return appts, nil
}

func getUpdateFieldFromDomain(appt *entity.Appointment) (bson.M, error) {
	oid, err := bson.ObjectIDFromHex(appt.TrainingID())
	if err != nil {
		return nil, fmt.Errorf("transform trainID fail: %w", err)
	}
	updateField := bson.M{
		"user_id":          appt.User().UserID(),
		"user_name":        appt.User().UserName(),
		"child_name":       appt.ChildName(),
		"training_date_id": oid,
		"status":           appt.Status(),
		"update_at":        appt.UpdateAt(),
	}
	if appt.VerifiedAt() != nil {
		updateField["verify_time"] = appt.VerifiedAt()
		updateField["is_checked_in"] = true
	}
	if leave := appt.LeaveInfo(); !leave.IsEmpty() {
		updateField["leave"] = leaveInfo{
			Reason:    leave.Reason(),
			Status:    string(leave.Status()),
			CreatedAt: leave.CreatedAt(),
		}
		updateField["is_on_leave"] = true
	} else {
		updateField["is_on_leave"] = false
		updateField["leave"] = nil
	}
	return updateField, nil
}

func (*apptRepoImpl) UpdateAppt(ctx context.Context, appt *entity.Appointment) repository.RepoError {
	const op = "update_appt"
	modelAppt, err := newModelAppt(withDomainAppt(appt))
	if err != nil {
		return newInternalError(op, err)
	}
	updateField, err := getUpdateFieldFromDomain(appt)
	if err != nil {
		return newInternalError(op, err)
	}
	_, err = mgo.UpdateById(ctx, modelAppt, bson.D{
		{Key: "$set", Value: updateField},
	})
	if err != nil {
		return newInternalError(op, err)
	}
	return nil
}

func (*apptRepoImpl) UpdateManyAppts(
	ctx context.Context, appts []*entity.Appointment,
) repository.RepoError {
	const op = "update_many_appts"
	bulkOpts, err := mgo.NewBulkOperation(appointmentCollectionName)
	if err != nil {
		return newInternalError(op, fmt.Errorf("new bulk operation fail: %w", err))
	}
	for _, appt := range appts {
		modelAppt, err := newModelAppt(withDomainAppt(appt))
		if err != nil {
			return newInternalError(op, err)
		}
		updateField, err := getUpdateFieldFromDomain(appt)
		if err != nil {
			return newInternalError(op, err)
		}
		bulkOpts = bulkOpts.UpdateById(modelAppt.ID, bson.D{
			{Key: "$set", Value: updateField},
		})
	}
	_, err = bulkOpts.Execute(ctx)
	if err != nil {
		return newInternalError(op, fmt.Errorf("execute bulk operation fail: %w", err))
	}
	return nil
}
