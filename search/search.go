package search

import (
	"fmt"
	"sort"
	"time"

	. "github.com/amanjpro/zahak/cache"
	. "github.com/amanjpro/zahak/engine"
	. "github.com/amanjpro/zahak/evaluation"
)

var STOP_SEARCH_GLOBALLY = false

var nodesVisited int64 = 0
var nodesSearched int64 = 0
var cacheHits int64 = 0
var pv = make([]*Move, 100)

type EvalMove struct {
	eval int
	move *Move
}

func (e *EvalMove) Move() *Move {
	return e.move
}

func (e *EvalMove) Eval() int {
	return e.eval
}

func Search(position *Position, depth int8, ply uint16) EvalMove {
	STOP_SEARCH_GLOBALLY = false
	nodesVisited = 0
	nodesSearched = 0
	cacheHits = 0
	var bestEval EvalMove
	var isMaximizingPlayer = position.Turn() == White
	// validMoves := position.LegalMoves()
	// evals := make(chan EvalMove)
	// evalIsSet := false
	start := time.Now()
	// for i := 0; i < len(validMoves); i++ {
	// 	p := position.copy()
	// 	move := validMoves[i]
	// 	p.MakeMove(move)
	// 	go parallelMinimax(p, move, depth, !isMaximizingPlayer, evals)
	// }
	// for i := 0; i < len(validMoves); i++ {
	// 	evalMove := <-evals
	//
	// 	mvStr := evalMove.move.ToString()
	// 	fmt.Printf("info nodes %d score cp %d currmove %s pv %s",
	// 		nodesVisited, int(evalMove.eval*100*dir), mvStr, mvStr)
	// 	for _, mv := range evalMove.line {
	// 		fmt.Printf(" %s", mv.ToString())
	// 	}
	// 	fmt.Print("\n\n")
	// 	if isMaximizingPlayer {
	// 		if !evalIsSet || evalMove.eval > bestEval.eval {
	// 			bestEval = evalMove
	// 			evalIsSet = true
	// 		}
	// 	} else {
	// 		if !evalIsSet || evalMove.eval < bestEval.eval {
	// 			bestEval = evalMove
	// 			evalIsSet = true
	// 		}
	// 	}

	bestMove, score := startMinimax(position, depth, isMaximizingPlayer, ply)
	// fmt.Printf("info nodes %d score cp %d currmove %s",
	// nodesVisited, int(score*dir), bestMove.ToString())
	// for i := 1; i < len(moves); i++ {
	// 	mv := moves[i]
	// 	fmt.Printf(" %s", mv.ToString())
	// }
	fmt.Print("\n\n")

	bestEval = EvalMove{score, bestMove}
	// }
	end := time.Now()
	// close(evals)
	fmt.Printf("Visited: %d, Selected: %d, Cache-hit: %d\n\n", nodesVisited, nodesSearched, cacheHits)
	fmt.Printf("Took %f seconds\n\n", end.Sub(start).Seconds())
	return bestEval
}

