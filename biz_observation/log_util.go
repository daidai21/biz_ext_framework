package biz_observation

import (
	"context"
)

type LogLevel string

const (
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
)

type LogField struct {
	Key   string
	Value any
}

type Logger interface {
	Log(ctx context.Context, level LogLevel, msg string, fields ...LogField)
}

type logFieldsContextKey struct{}

func WithLogFields(ctx context.Context, fields ...LogField) context.Context {
	if len(fields) == 0 {
		return ctx
	}
	existing := LogFieldsFromContext(ctx)
	merged := MergeLogFields(existing, fields)
	return context.WithValue(ctx, logFieldsContextKey{}, merged)
}

func LogFieldsFromContext(ctx context.Context) []LogField {
	if ctx == nil {
		return nil
	}
	fields, _ := ctx.Value(logFieldsContextKey{}).([]LogField)
	return append([]LogField(nil), fields...)
}

func MergeLogFields(groups ...[]LogField) []LogField {
	type fieldIndex struct {
		index int
		field LogField
	}

	ordered := make([]fieldIndex, 0)
	indexByKey := map[string]int{}
	for _, group := range groups {
		for _, field := range group {
			if field.Key == "" {
				continue
			}
			if idx, ok := indexByKey[field.Key]; ok {
				ordered[idx].field = field
				continue
			}
			indexByKey[field.Key] = len(ordered)
			ordered = append(ordered, fieldIndex{
				index: len(ordered),
				field: field,
			})
		}
	}

	fields := make([]LogField, 0, len(ordered))
	for _, item := range ordered {
		fields = append(fields, item.field)
	}
	return fields
}

func Log(ctx context.Context, logger Logger, level LogLevel, msg string, fields ...LogField) {
	if logger == nil {
		return
	}
	logger.Log(ctx, level, msg, MergeLogFields(LogFieldsFromContext(ctx), fields)...)
}
