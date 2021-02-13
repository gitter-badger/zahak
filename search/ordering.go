package search

import (
	"sort"

	. "github.com/amanjpro/zahak/cache"
	. "github.com/amanjpro/zahak/engine"
)

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

	// King safety (castling)
	castling := KingSideCastle | QueenSideCastle
	move1IsCastling := move1.HasTag(castling)
	move2IsCastling := move2.HasTag(castling)
	if move1IsCastling && !move2IsCastling {
		return true
	} else if move2IsCastling && !move1IsCastling {
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