func startMinimax(position *Position, depth int8,
	isMaximizingPlayer bool, ply uint16) (*Move, int) {
	legalMoves := position.LegalMoves()
	iterationEvals := make([]int, len(legalMoves))
	var bestMove *Move
	var bestScore int
	var dir = -1
	if isMaximizingPlayer {
		dir = 1
	}

	timeForSearch := 5 * time.Minute
	start := time.Now()

	for iterationDepth := int8(1); iterationDepth <= depth; iterationDepth++ {
		alpha := -MAX_INT
		beta := MAX_INT
		bestScore = -1 * dir * MAX_INT
		orderedMoves := orderIterationMoves(&IterationMoves{legalMoves, iterationEvals})
		for index, move := range orderedMoves {
			sendPv := false
			if time.Now().Sub(start) > timeForSearch {
				return bestMove, bestScore
			}
			cp, ep, tg := position.MakeMove(move)
			score := minimax(position, iterationDepth, 0, !isMaximizingPlayer, alpha, beta, ply)
			iterationEvals[index] = score
			position.UnMakeMove(move, tg, ep, cp)
			if isMaximizingPlayer {
				if score > bestScore {
					sendPv = true
					bestScore = score
					bestMove = move
				}
				if score > alpha {
					alpha = score
				}
				if score == CHECKMATE_EVAL {
					return move, CHECKMATE_EVAL
				}
			} else {
				if score < bestScore {
					sendPv = true
					bestScore = score
					bestMove = move
				}
				if score < beta {
					beta = score
				}
				if score == -CHECKMATE_EVAL {
					return move, -CHECKMATE_EVAL
				}
			}
			fmt.Printf("info currmove %s\n\n", bestMove.ToString())
			timeSpent := time.Now().Sub(start)
			if sendPv {
				fmt.Printf("info depth %d nps %d tbhits %d nodes %d score cp %d time %d pv %s",
					iterationDepth, nodesVisited/1000*int64(timeSpent.Seconds()),
					cacheHits, nodesVisited, int(bestScore*dir), timeSpent.Milliseconds(), bestMove.ToString())
				for i, move := range pv {
					if move == nil {
						break
					}
					fmt.Printf(" %s", move.ToString())
					pv[i] = nil
				}
			} else {
				fmt.Printf("info depth %d nps %d tbhits %d nodes %d score cp %d time %d",
					iterationDepth, nodesVisited/1000*int64(timeSpent.Seconds()),
					cacheHits, nodesVisited, int(bestScore*dir), timeSpent.Milliseconds())
			}
			fmt.Printf("\n\n")
		}
	}
	return bestMove, bestScore
}

func minimax(position *Position, depthLeft int8, searchHeight int8,
	isMaximizingPlayer bool, alpha int, beta int, ply uint16) int {
	nodesVisited += 1

	if depthLeft == 0 || STOP_SEARCH_GLOBALLY {
		// TODO: Perform all captures before giving up, to avoid the horizon effect
		// var dir float64 = -1
		// if isMaximizingPlayer {
		// 	dir = 1
		// }
		evl := Evaluate(position)
		// fmt.Printf("info nodes %d score cp %d currmove %s pv",
		// 	nodesVisited, int(evl*100*dir), baseMove.ToString())
		// for _, mv := range line {
		// 	fmt.Printf(" %s", mv.ToString())
		// }
		// fmt.Print("\n\n")

		return evl
	}

	nodesSearched += 1
	legalMoves := position.LegalMoves()
	orderedMoves := orderMoves(&ValidMoves{position, legalMoves})
	var bestMove *Move

	for _, move := range orderedMoves {
		score := getEval(position, depthLeft-1, searchHeight+1, !isMaximizingPlayer, alpha, beta, move, ply)
		if isMaximizingPlayer {
			if score >= beta {
				return beta
			}
			if score > alpha {
				bestMove = move
				alpha = score
			}
		} else {
			if score <= alpha {
				return alpha
			}
			if score < beta {
				bestMove = move
				beta = score
			}
		}
	}

	if len(orderedMoves) == 0 {
		return Evaluate(position)
	} else if isMaximizingPlayer {
		pv[searchHeight] = bestMove
		return alpha
	} else {
		pv[searchHeight] = bestMove
		return beta
	}
}

func getEval(position *Position, depthLeft int8, searchHeight int8, isMaximizingPlayer bool,
	alpha int, beta int, move *Move, ply uint16) int {
	var score int
	capturedPiece, oldEnPassant, oldTag := position.MakeMove(move)
	newPositionHash := position.Hash()
	cachedEval, found := TranspositionTable.Get(newPositionHash)
	if found &&
		(cachedEval.Eval == CHECKMATE_EVAL ||
			cachedEval.Eval == -CHECKMATE_EVAL ||
			cachedEval.Depth >= depthLeft) {
		cacheHits += 1
		score = cachedEval.Eval
	} else {
		v := minimax(position, depthLeft, searchHeight, isMaximizingPlayer, alpha, beta, ply)
		var tpe NodeType
		if isMaximizingPlayer {
			if v >= beta {
				tpe = LowerBound
			}
			if score > alpha {
				tpe = Exact
			}
		} else {
			if v <= alpha {
				tpe = UpperBound
			}
			if v < beta {
				tpe = Exact
			}
		}
		TranspositionTable.Set(newPositionHash, &CachedEval{position.Hash(), v, depthLeft, tpe, ply})
		score = v
	}
	position.UnMakeMove(move, oldTag, oldEnPassant, capturedPiece)
	return score
}

