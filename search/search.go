package search

import (
	"fmt"
	"time"

	. "github.com/amanjpro/zahak/cache"
	. "github.com/amanjpro/zahak/engine"
	. "github.com/amanjpro/zahak/evaluation"
)

var STOP_SEARCH_GLOBALLY = false

var nodesVisited int64 = 0
var nodesSearched int64 = 0
var cacheHits int64 = 0
var pv = NewPVLine(100)

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
	start := time.Now()
	bestMove, score := startMinimax(position, depth, ply)
	bestEval = EvalMove{score, bestMove}
	end := time.Now()
	fmt.Printf("Visited: %d, Selected: %d, Cache-hit: %d\n\n", nodesVisited, nodesSearched, cacheHits)
	fmt.Printf("Took %f seconds\n\n", end.Sub(start).Seconds())
	pv.Pop() // pop our move
	pv.Pop() // pop our opponent's move
	return bestEval
}

func startMinimax(position *Position, depth int8, ply uint16) (*Move, int) {

	// Collect evaluation for moves per iteration to help us order moves for the next iteration
	legalMoves := position.LegalMoves()
	iterationEvals := make([]int, len(legalMoves))

	var bestMove *Move
	var previousBestMove *Move

	bestScore := -MAX_INT

	timeForSearch := 2 * time.Minute // TODO: with time management this should go
	fruitlessIterations := 0

	alpha := -MAX_INT
	beta := MAX_INT

	// wp := WhitePawn
	// aspirationWindow := wp.Weight() / 4
	start := time.Now()
	for iterationDepth := int8(1); iterationDepth <= depth; iterationDepth++ {
		currentBestScore := -MAX_INT
		orderedMoves := orderIterationMoves(&IterationMoves{legalMoves, iterationEvals})
		line := NewPVLine(iterationDepth + 1)
		for index, move := range orderedMoves {
			if time.Now().Sub(start) > timeForSearch {
				if index != 0 {
					return bestMove, currentBestScore
				} else {
					return bestMove, bestScore
				}
			}
			fmt.Printf("info currmove %s currmovenumber %d\n\n", move.ToString(), index+1)
			sendPv := false
			cp, ep, tg := position.MakeMove(move)
			// score, newAlpha, newBeta, set := withAspirationWindow(position, iterationDepth, alpha, beta, ply, aspirationWindow, line)
			// if set && iterationDepth >= 4 {
			// 	alpha = newAlpha
			// 	beta = newBeta
			// }
			score := -pvSearch(position, iterationDepth, 1, -beta, -alpha, ply, line)
			// This only works, because checkmate eval is clearly distinguished from
			// maximum/minimum beta/alpha
			if score != MAX_INT && score != -MAX_INT { // no very hard alpha-beta cutoff
				iterationEvals[index] = score
			} else {
				iterationEvals[index] = -MAX_INT // if it is, then too bad, that is a bad move
			}
			position.UnMakeMove(move, tg, ep, cp)
			if score > currentBestScore && score < beta && score > alpha {
				sendPv = true
				pv.AddFirst(move)
				pv.ReplaceLine(line)
				currentBestScore = score
				bestMove = move
				bestScore = currentBestScore
			}
			if score == CHECKMATE_EVAL {
				return move, score
			}
			timeSpent := time.Now().Sub(start)
			if sendPv {
				fmt.Printf("info depth %d nps %d tbhits %d nodes %d score cp %d time %d pv %s",
					iterationDepth, nodesVisited/1000*int64(timeSpent.Seconds()),
					cacheHits, nodesVisited, currentBestScore, timeSpent.Milliseconds(), pv.ToString())
			} else {
				fmt.Printf("info depth %d nps %d tbhits %d nodes %d score cp %d time %d",
					iterationDepth, nodesVisited/1000*int64(timeSpent.Seconds()),
					cacheHits, nodesVisited, currentBestScore, timeSpent.Milliseconds())
			}
			fmt.Printf("\n\n")
		}
		if iterationDepth >= 5 && *previousBestMove == *bestMove {
			if fruitlessIterations <= 3 {
				fruitlessIterations++
			} else {
				break
			}
		} else {
			fruitlessIterations = 0
		}
		previousBestMove = bestMove
		// if true && iterationDepth >= 4 {
		// 	alpha = currentBestScore
		// 	beta = currentBestScore
		// }
		currentBestScore = -MAX_INT
	}
	return bestMove, bestScore
}

