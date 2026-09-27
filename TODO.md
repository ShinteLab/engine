# engine の TODO

`engine` を**単体で使えるエンジンとして仕上げる**ための積み残し。
実装状況そのものは `AGENTS.md` の「Status」節にあり、ここはその先の話。

## 前提が変わった（2026-08-08）

**`ikkyoku` は `engine` を Go の関数として import するのをやめ、USI を話す相手として
繋ぐことにした**（`ikkyoku/_docs/phase4-engine-usi.md` の案 B）。

```
ikkyoku(USIクライアント) ──stdin/stdout(USI)──> エンジンのプロセス
```

繋ぎ先は「USI を話すプロセス」なら何でもよく、やねうら王・水匠も `prokishi.exe` も
同じ口に入る。**`engine` はその選択肢の 1 つ**という位置づけになった。

これが意味すること:

- **`engine` の外向きの顔は Go の API ではなく USI になる。** 使いにくさが出るのは
  `usi.go` の側で、`search` の API ではない
- **最終的には `engine` が「エンジンの exe」を配る。** 今の `_cmd/think.go` +
  `_samples.ThinkEngine` はサンプルであって製品ではない
- **棋力と USI の作りが、そのまま検討ツールとしての使い勝手になる。**
  ikkyoku の構想（「次善手を選んだらどう転ぶかを辿る」）は候補手が複数並ぶことが前提

---

## 1. MultiPV が無い（**最優先**）

**ここが埋まらないと ikkyoku の中核が作れない。** 「最善手ではない手を選んだら
どうなるか」を辿るのが構想の中心で、`bestmove` 1 個では足りない。

- [ ] `search.Options` に `MultiPV int` を足す（`Result` は単一手・除外手リストも無い）
- [ ] ルートで上位 N 手を保持する（今は `searchRoot` が最善 1 手だけを返す）
- [ ] `info multipv <n>` を出す（下記 2 と一緒）
- [ ] `usi` 応答で `option name MultiPV type spin default 1 min 1 max 128` を宣言する

⚠️ **αβ は「上位 N 手を正しく順位付けする」ようにはできていない。** 2 番手以降は
枝刈りで正確な値が付かないので、MultiPV では**ルートの各手を独立した窓で探索する**
必要がある（一般的な実装では aspiration + 除外リスト）。単に上位 N 個を拾うだけでは
順位も評価値も嘘になる。

## 2. `info` 行が痩せている（`usi.go` の `formatInfo`）

今出しているのは `info depth <d> score cp <v> nodes <n> [pv <手>]` だけ。

- [ ] **`score mate <n>`。** `Info.Mate` はフィールドがあるだけで `search` が埋めていない。
      詰みも `score cp 1048571`（`MateScore - ply` の生値）として出るので、
      GUI 側は自前で「MateScore に近ければ詰み」と判定するしかない
- [ ] **`multipv <n>`**（上記 1 と対）
- [ ] **`time <ms>`。** 経過時間が出せないと GUI 側で NPS も進捗も出せない
- [ ] `nps` / `seldepth` / `hashfull`
- [ ] **`pv` が 1 手しかない。** `iterativeDeepen` が `PV: []*Action{res.Action}` を
      渡しているだけで、置換表から読み筋を復元していない。**読み筋は検討 UI の本体**
- [ ] `info string`（エンジン側の任意メッセージ。デバッグに効く）

## 3. `go` のパラメータが足りない（`usi.go` の `parseGoTime`）

読んでいるのは `btime` / `wtime` / `byoyomi` / `binc` / `winc` だけ。

- [ ] **`go movetime <ms>`。** 対局用の時間配分ではなく「この局面をこれだけ考えて」
      という指定で、**検討ツールが一番使うのはこれ**。今は解釈されず 0（無制限）に落ちる
- [ ] **`go infinite`。** 結果的に無制限にはなるが、意図して扱っているわけではない
      （`stop` が来るまで返さない、という契約を明示したい）
- [ ] `go depth <n>` / `go nodes <n>`
- [ ] `go ponder` / `ponderhit`（`AGENTS.md` の Known gaps にもある）

## 4. exe として配る

- [ ] **製品としてのエンジン本体を決める。** `_samples.ThinkEngine` は
      `thinkDepth = 4` 固定のサンプルで、時間制御もオプションも持たない
- [ ] `_samples` / `_cmd` は `_` 始まりなので `go build ./...` の対象外。
      **配る exe は通常のディレクトリに置く**（他プロジェクトから import される
      可能性のある `_samples` を実体にしない）
- [ ] `setoption` で深さ・スレッド数・置換表サイズを変えられるようにし、
      `usi` 応答で `option` として宣言する（今は `setOption` が受け取るだけで宣言が無い）
- [ ] ログの出力先。今は**カレントディレクトリに `shogi_<pid>.log`** を作る。
      GUI から起動されると GUI 側の作業ディレクトリに散らかる
- [ ] 名前・バージョン・作者を埋める（`_samples` は "thinking Engine" / "0.0.0"）

## 5. USI の細かい穴

- [ ] **入玉宣言の出口が無い。** `CanDeclareWin()`（27 点法）は実装されているのに、
      USI 側に `bestmove win` を送る経路が無い
- [ ] `gameover` を無視している（`usinewgame` も何もしない）。
      置換表の破棄など、局面が変わったときの後始末が要るなら入口はここ
- [ ] `usi` 応答が `id name` / `id author` / `usiok` だけ。`option` の宣言が 1 つも無い
- [ ] `stop` に対する `bestmove` は返るが、**`stop` 前に何も探索していない場合**の
      挙動が未確認（`go` 直後の `stop`）
- [ ] 千日手・連続王手の判定は履歴に依存する。**ikkyoku は履歴の無い局面を渡す**ので、
      `position sfen ...`（moves 無し）では判定できない。これは仕様として受け入れる

## 6. 棋力

**ikkyoku から見ると、ここが「使えるかどうか」を決める。**

- [ ] PST が手作り・未調整（`AGENTS.md` の Known gaps）。まず評価関数の素性を決める
- [ ] `Action()` が hash を毎回再計算している（`DoMove` だけが差分更新）
- [ ] 無駄合いを詰み探索で枝刈りしていない
- [ ] 定跡が無い

## 参考

- 検討の経緯: `ikkyoku/_docs/phase4-engine-usi.md`
- 全体の構想: ワークスペース（`ShinteLab/shinte`）の `TODO.md`
