package service

import (
	// Added context import
	"fmt"
	"seanAIgent/internal/db"
)

type Store struct {
	training    db.TrainingDateStore
	appointment db.AppointmentStore
}

func (store *Store) validate() error {
	if store.training == nil {
		return fmt.Errorf("training store is nil")
	}
	if store.appointment == nil {
		return fmt.Errorf("appointment store is nil")
	}

	return nil
}

type ServiceOption func(*Store)

func InitService(opts ...ServiceOption) Service {
	store := &Store{}
	for _, opt := range opts {
		opt(store)
	}
	err := store.validate()
	if err != nil {
		panic(err)
	}
	svcImpl := &svcImpl{
		TrainingDateService: newTrainingDateService(store.training, store.appointment),
	}
	return svcImpl
}

func WithTrainingStore(training db.TrainingDateStore) ServiceOption {
	return func(store *Store) {
		store.training = training
	}
}

func WithAppointmentStore(appointment db.AppointmentStore) ServiceOption {
	return func(store *Store) {
		store.appointment = appointment
	}
}

type Service interface {
	TrainingDateService
}

type svcImpl struct {
	TrainingDateService
}