func withAspirationWindow(position *Position, depth int8, alpha int, beta int, ply uint16, window int, pvline *PVLine) (int, int, int, bool) {

	for trials := 1; trials <= 3; trials++ {
		score := -alphaBeta(position, depth, 1, -beta, -alpha, ply, pvline)
		currentWindow := trials * window
		if score <= alpha {
			alpha -= currentWindow
		} else if score >= beta {
			beta += currentWindow
		} else {
			return score, alpha, beta, true
		}
	}
	return -alphaBeta(position, depth, 1, MAX_INT, -MAX_INT, ply, pvline), -MAX_INT, MAX_INT, false
}

func alphaBeta(position *Position, depthLeft int8, searchHeight int8, alpha int, beta int, ply uint16, pvline *PVLine) int {
	nodesVisited += 1
	outcome := position.Status()
	if outcome == Checkmate {
		return -CHECKMATE_EVAL
	} else if outcome == Draw {
		return 0
	}

	if depthLeft == 0 || STOP_SEARCH_GLOBALLY {
		// if searchHeight >= 4 {
		// 	return Evaluate(position)
		// }
		return quiescence(position, alpha, beta, 0)
	}

	nodesSearched += 1

	legalMoves := position.LegalMoves()
	orderedMoves := orderMoves(&ValidMoves{position, legalMoves, searchHeight})

	hash := position.Hash()
	cachedEval, found := TranspositionTable.Get(hash)
	if found &&
		(cachedEval.Eval == CHECKMATE_EVAL ||
			cachedEval.Eval == -CHECKMATE_EVAL ||
			cachedEval.Depth >= depthLeft) {
		cacheHits += 1
		score := cachedEval.Eval
		if score == CHECKMATE_EVAL || score == -CHECKMATE_EVAL {
			return cachedEval.Eval
		}
		if score >= beta && (cachedEval.Type != LowerBound || cachedEval.Type == Exact) {
			return beta
		}
		if score <= alpha && (cachedEval.Type != UpperBound || cachedEval.Type == Exact) {
			return alpha
		}
		if cachedEval.Type == Exact && score < beta && score > alpha {
			return score
		}
	}

	foundExact := false
	for _, move := range orderedMoves {
		capturedPiece, oldEnPassant, oldTag := position.MakeMove(move)
		line := NewPVLine(depthLeft - 1)
		score := -alphaBeta(position, depthLeft-1, searchHeight+1, -beta, -alpha, ply, line)
		position.UnMakeMove(move, oldTag, oldEnPassant, capturedPiece)
		if score >= beta {
			TranspositionTable.Set(hash, &CachedEval{hash, score, depthLeft, UpperBound, ply})
			return score
		}
		if score > alpha {
			TranspositionTable.Set(hash, &CachedEval{hash, alpha, depthLeft, LowerBound, ply})
			pvline.AddFirst(move)
			pvline.ReplaceLine(line)
			// Potential PV move, lets copy it to the current pv-line
			foundExact = true
			alpha = score
		}
	}
	if foundExact {
		TranspositionTable.Set(hash, &CachedEval{hash, alpha, depthLeft, Exact, ply})
	}
	return alpha
}

