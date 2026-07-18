package shogi_test

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"shogi"
	"strings"
	"sync"
	"testing"
	"time"
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

// StageF-4: USIプロトコル拡張(quit/go/stop/setoption)のテスト。
// io.Pipeで実際の入出力を模擬し、shogi.USI.Start()を並行実行して確認する。

// ctx対応のモックエンジン。GetBestContextはinfoを1回通知した後、
// ctxがキャンセルされるか短いタイムアウトが経過するまで待ってから
// 手を返す(stopによる打ち切りを確認するため)。
type mockCtxEngine struct {
	waitFor time.Duration
}

func (e *mockCtxEngine) GetName() string    { return "mockctx" }
func (e *mockCtxEngine) GetVersion() string { return "1.0.0" }
func (e *mockCtxEngine) GetAuthor() string  { return "test" }

func (e *mockCtxEngine) GetBest(b *shogi.Board) (*shogi.Action, error) {
	return shogi.NewAction("7g7f"), nil
}

func (e *mockCtxEngine) GetBestContext(ctx context.Context, b *shogi.Board, info func(shogi.Info)) (*shogi.Action, error) {
	if info != nil {
		info(shogi.Info{Depth: 1, ScoreCP: 0, Nodes: 1})
	}
	wait := e.waitFor
	if wait == 0 {
		wait = 2 * time.Second
	}
	select {
	case <-ctx.Done():
	case <-time.After(wait):
	}
	return shogi.NewAction("7g7f"), nil
}

// setoptionを記録するだけのオプション対応エンジン
type mockOptionEngine struct {
	mockCtxEngine
	mu   sync.Mutex
	opts map[string]string
}

func (e *mockOptionEngine) SetOption(name, value string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.opts == nil {
		e.opts = make(map[string]string)
	}
	e.opts[name] = value
}

// StageG: go mate (詰将棋探索)対応のモックエンジン。
type mockMateEngine struct {
	moves   []*shogi.Action
	mateErr error
}

func (e *mockMateEngine) GetName() string    { return "mockmate" }
func (e *mockMateEngine) GetVersion() string { return "1.0.0" }
func (e *mockMateEngine) GetAuthor() string  { return "test" }

func (e *mockMateEngine) GetBest(b *shogi.Board) (*shogi.Action, error) {
	return shogi.NewAction("7g7f"), nil
}

func (e *mockMateEngine) GetMate(ctx context.Context, b *shogi.Board) ([]*shogi.Action, error) {
	return e.moves, e.mateErr
}

// USIをio.Pipe経由で起動するテストハーネス。
func startTestUSI(t *testing.T, en shogi.Engine) (send func(string), readLine func(timeout time.Duration) (string, error), done chan error) {
	t.Helper()

	inR, inW := io.Pipe()
	outR, outW := io.Pipe()

	usi := shogi.NewUSI(outW, inR, en)

	done = make(chan error, 1)
	go func() {
		done <- usi.Start()
	}()

	send = func(line string) {
		io.WriteString(inW, line+"\n")
	}

	lines := make(chan string, 16)
	go func() {
		r := bufio.NewReader(outR)
		for {
			l, err := r.ReadString('\n')
			if l != "" {
				lines <- strings.TrimRight(l, "\n")
			}
			if err != nil {
				return
			}
		}
	}()

	readLine = func(timeout time.Duration) (string, error) {
		select {
		case l := <-lines:
			return l, nil
		case <-time.After(timeout):
			return "", fmt.Errorf("timeout waiting for output line")
		}
	}

	t.Cleanup(func() {
		inW.Close()
		outW.Close()
	})

	return send, readLine, done
}

// (a) quit で Start() が nil を返すこと。
func TestUSIQuitReturnsNil(t *testing.T) {
	send, _, done := startTestUSI(t, &engine{})

	send("quit")

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("expected nil error on quit, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for Start() to return after quit")
	}
}

// (b) position -> go で bestmove が出力されること。
func TestUSIPositionGoBestmove(t *testing.T) {
	en := &mockCtxEngine{waitFor: 10 * time.Millisecond}
	send, readLine, done := startTestUSI(t, en)
	defer func() {
		send("quit")
		<-done
	}()

	send("position startpos")
	send("go")

	found := false
	for i := 0; i < 5; i++ {
		line, err := readLine(2 * time.Second)
		if err != nil {
			t.Fatalf("readLine error: %v", err)
		}
		if strings.HasPrefix(line, "bestmove ") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected a bestmove line")
	}
}

// (c) go -> (探索中) -> stop で bestmove が出力されること。
// Movetime長め(実質無制限に近い待ち)のエンジンに対し、stopをすぐ送って
// 打ち切られることを確認する。
func TestUSIGoStopBestmove(t *testing.T) {
	en := &mockCtxEngine{waitFor: 5 * time.Second}
	send, readLine, done := startTestUSI(t, en)
	defer func() {
		send("quit")
		<-done
	}()

	send("position startpos")
	send("go")

	//探索が始まる猶予を少し置いてからstopを送る
	time.Sleep(50 * time.Millisecond)
	send("stop")

	found := false
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		line, err := readLine(2 * time.Second)
		if err != nil {
			t.Fatalf("readLine error: %v", err)
		}
		if strings.HasPrefix(line, "bestmove ") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected a bestmove line after stop")
	}
}

// go mate: MateEngine実装済みエンジンで、1手詰め局面のcheckmate応答を確認する。
func TestUSIGoMateCheckmateResponse(t *testing.T) {
	en := &mockMateEngine{moves: []*shogi.Action{shogi.NewAction("G*5b")}}
	send, readLine, done := startTestUSI(t, en)
	defer func() {
		send("quit")
		<-done
	}()

	send("position sfen 3pkp3/3s1s3/4S4/9/9/9/9/9/9 b G 1")
	send("go mate 1000")

	line, err := readLine(2 * time.Second)
	if err != nil {
		t.Fatalf("readLine error: %v", err)
	}
	if line != "checkmate G*5b" {
		t.Errorf("expected \"checkmate G*5b\", got %q", line)
	}
}

// go mate: MateEngine未実装エンジンでは checkmate notimplemented が返ること。
func TestUSIGoMateNotImplemented(t *testing.T) {
	en := &mockCtxEngine{}
	send, readLine, done := startTestUSI(t, en)
	defer func() {
		send("quit")
		<-done
	}()

	send("position startpos")
	send("go mate 1000")

	line, err := readLine(2 * time.Second)
	if err != nil {
		t.Fatalf("readLine error: %v", err)
	}
	if line != "checkmate notimplemented" {
		t.Errorf("expected \"checkmate notimplemented\", got %q", line)
	}
}

// (d) setoption がエラーにならず、後続コマンドも正しく処理されること。
func TestUSISetOptionThenIsReady(t *testing.T) {
	en := &mockOptionEngine{}
	send, readLine, done := startTestUSI(t, en)
	defer func() {
		send("quit")
		<-done
	}()

	send("setoption name USI_Hash value 1024")
	send("isready")

	line, err := readLine(2 * time.Second)
	if err != nil {
		t.Fatalf("readLine error: %v", err)
	}
	if line != "readyok" {
		t.Errorf("expected readyok after setoption, got %q", line)
	}
}
