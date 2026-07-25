# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

A Go shogi engine library: bitboard-based board representation, fully legal move generation (perft-verified), the USI protocol loop, and a `search` subpackage (alpha-beta + quiescence + transposition table + Lazy SMP). An engine author implements the `Engine` interface (engine.go) and calls `shogi.Start(engine)`; optional interfaces `ContextEngine` (info output / stop), `MateEngine` (`go mate`), and `OptionEngine` (`setoption`) unlock USI extensions. Comments and docs are in Japanese.

This directory is `package shogi` inside the repo-wide module `shinte` (there is no `module shogi` of its own any more — the old standalone module was folded in). Import paths are `shinte/engine` and `shinte/engine/search`. See the repository root `CLAUDE.md` for cross-project rules.

### Dependency on `core`

SFEN/USI notation is **not implemented here** — it lives in `shinte/core/sfen` and `shinte/core/usi`, shared with suteme / prokishi / the frontend. Used from board.go, piece.go (`sfen`), action.go, pos.go, search/tt.go (`usi`).

- `core/sfen`'s piece codes are numerically identical to this package's `PieceType` base values, so engine passes its own piece types through **without conversion**. Do not renumber `PieceType` without changing `core/sfen` in the same edit.
- Keep conversions at the I/O boundary (`position` parsing, `bestmove` output). Never add SFEN/USI string work inside move generation or search.
- If you need a new notation helper, add it to `core`, not here.

## Commands

```powershell
go build ./...              # library + search package (underscore dirs are excluded by the go tool)
go test ./...               # all tests (root package + search)
go test . -run TestPerftDepth3      # single test
go test . -bench "Candidate|Perft" -benchmem -run XXX   # benchmarks (root)
go test ./search -bench Depth4 -benchmem -run XXX       # search benchmarks
```

Sample engines in `_cmd` hold one `main` per file (they conflict; build one file at a time):

```powershell
go build -o _dist/think.exe _cmd/think.go
```

E2E smoke test (PowerShell 5.1 pipes inject a UTF-8 BOM that corrupts the first USI command — use cmd.exe piping, with `ping` for pacing so the search finishes before `quit`):

```powershell
cmd /c "(echo usi& echo isready& echo position startpos& echo go byoyomi 1000& ping -n 4 127.0.0.1 >nul& echo quit) | .\_dist\think.exe"
```

Known environment limitation: `go test -race` fails to load on this machine (`0xc0000139`, mingw-w64/race-runtime mismatch) — it is not a code problem. Concurrency safety is covered instead by the TT xor-verify stress test in `search/tt` tests.

Engines log to `shogi_<pid>.log` in the working directory; `*.log` and `_dist` are gitignored.

## Architecture

`engine/` itself is one package, `shogi`; `search/` is the only subpackage.

**Board model**: `Board` (board.go) holds `turn`, move number, `camps [2]*CampBoard` (index 0 = black `b` via `TurnType.Index()`), the Zobrist `hash`, and position `history` for repetition detection. `CampBoard` (board_camp.go) holds one side's 14 per-piece-type `BitBoard`s plus an aggregate `board`, the hand (`has Pieces`), and a mutual `enemy` pointer. `BitBoard` (board_bit.go) packs 81 squares into `[2]uint64`, linear `sq = (y-1)*9 + (x-1)`; iterate with `forEach` (TrailingZeros64).

**Move generation** (attack.go, legal.go, board_camp.go): precomputed attack tables built in `init()` — `attackTable` masks for non-sliders, `rayTable` nearest-first ray lists for sliders — with single entry point `attacks(turn, t, sq, occ)`. `Board.Candidate()` returns fully legal moves: pseudo-legal generation, then pin-based legality (`legalInfo` in legal.go computes the enemy danger map with the king removed from occupancy, checkers via reverse lookup, pin lines, and block squares once per position — no per-move board copy). Uchifuzume (pawn-drop mate) is the one case still verified by applying the move. The old copy-based filter survives as `legalCandidateSlow`, used as the reference in the random-playout differential test.