type ValidMoves struct {
	position *Position
	moves    []*Move
}

func (validMoves *ValidMoves) Len() int {
	return len(validMoves.moves)
}

func (validMoves *ValidMoves) Swap(i, j int) {
	moves := validMoves.moves
	moves[i], moves[j] = moves[j], moves[i]
}

func (validMoves *ValidMoves) Less(i, j int) bool {
	moves := validMoves.moves
	move1, move2 := moves[i], moves[j]
	board := validMoves.position.Board
	// // Is in PV?
	// if pv != nil && len(pv) > validMoves.depth {
	// 	if pv[validMoves.depth] == move1 {
	// 		return true
	// 	}
	// }

	// Is in Transition table ???
	// TODO: This is slow, that tells us either cache access is slow or has computation is
	// Or maybe (unlikely) make/unmake move is slow
	cp1, ep1, tg1 := validMoves.position.MakeMove(move1)
	hash1 := validMoves.position.Hash()
	validMoves.position.UnMakeMove(move1, tg1, ep1, cp1)
	eval1, ok1 := TranspositionTable.Get(hash1)

	cp2, ep2, tg2 := validMoves.position.MakeMove(move2)
	hash2 := validMoves.position.Hash()
	validMoves.position.UnMakeMove(move2, tg2, ep2, cp2)
	eval2, ok2 := TranspositionTable.Get(hash2)

	if ok1 && ok2 {
		if eval1.Type == Exact && eval2.Type != Exact {
			return true
		} else if eval2.Type == Exact && eval1.Type != Exact {
			return false
		}
		if eval1.Eval > eval2.Eval ||
			(eval1.Eval == eval2.Eval && eval1.Depth >= eval2.Depth) {
			return true
		} else if eval1.Eval < eval2.Eval {
			return false
		}
	} else if ok1 {
		return true
	} else if ok2 {
		return false
	}
	//
	// capture ordering
	if move1.HasTag(Capture) && move2.HasTag(Capture) {
		// What are we capturing?
		piece1 := board.PieceAt(move1.Destination)
		piece2 := board.PieceAt(move2.Destination)
		if piece1.Type() > piece2.Type() {
			return true
		}
		// Who is capturing?
		piece1 = board.PieceAt(move1.Source)
		piece2 = board.PieceAt(move2.Source)
		if piece1.Type() <= piece2.Type() {
			return true
		}
		return false
	} else if move1.HasTag(Capture) {
		return true
	}

	piece1 := board.PieceAt(move1.Source)
	piece2 := board.PieceAt(move2.Source)

	// prefer checks
	if move1.HasTag(Check) {
		return true
	}
	if move2.HasTag(Check) {
		return false
	}
	// Prefer smaller pieces
	if piece1.Type() <= piece2.Type() {
		return true
	}

	return false
}

func orderMoves(validMoves *ValidMoves) []*Move {
	sort.Sort(validMoves)
	return validMoves.moves
}

type IterationMoves struct {
	moves []*Move
	evals []int
}

func (iter *IterationMoves) Len() int {
	return len(iter.moves)
}

func (iter *IterationMoves) Swap(i, j int) {
	evals := iter.evals
	moves := iter.moves
	moves[i], moves[j] = moves[j], moves[i]
	evals[i], evals[j] = evals[j], evals[i]
}

func (iter *IterationMoves) Less(i, j int) bool {
	evals := iter.evals
	return evals[i] <= evals[j]
}

func orderIterationMoves(iter *IterationMoves) []*Move {
	sort.Sort(iter)
	return iter.moves
}