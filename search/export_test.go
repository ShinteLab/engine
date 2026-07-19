package search

import "context"
import "shogi"

//this test file exports private symbols for search_test package tests

// negamax は searcher のメソッドになったため、使い捨て searcher(tt のみ
// 引き継ぐ)を生成して呼び出すラッパを公開する。killer/history の検証は
// このラッパでは行わない(既存テストは repetition の早期return確認のみ)。
func ExportNegamax(ctx context.Context, b *shogi.Board, depth, ply, alpha, beta int, nodes *int64, tt *transTable) int {
	s := newSearcher(tt)
	return s.negamax(ctx, b, depth, ply, alpha, beta, nodes)
}

var ExportEval = eval
