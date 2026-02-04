package appointment

import (
	"context"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"

	"github.com/94peter/vulpes/db/mgo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// 同時存資料到v1的database
type migrateV1ToV2apptRepoImpl struct {
	repository.AppointmentRepository
}

// 請假主要只透過updateAppt，所以只在updateAppt時才會存到v1的leave database
func (m *migrateV1ToV2apptRepoImpl) UpdateAppt(
	ctx context.Context, appt *entity.Appointment,
) repository.RepoError {
	bookid, err := bson.ObjectIDFromHex(appt.ID())
	if err != nil {
		return repository.NewRepoInvalidDocumentIDError("migrateV1ToV2apptRepoImpl", "updateAppt", "invalid appt id", err)
	}
	var rollbackFunc func() error
	if appt.LeaveInfo().IsEmpty() {
		// 刪除leave
		leav := newLeave()
		err = mgo.FindOne(ctx, leav, bson.M{"booking_id": bookid})
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return m.AppointmentRepository.UpdateAppt(ctx, appt)
			}
			return newInternalError("migrateV1ToV2apptRepoImpl_updateAppt", err)
		}
		// 有資料才刪除
		if !leav.ID.IsZero() {
			_, err = mgo.DeleteById(ctx, leav)
			if err != nil {
				return newInternalError("migrateV1ToV2apptRepoImpl_updateAppt", err)
			}
			rollbackFunc = func() error {
				_, err = mgo.Save(ctx, leav)
				return err
			}
		}
	} else {
		// 新增或更新leave
		leav := newLeave()
		leav.ID = bson.NewObjectID()
		leav.UserID = appt.User().UserID()
		leav.ChildName = appt.ChildName()
		leav.Reason = appt.LeaveInfo().Reason()
		leav.Status = v2StatusToV1Status(appt)
		leav.BookingID = bookid
		leav.UpdatedAt = appt.UpdateAt()
		leav.CreatedAt = appt.LeaveInfo().CreatedAt()
		_, err = mgo.Save(ctx, leav)
		if err != nil {
			return newInternalError("migrateV1ToV2apptRepoImpl_updateAppt", err)
		}
		rollbackFunc = func() error {
			_, err = mgo.DeleteById(ctx, leav)
			return err
		}
	}
	err = m.AppointmentRepository.UpdateAppt(ctx, appt)
	if err != nil {
		if rollbackFunc != nil {
			_ = rollbackFunc()
		}
		return newInternalError("migrateV1ToV2apptRepoImpl_updateAppt", err)
	}
	return nil
}

func v2StatusToV1Status(appt *entity.Appointment) LeaveStatus {
	switch appt.LeaveInfo().Status() {
	case entity.LeaveStatusNone:
		return LeaveStatusNone
	case entity.LeaveStatusApproved:
		return LeaveStatusApproved
	case entity.LeaveStatusRejected:
		return LeaveStatusRejected
	case entity.LeaveStatusPending:
		return LeaveStatusPending
	default:
		return LeaveStatusNone
	}

}
