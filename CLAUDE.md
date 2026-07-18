# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

A Go library (`module shogi`) that implements the foundation for building shogi (Japanese chess) engines: board representation, legal-move generation, and the USI (Universal Shogi Interface) protocol loop. An engine author implements the `Engine` interface (engine.go) and calls `shogi.Start(engine)`; the library handles USI communication with a GUI over stdin/stdout. Comments and docs are in Japanese.

## Commands

```powershell
go build ./...          # build the library (underscore dirs _cmd/_samples are excluded by the go tool)
go test .               # run all tests
go test . -run TestBoardParse   # run a single test
go test . -v -run TestCanBit    # verbose single test
```

The `_cmd` directory holds one `main` per file (they conflict, so build a single file at a time), producing runnable USI engines that wrap the engines in `_samples`:

```powershell
go build -o _dist/think.exe _cmd/think.go
```

Because `_cmd` and `_samples` start with `_`, the go tool ignores them in `./...` patterns; they are compiled only when named explicitly. `_samples` is imported as `shogi/_samples`.

Engines log to `shogi_<pid>.log` in the working directory (see logger.go); `*.log` files are gitignored.

## Architecture

Everything in the repo root is one package, `shogi`.

**USI layer** (usi.go, shogi.go): `USI.Start()` runs two goroutines — one scanning stdin for GUI commands (`usi`, `isready`, `position`, `go`, `quit`, ...) and a select loop writing responses from the `sender` channel to stdout. `position` builds a `Board` via `NewBoard(sfen)`; `go` calls `engine.GetBest(board)` and replies `bestmove`. All protocol traffic is logged via `slog`.

**Board model** — three layers, each side ("camp") stored separately:

- `Board` (board.go): whole game state — `turn`, move number, and `camps [2]*CampBoard` (index 0 = black `b`, 1 = white `w`, via `TurnType.Index()`). Parses SFEN/`startpos ... moves ...` strings, applies moves (`Action(a)` switches turn), and exposes `Candidate()` for legal moves of the side to move. `Copy()` makes a deep copy for search (used heavily by the thinking sample's parallel search).
- `CampBoard` (board_camp.go): one player's pieces — 14 per-piece-type `BitBoard`s plus an aggregate `board` bitboard maintained as their `parent`, the hand (`has Pieces`), and a mutual `enemy` pointer wired by `setEnemy`. Move generation lives here: `candidate()` combines board moves (`canBit`) with drops onto `emptyPos()` (including the no-two-pawns-per-file check via `hasX`).
- `BitBoard` (board_bit.go): 81 squares in `[3]uint32` (three 27-bit ranks of 3 rows × 9 columns); `set`/`clear` propagate to the `parent` aggregate board.

**Move generation** (can.go): `canLogic(turn, pieceType, pos)` returns geometric move rays (`[]Vector`) per piece type ignoring other pieces; `canFilter` then prunes by own/enemy occupancy, truncates sliding rays at blockers, and emits promotion variants (`+` suffix) where `GrowthArea` applies. Promoted piece types are the `Growth*` constants (piece.go), which share movement with gold/king patterns as appropriate.

**Moves** (action.go, pos.go): `Action` is parsed from/rendered to USI move strings (`7g7f`, `7g7f+` promotion, `P*5e` drop, `resign`). `Pos` encodes a square; USI coordinates are file 9→1 left-to-right, rank a–i as y 1–9.

**Coordinate convention**: throughout the code `x, y` are 1–9; `isArea` bounds-checks. Black (`b`, lowercase-`b` turn, uppercase SFEN letters) moves in the −y direction, white in +y.

## Tests

Tests live alongside the code (`*_test.go` in package `shogi`). Private functions are exposed to tests through export_test.go (`ExportCanLogic`, `ExportBitBoardGet`, ...) — add new exports there when testing unexported behavior rather than changing visibility.
