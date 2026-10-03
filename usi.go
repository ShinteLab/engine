package shogi

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/xerrors"
)

type USI struct {
	out  io.Writer
	recv *bufio.Scanner

	sender chan string
	engine Engine

	nowBoard *Board

	options map[string]string

	mu        sync.Mutex
	cancel    context.CancelFunc
	searching bool
}

func NewUSI(out io.Writer, in io.Reader, en Engine) *USI {
	var inst USI
	inst.out = out

	inst.sender = make(chan string)

	inst.recv = bufio.NewScanner(in)
	inst.engine = en
	inst.options = make(map[string]string)
	return &inst
}

var Quit = fmt.Errorf("quit")

func (usi *USI) Start() error {

	logger().Info("usi start")

	quit := make(chan error)

	defer func() {
		err := recover()
		if err != nil {
			logger().Error(fmt.Sprintf("Panic:\n%+v", err))
		}
	}()

	go func() {
		for {
			err := usi.Wait()
			if err != nil {
				quit <- err
			}
		}
	}()

	//コマンド送信側に送る
	for {
		select {
		case cmd := <-usi.sender:
			err := usi.sendCommand(cmd)
			if err != nil {
				quit <- err
			}
		case err := <-quit:
			if errors.Is(err, Quit) {
				return nil
			} else if err != nil {
				logger().Error(fmt.Sprintf("%+v", err))
				return xerrors.Errorf("Wait() error: %w", err)
			}
			return nil
		}
	}
}

func (usi *USI) Wait() error {

	line, err := usi.wait()
	if err != nil {
		return xerrors.Errorf("usi.wait() error: %w", err)
	}

	if line == "" && err == nil {
		return fmt.Errorf("EOF error")
	}

	logger().Debug(fmt.Sprintf("USER> [%s]", line))

	//strings.Fields は連続する空白・前後の余分な空白を無視してトークン化する。
	//strings.Split(line," ")では末尾の空白等が空文字列トークンを生み、
	//下流(NewBoard等)でのインデックス範囲外パニックにつながるため使わない。
	cmds := strings.Fields(line)
	if len(cmds) == 0 {
		return nil
	}
	var args []string
	cmd := cmds[0]
	if len(cmds) >= 2 {
		args = cmds[1:]
	}

	return usi.run(cmd, args...)
}

func (usi *USI) sendCommand(cmd string) error {

	logger().Debug(fmt.Sprintf("USER< [%s]", cmd))

	var buf bytes.Buffer
	_, err := buf.WriteString(cmd + "\n")
	if err != nil {
		return err
	}

	_, err = buf.WriteTo(usi.out)
	return err
}

func (usi *USI) run(cmd string, args ...string) error {

	switch cmd {
	case "quit":
		return Quit
	case "usi":
		return usi.sendInformation()
	case "setoption":
		return usi.setOption(args)
	case "isready":
		return usi.sendReady()
	case "usinewgame":
		return nil
	case "position":
		return usi.setBoard(args)
	case "gameover":

		//msg="receive command[gameover lose]"
		//対局結果の通知。現状は特に何もしない(無視する)。
		return nil

	case "go":

		//msg="receive command[go btime 559935 wtime 600000 byoyomi 30000]"
		//
		// go ponder -> 予想局面
		// go ponderhit -> 予想があったった
		// stop -> 外れた場合
		// go mate <ms>/go mate infinite -> 詰め将棋探索
		//     checkmate <手順> / checkmate nomate / checkmate timeout
		if len(args) > 0 && args[0] == "mate" {
			return usi.mateCommand(args[1:])
		}
		return usi.goCommand(args)
	case "stop":
		usi.stopSearch()
		return nil
	}
	return fmt.Errorf("invalid command.[%s]", cmd)
}

func (usi *USI) wait() (string, error) {
	ok := usi.recv.Scan()
	if !ok {
		return "", usi.recv.Err()
	}
	return usi.recv.Text(), nil
}

func (usi *USI) sendInformation() error {

	usi.sender <- "id name " + usi.engine.GetName()
	usi.sender <- "id author " + usi.engine.GetAuthor()
	usi.sender <- "usiok"

	return nil
}

func (usi *USI) sendReady() error {
	usi.sender <- "readyok"
	return nil
}

// setoption name <id> value <x> をパースして保持する。
// エンジンが OptionEngine を実装していれば転送する。
// USI_Ponder 等の option 宣言(usi応答時)は最小実装のため行わない。
func (usi *USI) setOption(args []string) error {

	if len(args) == 0 || args[0] != "name" {
		//フォーマット不正は無視する(最小実装)
		return nil
	}

	valueIdx := -1
	for i, a := range args {
		if a == "value" {
			valueIdx = i
			break
		}
	}

	var name, value string
	if valueIdx == -1 {
		name = strings.Join(args[1:], " ")
	} else {
		name = strings.Join(args[1:valueIdx], " ")
		value = strings.Join(args[valueIdx+1:], " ")
	}

	if name == "" {
		return nil
	}

	usi.options[name] = value

	if oe, ok := usi.engine.(OptionEngine); ok {
		oe.SetOption(name, value)
	}

	return nil
}

func (usi *USI) setBoard(args []string) error {

	leng := len(args)
	if leng == 0 {
		return fmt.Errorf("position arguments error[zero]")
	}
	var err error
	usi.nowBoard, err = NewBoard(strings.Join(args, " "))
	if err != nil {
		return fmt.Errorf("SFEN parse error: %w", err)
	}
	return nil
}

const (
	Resign = "resign"
)