func pvSearch(position *Position, depthLeft int8, searchHeight int8, alpha int, beta int, ply uint16, pvline *PVLine) int {
	nodesVisited += 1
	outcome := position.Status()
	if outcome == Checkmate {
		return -CHECKMATE_EVAL
	} else if outcome == Draw {
		return 0
	}

	if depthLeft == 0 || STOP_SEARCH_GLOBALLY {
		// if searchHeight >= 4 {
		// 	return Evaluate(position)
		// }
		return quiescence(position, alpha, beta, 0)
	}

	nodesSearched += 1

	legalMoves := position.LegalMoves()
	orderedMoves := orderMoves(&ValidMoves{position, legalMoves, searchHeight})

	hash := position.Hash()
	cachedEval, found := TranspositionTable.Get(hash)
	if found &&
		(cachedEval.Eval == CHECKMATE_EVAL ||
			cachedEval.Eval == -CHECKMATE_EVAL ||
			cachedEval.Depth >= depthLeft) {
		cacheHits += 1
		score := cachedEval.Eval
		if score == CHECKMATE_EVAL || score == -CHECKMATE_EVAL {
			return cachedEval.Eval
		}
		if score >= beta && (cachedEval.Type != LowerBound || cachedEval.Type == Exact) {
			return beta
		}
		if score <= alpha && (cachedEval.Type != UpperBound || cachedEval.Type == Exact) {
			return alpha
		}
		if cachedEval.Type == Exact && score < beta && score > alpha {
			return score
		}
	}

	searchPv := false

	foundExact := false
	for _, move := range orderedMoves {
		capturedPiece, oldEnPassant, oldTag := position.MakeMove(move)
		line := NewPVLine(depthLeft - 1)
		score := -MAX_INT
		if searchPv {
			score = -pvSearch(position, depthLeft-1, searchHeight+1, -beta, -alpha, ply, line)
		} else {
			score = -zeroWindowSearch(position, depthLeft-1, searchHeight+1, -alpha, ply)
			if score > alpha { // in fail-soft ... && score < beta ) is common
				score = -pvSearch(position, depthLeft-1, searchHeight+1, -beta, -alpha, ply, line) // re-search
			}
		}
		position.UnMakeMove(move, oldTag, oldEnPassant, capturedPiece)
		if score >= beta {
			TranspositionTable.Set(hash, &CachedEval{hash, score, depthLeft, UpperBound, ply})
			return score
		}
		if score > alpha {
			TranspositionTable.Set(hash, &CachedEval{hash, alpha, depthLeft, LowerBound, ply})
			// Potential PV move, lets copy it to the current pv-line
			pvline.AddFirst(move)
			pvline.ReplaceLine(line)
			foundExact = true
			searchPv = false
			alpha = score
		}
	}
	if foundExact {
		TranspositionTable.Set(hash, &CachedEval{hash, alpha, depthLeft, Exact, ply})
	}
	return alpha
}

func zeroWindowSearch(position *Position, depthLeft int8, searchHeight int8, beta int, ply uint16) int {
	nodesVisited += 1
	outcome := position.Status()
	if outcome == Checkmate {
		return -CHECKMATE_EVAL
	} else if outcome == Draw {
		return 0
	}

	if depthLeft == 0 || STOP_SEARCH_GLOBALLY {
		// return quiescence(p, beta-1, beta, 0)
		// if searchHeight >= 4 {
		// 	return Evaluate(position)
		// }
		return quiescence(position, beta-1, beta, 0)
	}

	nodesSearched += 1

	legalMoves := position.LegalMoves()
	orderedMoves := orderMoves(&ValidMoves{position, legalMoves, searchHeight})

	hash := position.Hash()
	cachedEval, found := TranspositionTable.Get(hash)
	if found &&
		(cachedEval.Eval == CHECKMATE_EVAL ||
			cachedEval.Eval == -CHECKMATE_EVAL ||
			cachedEval.Depth >= depthLeft) {
		cacheHits += 1
		score := cachedEval.Eval
		if score == CHECKMATE_EVAL || score == -CHECKMATE_EVAL {
			return cachedEval.Eval
		}
		if score >= beta && (cachedEval.Type != LowerBound || cachedEval.Type == Exact) {
			return beta
		}
		// if score <= alpha && (cachedEval.Type != UpperBound || cachedEval.Type == Exact) {
		// 	return alpha
		// }
		if cachedEval.Type == Exact && score < beta {
			return score
		}
	}

	for _, move := range orderedMoves {
		capturedPiece, oldEnPassant, oldTag := position.MakeMove(move)
		score := -zeroWindowSearch(position, depthLeft-1, searchHeight+1, 1-beta, ply)
		position.UnMakeMove(move, oldTag, oldEnPassant, capturedPiece)
		if score >= beta {
			return beta // fail-hard beta-cutoff
		}
	}
	return beta - 1 // fail-hard, return alpha
}
