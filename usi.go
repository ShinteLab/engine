package shogi

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"golang.org/x/xerrors"
)

type USI struct {
	out  io.Writer
	recv *bufio.Scanner

	sender chan string
	engine Engine

	nowBoard *Board
}

func NewUSI(out io.Writer, in io.Reader, en Engine) *USI {
	var inst USI
	inst.out = out

	inst.sender = make(chan string)

	inst.recv = bufio.NewScanner(in)
	inst.engine = en
	return &inst
}

var Quit = fmt.Errorf("quit")

func (usi *USI) Start() error {

	slog.Info("usi start")

	quit := make(chan error)

	defer func() {
		err := recover()
		if err != nil {
			slog.Error("Panic:\n%+v", err)
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
				slog.Error(fmt.Sprintf("%+v", err))
				return xerrors.Errorf("Wait() error: %w", err)
			}
			return nil
		}
	}

	return nil
}

func (usi *USI) Wait() error {

	line, err := usi.wait()
	if err != nil {
		return xerrors.Errorf("usi.wait() error: %w", err)
	}

	if line == "" && err == nil {
		return fmt.Errorf("EOF error")
	}

	slog.Info(fmt.Sprintf("USER> [%s]", line))

	cmds := strings.Split(line, " ")
	var args []string
	cmd := cmds[0]
	if len(cmds) >= 2 {
		args = cmds[1:]
	}

	return usi.run(cmd, args...)
}

func (usi *USI) sendCommand(cmd string) error {

	slog.Info(fmt.Sprintf("USER< [%s]", cmd))

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
		return fmt.Errorf("quit")
	case "usi":
		return usi.sendInformation()
	case "setoption":
		//msg="receive command[setoption name USI_Hash value 1024]"
		//msg="receive command[setoption name USI_Ponder value false]"
		return nil
	case "isready":
		return usi.sendReady()
	case "usinewgame":
		return nil
	case "position":
		return usi.setBoard(args)
	case "gameover":

		//msg="receive command[gameover lose]"
		return nil

	case "go":

		//msg="receive command[go btime 559935 wtime 600000 byoyomi 30000]"
		//
		// go ponder -> 予想局面
		// go ponderhit -> 予想があったった
		// stop -> 外れた場合
		// go mate -> 詰め将棋
		//     checkmate
		return usi.sendBest()
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

const ()

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

func (usi *USI) sendBest() error {

	slog.Info("Call GetBest()")

	action, err := usi.engine.GetBest(usi.nowBoard)
	if err != nil {
		return err
	}
	slog.Info(fmt.Sprintf("Ans:%v", action))

	usi.sender <- "bestmove " + action.String()
	return nil
}
