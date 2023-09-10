package shogi

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"golang.org/x/xerrors"
)

type USI struct {
	out  io.Writer
	recv *bufio.Scanner

	engine Engine

	nowBoard *Board
}

func NewUSI(out io.Writer, in io.Reader, en Engine) *USI {
	var inst USI
	inst.out = out

	inst.recv = bufio.NewScanner(in)
	inst.engine = en
	return &inst
}

func (usi *USI) Start() error {
	slog.Info("usi start")
	for {
		err := usi.Wait()
		if err != nil {
			slog.Error(fmt.Sprintf("%+v", err))
			return xerrors.Errorf("Wait() error: %w", err)
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

	usi.sendCommand("id name " + usi.engine.GetName())
	usi.sendCommand("id author " + usi.engine.GetAuthor())
	usi.sendCommand("usiok")

	return nil
}

func (usi *USI) sendReady() error {
	return usi.sendCommand("readyok")
}

const (
	FirstPosSFEN = "lnsgkgsnl/1r5b1/ppppppppp/9/9/9/PPPPPPPPP/1B5R1/LNSGKGSNL"
	StartPos     = "startpos"
	Moves        = "moves"
	None         = "-" //手駒
)

func (usi *USI) setBoard(args []string) error {

	leng := len(args)
	if leng == 0 {
		return fmt.Errorf("position arguments error[zero]")
	}

	t := args[0]
	idx := 1
	sfen := ""

	if t == "sfen" {
		if leng <= 1 {
			return fmt.Errorf("position arguments error[sfen]")
		}
		sfen = args[1]
		idx = 2
	} else if t == StartPos {
		sfen = FirstPosSFEN
	}

	//後手で相手が指した時
	//msg="receive command[position startpos moves 8g8f]"
	//上手で指す時
	//msg="receive command[position sfen lnsgkgsnl/7b1/ppppppppp/9/9/9/PPPPPPPPP/1B5R1/LNSGKGSNL w - 1]"
	//下手で相手が指した時
	//"USER> [position sfen lnsgkgsnl/7b1/ppppppppp/9/9/9/PPPPPPPPP/1B5R1/LNSGKGSNL w - 1 moves 8c8d]"
	//"USER> [position startpos moves 6g6f 3c3d 5g5f]"

	nowBoard, err := NewBoard(sfen)
	if err != nil {
		return xerrors.Errorf("NewBoard() error: %w", err)
	}

	if leng > idx {
		next := args[idx]
		if next != Moves {
			if next == "w" {
				nowBoard.setTurn(TurnWhite)
			} else if next != "b" {
				return fmt.Errorf("position arguments error[turn]")
			}

			pieces := args[idx+1]
			nowBoard.setHave(pieces)
			idx = idx + 3
		}

		next = args[idx]
		if next == Moves {
			idx++
			ms := args[idx:]
			for _, mov := range ms {
				nowBoard.Action(NewAction(mov))
			}
		}
	}

	usi.nowBoard = nowBoard
	return nil
}

const (
	Resign = "resign"
)

func (usi *USI) sendBest() error {

	action, err := usi.engine.GetBest(usi.nowBoard)
	if err != nil {
		return err
	}
	return usi.sendCommand("bestmove " + action.String())
}
