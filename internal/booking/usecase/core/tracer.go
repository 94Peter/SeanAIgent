package core

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type writeOtelDecorator[T any, R any] struct {
	next   WriteUseCase[T, R]
	tracer trace.Tracer
}

func (d *writeOtelDecorator[T, R]) Name() string {
	return d.next.Name()
}

func (d *writeOtelDecorator[T, R]) Execute(ctx context.Context, input T) (R, UseCaseError) {
	attrs := []attribute.KeyValue{
		attribute.String("usecase.type", "write"),
	}
	// 動態判斷是否實作了 Traceable
	if t, ok := any(input).(Traceable); ok {
		attrs = append(attrs, t.ToAttributes()...)
	}

	// 1. 啟動 Span
	ctx, span := d.tracer.Start(ctx, "UseCase."+d.next.Name(),
		trace.WithAttributes(attrs...),
	)
	defer span.End()

	// 2. 執行業務邏輯
	r, err := d.next.Execute(ctx, input)

	// 3. 異常紀錄
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "success")
	}

	return r, err
}

func WithWriteOTel[T any, R any](next WriteUseCase[T, R]) WriteUseCase[T, R] {
	return &writeOtelDecorator[T, R]{
		next:   next,
		tracer: otel.Tracer("usecase-write-layer"),
	}
}

type readOtelDecorator[I any, O any] struct {
	next   ReadUseCase[I, O]
	tracer trace.Tracer
}

func (d *readOtelDecorator[I, O]) Name() string {
	return d.next.Name()
}

func (d *readOtelDecorator[I, O]) Execute(ctx context.Context, input I) (O, UseCaseError) {
	attrs := []attribute.KeyValue{
		attribute.String("usecase.type", "read"),
	}
	// 動態判斷是否實作了 Traceable
	if t, ok := any(input).(Traceable); ok {
		attrs = append(attrs, t.ToAttributes()...)
	}

	// 1. 啟動 Span
	ctx, span := d.tracer.Start(ctx, "UseCase."+d.next.Name(),
		trace.WithAttributes(attrs...),
	)
	defer span.End()

	// 2. 執行業務邏輯
	output, err := d.next.Execute(ctx, input)

	// 3. 異常紀錄
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "success")
	}

	return output, err
}

func WithReadOTel[I any, O any](next ReadUseCase[I, O]) ReadUseCase[I, O] {
	return &readOtelDecorator[I, O]{
		next:   next,
		tracer: otel.Tracer("usecase-read-layer"),
	}
}