**Applying moves**: two tiers. `Action()`/`Copy()` maintain hash + history (USI/game path); `DoMove`/`UndoMove` (make/unmake with incremental Zobrist XOR) are the search/perft path — every `DoMove` must be paired with an `UndoMove` on all return paths. `Move(a)` is the validated API for untrusted input (checks membership in `Candidate()`).

**Rules**: `IsCheck`, repetition (`Repetition()` — draw on 4-fold, perpetual-check loss, in repetition.go), 入玉宣言 27-point rule (`CanDeclareWin()`, declare.go), nifu / forced promotion / uchifuzume inside generation. Zobrist tables in zobrist.go (fixed seed); `computeHash()` full recompute must stay consistent with DoMove's incremental XOR — the random-playout test in zobrist_test.go enforces this.

**Search** (`search/`): `Best`/`BestContext` — iterative-deepening negamax with alpha-beta, quiescence at depth 0 (captures only, full 1-ply extension when in check, 8-ply cap; quiesce.go), move ordering ttBest > captures (value desc) > killers > history (search.go), material + PST evaluation (eval.go, black-oriented tables mirrored via `80-sq`), mate scores as `MateScore-ply`. Transposition table (tt.go) is lock-free xor-verify (`data`/`check=key^data` via atomics, moves packed into 20 bits and rebuilt through `NewAction`). `Options.Parallel` runs Lazy SMP: workers share the TT, each with own board copy and killer/history; worker 0 is authoritative and staggered depths diversify the rest. `Mate()` (mate.go) is AND-OR mate search used by `go mate`.

**USI layer** (usi.go, shogi.go, engine.go): scanner goroutine + sender-channel select loop. `go` runs asynchronously so `stop` can cancel via context; time budgeting from `btime/wtime/byoyomi/binc/winc` (own/40 + byoyomi×0.8 + inc); `info depth/score cp/nodes/pv` lines flow through the sender channel. Command lines are tokenized with `strings.Fields` — do not revert to `strings.Split` (trailing-whitespace input used to panic the engine).

**Coordinates**: `x, y` are 1–9; USI strings are file `10-x`, rank `'a'+y-1`. Black moves in the −y direction. Uppercase SFEN letters = black.

## Tests

Test conventions worth preserving:
- **perft** (perft_test.go): depths 1–4 must stay 30 / 900 / 25470 / 719731. This is the primary correctness gate for anything touching move generation or make/unmake; never adjust the expected values.
- **Golden tests** (can_golden_test.go + golden_values_test.go): pseudo-legal candidate sets for 5 fixed positions, asserted against `ExportBoardPseudoCandidate`.
- **Differential tests**: random playouts with fixed seeds comparing fast vs. reference implementations (legal_diff_test.go: fast vs. `legalCandidateSlow`; zobrist_test.go: incremental vs. `computeHash`). Extend this pattern when adding an optimized path beside a simple one.
- Private functions are exposed to `shogi_test` via export_test.go — add exports there rather than widening visibility.
- Benchmark functions keep a comment trail of historical numbers; append new results rather than replacing old ones.

## Status (2026-07 rebuild)

Implemented: bitboard move generation with precomputed attacks; fully legal `Candidate()` (check evasion, pins, uchifuzume); repetition + perpetual check; 27-point declaration; make/unmake with incremental Zobrist; alpha-beta + quiescence + killers/history + PST; lock-free shared TT + Lazy SMP; USI info/stop/time management/setoption/`go mate`; validated `Move()` for untrusted input.

Known gaps / future work: 無駄合い (useless interposition) not pruned in mate search; `Action()` still recomputes hash fully (only `DoMove` is incremental); PST values are hand-rolled and untuned; no ponder support; no SFEN serialization of `Board`; `setMoveInfo` clears the source square before promotion validation (harmless today — unreachable via validated callers — but relevant if `DoMove` is ever exposed to unvalidated input).
