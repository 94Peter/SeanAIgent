package core

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
)

type WriteUseCase[T any, R any] interface {
	Execute(ctx context.Context, input T) (R, UseCaseError)
	Name() string
}

type ReadUseCase[T any, R any] interface {
	Execute(ctx context.Context, input T) (R, UseCaseError)
	Name() string
}

type Traceable interface {
	ToAttributes() []attribute.KeyValue
}

type Empty struct{}
