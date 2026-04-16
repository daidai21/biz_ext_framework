package biz_observation

import (
	"context"
	"testing"
)

type testLogger struct {
	level  LogLevel
	msg    string
	fields []LogField
}

func (l *testLogger) Log(ctx context.Context, level LogLevel, msg string, fields ...LogField) {
	l.level = level
	l.msg = msg
	l.fields = append([]LogField(nil), fields...)
}

func TestWithLogFields(t *testing.T) {
	ctx := context.Background()
	ctx = WithLogFields(ctx, LogField{Key: "trace_id", Value: "t1"})
	ctx = WithLogFields(ctx, LogField{Key: "user_id", Value: 1}, LogField{Key: "trace_id", Value: "t2"})

	fields := LogFieldsFromContext(ctx)
	if len(fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(fields))
	}
	if fields[0].Key != "trace_id" || fields[0].Value != "t2" {
		t.Fatalf("unexpected trace field: %+v", fields[0])
	}
}

func TestLog(t *testing.T) {
	logger := &testLogger{}
	ctx := WithLogFields(context.Background(), LogField{Key: "trace_id", Value: "t1"})

	Log(ctx, logger, LogLevelInfo, "created", LogField{Key: "order_id", Value: "o1"})

	if logger.level != LogLevelInfo || logger.msg != "created" {
		t.Fatalf("unexpected log meta: %+v", logger)
	}
	if len(logger.fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(logger.fields))
	}
}
