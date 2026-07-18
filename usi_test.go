package shogi_test

import (
	"bytes"
	"io"
	"shogi"
	"testing"
)

var mock *MockUI

func init() {
	var in bytes.Buffer
	var out bytes.Buffer
	mock = NewMock(&out, &in)
}

func TestNewUSI(t *testing.T) {
	var e engine
	var r bytes.Buffer
	usi := shogi.NewUSI(io.Discard, io.NopCloser(&r), &e)
	if usi == nil {
		t.Errorf("NewUSI not nil")
	}
	//TODO 非同期でStartして処理が終わらないことを確認
}

type MockUI struct {
	in  io.Reader
	out io.Writer
}

func NewMock(out io.Writer, in io.Reader) *MockUI {
	var mock MockUI
	mock.out = out
	mock.in = in
	return &mock
}

type engine struct {
}

func (e *engine) GetName() string {
	return "name"
}

func (e *engine) GetVersion() string {
	return "1.0.0"
}

func (e *engine) GetAuthor() string {
	return "author"
}

func (e *engine) GetBest(b *shogi.Board) (*shogi.Action, error) {
	return nil, nil
}
