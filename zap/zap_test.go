package zap_test

import (
	"errors"
	"testing"

	"github.com/wakaranakattari/zerr"
	zerrzap "github.com/wakaranakattari/zerr/zap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func logField(err error) map[string]any {
	core, logs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	logger.Info("x", zerrzap.Field(err))
	return logs.All()[0].ContextMap()
}

func TestFieldNested(t *testing.T) {
	err := zerr.WithCode(
		zerr.Wrap(zerr.New("no rows"), "db", "table", "users"),
		zerr.Code("not_found"), "load", "id", 42,
	)
	m := logField(err)

	e := m["err"].(map[string]any)
	if got := e["op"]; got != "load" {
		t.Errorf("err.op = %v, want load", got)
	}
	if got := e["code"]; got != "not_found" {
		t.Errorf("err.code = %v, want not_found", got)
	}
	if got := e["id"]; got != 42 {
		t.Errorf("err.id = %v, want 42", got)
	}
	cause := e["cause"].(map[string]any)
	if got := cause["op"]; got != "db" {
		t.Errorf("err.cause.op = %v, want db", got)
	}
	if got := cause["table"]; got != "users" {
		t.Errorf("err.cause.table = %v, want users", got)
	}
	if _, ok := e["msg"]; ok {
		t.Error("a Wrap node has no message; msg must be absent")
	}
}

func TestFieldPrivateAttrHidden(t *testing.T) {
	err := zerr.Wrap(zerr.New("boom"), "load", zerr.Sec("password", "hunter2"))
	m := logField(err)
	e := m["err"].(map[string]any)
	if _, ok := e["password"]; ok {
		t.Error("private attributes must never reach zap")
	}
}

func TestFieldPublicExcluded(t *testing.T) {
	err := zerr.Public(zerr.New("boom"), "user message")
	m := logField(err)
	if _, ok := m["err"].(map[string]any)["public"]; ok {
		t.Error("the public message is for clients, not logs")
	}
}

func TestFieldForeignAndNil(t *testing.T) {
	m := logField(errors.New("plain"))
	e := m["err"].(map[string]any)
	if got := e["msg"]; got != "plain" {
		t.Errorf("err.msg = %v, want plain", got)
	}
	if m := logField(nil); len(m["err"].(map[string]any)) != 0 {
		t.Errorf("nil error must encode to an empty object, got %v", m)
	}
}

func TestFieldDepthBound(t *testing.T) {
	err := zerr.New("boom")
	for i := 0; i < 7; i++ {
		err = zerr.Wrap(err, "op")
	}
	e := logField(err)["err"].(map[string]any)
	for depth := 0; depth < 5; depth++ {
		cause, ok := e["cause"]
		if !ok {
			t.Fatalf("expected cause at depth %d", depth)
		}
		if depth == 3 {
			if _, isString := cause.(string); !isString {
				t.Fatalf("depth cap must collapse after three levels, got %T at depth %d", cause, depth)
			}
			return
		}
		e = cause.(map[string]any)
	}
}