// 現在探索中であればキャンセルする。探索中でなければ何もしない。
func (usi *USI) stopSearch() {
	usi.mu.Lock()
	c := usi.cancel
	searching := usi.searching
	usi.mu.Unlock()

	if searching && c != nil {
		c()
	}
}

// go コマンドを処理する。既に探索中なら前の探索をキャンセルしてから
// 新しい探索を開始する。
//
// engine が ContextEngine を実装していれば非同期(goroutine)に
// GetBestContext を実行し、info コールバックを usi.sender へ転送する。
// bestmove は探索完了時に送出する。
// 従来の Engine のみ実装している場合は同期的に GetBest を呼ぶ
// (stop によるキャンセルは効かない)。
func (usi *USI) goCommand(args []string) error {

	//前の探索が残っていればキャンセルしてから開始する
	usi.stopSearch()

	if usi.nowBoard == nil {
		return fmt.Errorf("go error: position is not set")
	}

	ctx, cancel := context.WithCancel(context.Background())

	movetime := parseGoTime(args, usi.nowBoard.Turn())
	if movetime > 0 {
		ctx, cancel = context.WithTimeout(ctx, movetime)
	}

	usi.mu.Lock()
	usi.cancel = cancel
	usi.searching = true
	usi.mu.Unlock()

	board := usi.nowBoard

	go func() {
		defer func() {
			usi.mu.Lock()
			usi.searching = false
			usi.cancel = nil
			usi.mu.Unlock()
			cancel()
		}()

		var action *Action
		var err error

		if ce, ok := usi.engine.(ContextEngine); ok {
			action, err = ce.GetBestContext(ctx, board, func(info Info) {
				usi.sender <- formatInfo(info)
			})
		} else {
			action, err = usi.engine.GetBest(board)
		}

		if err != nil {
			logger().Error(fmt.Sprintf("GetBest error: %v", err))
			return
		}
		if action == nil {
			logger().Error("GetBest returned nil action")
			return
		}

		logger().Debug(fmt.Sprintf("Ans:%v", action))
		usi.sender <- "bestmove " + action.String()
	}()

	return nil
}

// go mate <ms> / go mate infinite を処理する。
// engine が MateEngine を実装していなければ "checkmate notimplemented" を
// 即座に送出する。実装していれば goroutine で GetMate を実行し、
// 詰みあり: "checkmate <手順>"、詰みなし: "checkmate nomate"、
// タイムアウト/キャンセル: "checkmate timeout" を送出する。
// 探索中の stop は通常探索と同様にキャンセルとして扱われる
// (usi.cancel/usi.searching を共有しているため)。
func (usi *USI) mateCommand(args []string) error {

	usi.stopSearch()

	me, ok := usi.engine.(MateEngine)
	if !ok {
		usi.sender <- "checkmate notimplemented"
		return nil
	}

	if usi.nowBoard == nil {
		return fmt.Errorf("go mate error: position is not set")
	}

	ctx, cancel := context.WithCancel(context.Background())

	if len(args) > 0 && args[0] != "infinite" {
		if ms, err := strconv.ParseInt(args[0], 10, 64); err == nil && ms > 0 {
			ctx, cancel = context.WithTimeout(ctx, time.Duration(ms)*time.Millisecond)
		}
	}

	usi.mu.Lock()
	usi.cancel = cancel
	usi.searching = true
	usi.mu.Unlock()

	board := usi.nowBoard

	go func() {
		defer func() {
			usi.mu.Lock()
			usi.searching = false
			usi.cancel = nil
			usi.mu.Unlock()
			cancel()
		}()

		moves, err := me.GetMate(ctx, board)

		if err != nil {
			logger().Error(fmt.Sprintf("GetMate error: %v", err))
			usi.sender <- "checkmate timeout"
			return
		}

		if len(moves) == 0 {
			usi.sender <- "checkmate nomate"
			return
		}

		strs := make([]string, 0, len(moves))
		for _, m := range moves {
			if m == nil {
				continue
			}
			strs = append(strs, m.String())
		}
		usi.sender <- "checkmate " + strings.Join(strs, " ")
	}()

	return nil
}

// info 行を USI プロトコル形式にフォーマットする。
func formatInfo(info Info) string {

	pvStrs := make([]string, 0, len(info.PV))
	for _, a := range info.PV {
		if a == nil {
			continue
		}
		pvStrs = append(pvStrs, a.String())
	}

	buf := fmt.Sprintf("info depth %d score cp %d nodes %d", info.Depth, info.ScoreCP, info.Nodes)
	if len(pvStrs) > 0 {
		buf += " pv " + strings.Join(pvStrs, " ")
	}
	return buf
}

// go btime/wtime/byoyomi/binc/winc をパースし、簡易な時間配分で
// Movetime を計算する: 持ち時間/40 + byoyomiの80% + increment。
// パースできる時間指定が無ければ 0(無制限)を返す。
func parseGoTime(args []string, turn TurnType) time.Duration {

	var btime, wtime, byoyomi, binc, winc int64
	for i := 0; i < len(args)-1; i++ {
		v, err := strconv.ParseInt(args[i+1], 10, 64)
		if err != nil {
			continue
		}
		switch args[i] {
		case "btime":
			btime = v
		case "wtime":
			wtime = v
		case "byoyomi":
			byoyomi = v
		case "binc":
			binc = v
		case "winc":
			winc = v
		}
	}

	var own, inc int64
	if turn == TurnBlack {
		own = btime
		inc = binc
	} else {
		own = wtime
		inc = winc
	}

	if own <= 0 && byoyomi <= 0 {
		return 0
	}

	ms := own/40 + (byoyomi*80)/100 + inc
	if ms <= 0 {
		return 0
	}

	return time.Duration(ms) * time.Millisecond
}
