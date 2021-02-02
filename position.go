package main

import (
	"fmt"
)

type Position struct {
	board     Bitboard
	enPassant Square
	tag       PositionTag
}

type PositionTag uint8

const (
	WhiteCanCastleKingSide PositionTag = 1 << iota
	WhiteCanCastleQueenSide
	BlackCanCastleKingSide
	BlackCanCastleQueenSide
	BlackToMove
	WhiteToMove
)

func (p *Position) SetTag(tag PositionTag)      { p.tag |= tag }
func (p *Position) ClearTag(tag PositionTag)    { p.tag &= ^tag }
func (p *Position) ToggleTag(tag PositionTag)   { p.tag ^= tag }
func (p *Position) HasTag(tag PositionTag) bool { return p.tag&tag != 0 }

func (p *Position) Turn() Color {
	if p.HasTag(WhiteToMove) {
		return White
	}
	return Black
}

func (p *Position) ToggleTurn() {
	p.ToggleTag(BlackToMove)
	p.ToggleTag(WhiteToMove)
}

func (p *Position) MakeMove(move Move) {
	p.board.Move(move.source, move.destination)
	movingPiece := p.board.PieceAt(move.source)

	p.enPassant = findEnPassantSquare(move, movingPiece)

	// Do promotion
	if move.promoType != NoType {
		p.board.UpdateSquare(move.destination, getPiece(move.promoType, p.Turn()))
	}

	if movingPiece == BlackKing {
		p.ClearTag(BlackCanCastleKingSide)
		p.ClearTag(BlackCanCastleQueenSide)
	} else if movingPiece == WhiteKing {
		p.ClearTag(WhiteCanCastleKingSide)
		p.ClearTag(WhiteCanCastleQueenSide)
	} else if movingPiece == BlackRook && move.source == A8 {
		p.ClearTag(BlackCanCastleQueenSide)
	} else if movingPiece == BlackRook && move.source == H8 {
		p.ClearTag(BlackCanCastleKingSide)
	} else if movingPiece == WhiteRook && move.source == A1 {
		p.ClearTag(WhiteCanCastleQueenSide)
	} else if movingPiece == WhiteRook && move.source == H1 {
		p.ClearTag(WhiteCanCastleKingSide)
	}

	p.ToggleTurn()
}

func (p *Position) UnMakeMove(move Move, tag PositionTag, enPassant Square, capturedPiece Piece) {
	p.tag = tag
	p.enPassant = enPassant
	p.board.Move(move.destination, move.source)

	// Undo enpassant
	if move.HasTag(EnPassant) && move.HasTag(Capture) {
		movingPiece := p.board.PieceAt(move.destination)
		sq := findEnPassantSquare(move, movingPiece)
		p.board.UpdateSquare(sq, capturedPiece)
	} else if move.HasTag(Capture) { // Undo capture
		p.board.UpdateSquare(move.destination, capturedPiece)
	}

	// Undo promotion
	if move.promoType != NoType {
		p.board.UpdateSquare(move.source, getPiece(Pawn, p.Turn()))
	}
	if move.HasTag(QueenSideCastle) {
		// white
		if move.destination == C1 {
			p.board.Move(D1, A1)
		} else { // black
			p.board.Move(D8, A8)
		}
	} else if move.HasTag(KingSideCastle) {
		// white
		if move.destination == G1 {
			p.board.Move(F1, H1)
		} else { // black
			p.board.Move(F8, H8)
		}
	}
}

func (p *Position) Hash() uint64 {
	return generateZobristHash(p)
}

func (p *Position) ValidMoves() []Move {
	var allMoves = make([]Move, 0, 256)
	var whiteChecksRook uint64 = 0
	var whiteChecksBishop uint64 = 0
	var whiteChecksQueen uint64 = 0
	var whiteChecksKnight uint64 = 0
	var whiteChecksPawn uint64 = 0
	// var whiteChecksKing uint64 = 0
	var blackChecksRook uint64 = 0
	var blackChecksBishop uint64 = 0
	var blackChecksQueen uint64 = 0
	var blackChecksKnight uint64 = 0
	var blackChecksPawn uint64 = 0
	// var blackChecksKing uint64 = 0

	board := p.board
	allPieces := board.AllPieces()
	blackKingIndex := NoSquare
	whiteKingIndex := NoSquare

	for sq, piece := range allPieces {
		if piece == WhiteRook {
			checks, moves := rookLikeMoves(&board, WhiteRook, sq, allMoves)
			whiteChecksRook |= checks
			allMoves = moves
		} else if piece == BlackRook {
			checks, moves := rookLikeMoves(&board, BlackRook, sq, allMoves)
			blackChecksRook |= checks
			allMoves = moves
		} else if piece == WhiteQueen {
			checks1, moves1 := rookLikeMoves(&board, WhiteQueen, sq, allMoves)
			checks2, moves2 := bishopLikeMoves(&board, WhiteQueen, sq, moves1)
			whiteChecksQueen |= (checks1 | checks2)
			allMoves = moves2
		} else if piece == BlackQueen {
			checks1, moves1 := rookLikeMoves(&board, BlackQueen, sq, allMoves)
			checks2, moves2 := bishopLikeMoves(&board, BlackQueen, sq, moves1)
			blackChecksQueen |= (checks1 | checks2)
			allMoves = moves2
		} else if piece == WhiteBishop {
			checks, moves := bishopLikeMoves(&board, WhiteBishop, sq, allMoves)
			whiteChecksBishop |= checks
			allMoves = moves
		} else if piece == BlackBishop {
			checks, moves := bishopLikeMoves(&board, BlackBishop, sq, allMoves)
			blackChecksBishop |= checks
			allMoves = moves
		} else if piece == WhiteKnight {
			checks, moves := knightMoves(&board, White, sq, &allMoves)
			whiteChecksKnight |= checks
			allMoves = moves
		} else if piece == BlackKnight {
			checks, moves := knightMoves(&board, Black, sq, &allMoves)
			blackChecksKnight |= checks
			allMoves = moves
		} else if piece == WhitePawn {
			checks, moves := pawnMoves(&board, White, sq, p.enPassant, &allMoves)
			blackChecksPawn |= checks
			allMoves = moves
		} else if piece == BlackPawn {
			checks, moves := pawnMoves(&board, Black, sq, p.enPassant, &allMoves)
			whiteChecksPawn |= checks
			allMoves = moves
		} else if piece == WhiteKing {
			whiteKingIndex = sq
		} else if piece == BlackKing {
			blackKingIndex = sq
		}
	}

	fmt.Println(blackKingIndex, whiteKingIndex)
	return allMoves
}

func pawnCheckTag(b *Bitboard, sq1 Square, sq2 Square) MoveTag {
	if sq1 != NoSquare {
		p := b.PieceAt(sq1)
		if p.Type() == King {
			return Check
		}
	}
	if sq2 != NoSquare {
		p := b.PieceAt(sq2)
		if p.Type() == King {
			return Check
		}
	}
	return 0
}

func pawnAttacks(b *Bitboard, color Color, src Square) (Square, Square) {
	direction := 1
	if color == Black {
		direction = -1
	}

	file := src.File()

	right := src + Square(direction*9)
	left := src + Square(direction*7)

	pr := b.PieceAt(right)
	pl := b.PieceAt(left)

	f := NoSquare
	s := NoSquare

	if file == FileH {
		if color == White && pl.Color() == Black {
			return left, s
		} else if pr.Color() == White {
			return right, s
		}
	} else if file == FileA {
		if color == White && pr.Color() == Black {
			return right, s
		} else if pl.Color() == White {
			return left, s
		}
	} else {

		if pr.Color() != color && pr != NoPiece {
			f = right
		}

		if pl.Color() != color && pl != NoPiece {
			s = left
		}
	}

	return f, s
}

func pawnMoves(b *Bitboard, color Color, srcSq Square,
	enPassant Square, allMoves *[]Move) (uint64, []Move) {
	moves := *allMoves
	var checkSet uint64 = 0
	rank := srcSq.Rank()
	file := srcSq.File()

	f, s := pawnAttacks(b, color, srcSq)
	if f != NoSquare {
		setCheckSet(checkSet, f)
	}
	if s != NoSquare {
		setCheckSet(checkSet, f)
	}

	baseRank := Rank2
	promoRank := Rank7
	direction := 1
	enPassantRank := Rank3
	if color == Black {
		promoRank = Rank2
		baseRank = Rank7
		enPassantRank = Rank6
		direction = -1
	}
	oneStep := SquareOf(file, Rank(int(rank)+direction))
	twoSteps := SquareOf(file, Rank(int(rank)+direction+direction))
	// forward 1 step
	if rank != promoRank && b.PieceAt(oneStep) == NoPiece {
		f, s := pawnAttacks(b, color, oneStep)
		tag := pawnCheckTag(b, f, s)
		moves = append(moves, Move{srcSq, oneStep, NoType, tag})
	}

	// forward 2 steps
	if rank == baseRank && b.PieceAt(oneStep) == NoPiece &&
		b.PieceAt(twoSteps) == NoPiece {
		f, s := pawnAttacks(b, color, twoSteps)
		tag := pawnCheckTag(b, f, s)
		moves = append(moves, Move{srcSq, twoSteps, NoType, tag})
	} else if rank == promoRank && b.PieceAt(oneStep) == NoPiece {
		// promotion
		moves = append(moves,
			Move{srcSq, twoSteps, Queen, queenCheckTag(b, color, oneStep)},
			Move{srcSq, twoSteps, Rook, rookLikeCheckTag(b, color, oneStep)},
			Move{srcSq, twoSteps, Bishop, bishopLikeCheckTag(b, color, oneStep)},
			Move{srcSq, twoSteps, Knight, knightCheckTag(b, color, oneStep)},
		)
	}

	// capture
	attackF, attackS := pawnAttacks(b, color, srcSq)

	fPiece := b.PieceAt(attackF)
	sPiece := b.PieceAt(attackS)

	if attackF != NoSquare && fPiece.Type() != King {
		f, s := pawnAttacks(b, color, attackF)
		tag := pawnCheckTag(b, f, s)
		moves = append(moves, Move{srcSq, attackF, NoType, tag | Capture})
	}

	if attackS != NoSquare && sPiece.Type() != King {
		f, s := pawnAttacks(b, color, attackS)
		tag := pawnCheckTag(b, f, s)
		moves = append(moves, Move{srcSq, attackS, NoType, tag | Capture})
	}

	// enpassant
	if enPassant != NoSquare && enPassant.Rank() == enPassantRank &&
		srcSq.Rank() == Rank(int(enPassantRank)+direction) {
		f, s := pawnAttacks(b, color, enPassant)
		tag := pawnCheckTag(b, f, s)
		moves = append(moves, Move{srcSq, enPassant, NoType, tag | Capture | EnPassant})
	}
	return checkSet, moves
}

func queenCheckTag(b *Bitboard, color Color, srcSq Square) MoveTag {
	tag := rookLikeCheckTag(b, color, srcSq)
	if tag != Check {
		tag = bishopLikeCheckTag(b, color, srcSq)
	}
	return tag
}

func rookLikeMoves(b *Bitboard, p Piece, srcSq Square,
	allMoves []Move) (uint64, []Move) {
	color := p.Color()
	var checkSet uint64 = 0

	// vertical up moves
	for destSq := srcSq; destSq >= 0; destSq -= 8 {
		piece := b.PieceAt(destSq)
		if piece == NoPiece {
			tag := rookLikeCheckTag(b, color, destSq)
			if tag != Check && p.Type() == Queen {
				tag = bishopLikeCheckTag(b, color, destSq)
			}
			allMoves = append(allMoves, Move{srcSq, destSq, NoType, tag})
			setCheckSet(checkSet, destSq)
		} else if piece.Color() == color { // Rook & Queen cannot jump
			break
		} else if piece.Type() != King { // This is a capture
			tag := rookLikeCheckTag(b, color, destSq)
			if tag != Check && p.Type() == Queen {
				tag = bishopLikeCheckTag(b, color, destSq)
			}
			allMoves = append(allMoves, Move{srcSq, destSq, piece.Type(), Capture | tag})
			break
		} else if piece.Type() == King {
			// This is a check? what is that? don't add anything yet
			// allMoves = append(allMoves, Move{srcSq, &destSq, piece.Type(), Capture})
			break
		}
	}

	// vertical down moves
	for destSq := srcSq; destSq < 64; destSq += 8 {
		piece := b.PieceAt(destSq)
		if piece == NoPiece {
			tag := rookLikeCheckTag(b, color, destSq)
			if tag != Check && p.Type() == Queen {
				tag = bishopLikeCheckTag(b, color, destSq)
			}
			allMoves = append(allMoves, Move{srcSq, destSq, NoType, tag})
			setCheckSet(checkSet, destSq)
		} else if piece.Color() == color { // Rook & Queen cannot jump
			break
		} else if piece.Type() != King { // This is a capture
			tag := rookLikeCheckTag(b, color, destSq)
			if tag != Check && p.Type() == Queen {
				tag = bishopLikeCheckTag(b, color, destSq)
			}
			allMoves = append(allMoves, Move{srcSq, destSq, piece.Type(), tag | Capture})
			break
		} else if piece.Type() == King {
			// This is a check? what is that? don't add anything yet
			// allMoves = append(allMoves, Move{srcSq, &destSq, piece.Type(), Capture})
			break
		}
	}

	// horizontal up moves
	for destSq := srcSq; destSq >= 0; destSq-- {
		piece := b.PieceAt(destSq)
		if piece == NoPiece {
			tag := rookLikeCheckTag(b, color, destSq)
			if tag != Check && p.Type() == Queen {
				tag = bishopLikeCheckTag(b, color, destSq)
			}
			allMoves = append(allMoves, Move{srcSq, destSq, NoType, tag})
			setCheckSet(checkSet, destSq)
		} else if piece.Color() == color { // Rook & Queen cannot jump
			break
		} else if piece.Type() != King { // This is a capture
			tag := rookLikeCheckTag(b, color, destSq)
			if tag != Check && p.Type() == Queen {
				tag = bishopLikeCheckTag(b, color, destSq)
			}
			allMoves = append(allMoves, Move{srcSq, destSq, piece.Type(), Capture | tag})
			break
		} else if piece.Type() == King {
			// This is a check? what is that? don't add anything yet
			// allMoves = append(allMoves, Move{srcSq, &destSq, piece.Type(), Capture})
			break
		}
	}

	// horizontal down moves
	for destSq := srcSq; destSq < 8; destSq++ {
		piece := b.PieceAt(destSq)
		if piece == NoPiece {
			tag := rookLikeCheckTag(b, color, destSq)
			if tag != Check && p.Type() == Queen {
				tag = bishopLikeCheckTag(b, color, destSq)
			}
			allMoves = append(allMoves, Move{srcSq, destSq, NoType, tag})
			setCheckSet(checkSet, destSq)
		} else if piece.Color() == color { // Rook & Queen cannot jump
			break
		} else if piece.Type() != King { // This is a capture
			tag := rookLikeCheckTag(b, color, destSq)
			if tag != Check && p.Type() == Queen {
				tag = bishopLikeCheckTag(b, color, destSq)
			}
			allMoves = append(allMoves, Move{srcSq, destSq, piece.Type(), Capture | tag})
			break
		} else if piece.Type() == King {
			// This is a check? what is that? don't add anything yet
			// allMoves = append(allMoves, Move{srcSq, &destSq, piece.Type(), Capture})
			break
		}
	}
	return checkSet, allMoves
}

func bishopLikeMoves(b *Bitboard, p Piece, srcSq Square,
	allMoves []Move) (uint64, []Move) {
	color := p.Color()
	var checkSet uint64 = 0

	// up-right
	for destSq := srcSq; destSq >= 0; destSq -= 7 {
		piece := b.PieceAt(destSq)
		if piece == NoPiece {
			tag := bishopLikeCheckTag(b, color, destSq)
			if tag != Check && p.Type() == Queen {
				tag = rookLikeCheckTag(b, color, destSq)
			}
			allMoves = append(allMoves, Move{srcSq, destSq, NoType, tag})
			setCheckSet(checkSet, destSq)
		} else if piece.Color() == color { // Bishop & Queen cannot jump
			break
		} else if piece.Type() != King { // This is a capture
			tag := bishopLikeCheckTag(b, color, destSq)
			if tag != Check && p.Type() == Queen {
				tag = rookLikeCheckTag(b, color, destSq)
			}
			allMoves = append(allMoves, Move{srcSq, destSq, piece.Type(), Capture | tag})
			break
		} else if piece.Type() == King {
			break
			// This is a check? what is that? don't add anything yet
			// allMoves = append(allMoves, Move{srcSq, &destSq, piece.Type(), Capture})
		}
	}

	// down-left
	for destSq := srcSq; destSq < 64; destSq += 7 {
		piece := b.PieceAt(destSq)
		if piece == NoPiece {
			tag := bishopLikeCheckTag(b, color, destSq)
			if tag != Check && p.Type() == Queen {
				tag = rookLikeCheckTag(b, color, destSq)
			}
			allMoves = append(allMoves, Move{srcSq, destSq, NoType, tag})
			setCheckSet(checkSet, destSq)
		} else if piece.Color() == color { // Bishop & Queen cannot jump
			break
		} else if piece.Type() != King { // This is a capture
			tag := bishopLikeCheckTag(b, color, destSq)
			if tag != Check && p.Type() == Queen {
				tag = rookLikeCheckTag(b, color, destSq)
			}
			allMoves = append(allMoves, Move{srcSq, destSq, piece.Type(), tag | Capture})
			break
		} else if piece.Type() == King {
			break
			// This is a check? what is that? don't add anything yet
			// allMoves = append(allMoves, Move{srcSq, &destSq, piece.Type(), Capture})
		}
	}

	// up-left
	for destSq := srcSq; destSq >= 0; destSq -= 9 {
		piece := b.PieceAt(destSq)
		if piece == NoPiece {
			tag := bishopLikeCheckTag(b, color, destSq)
			if tag != Check && p.Type() == Queen {
				tag = rookLikeCheckTag(b, color, destSq)
			}
			allMoves = append(allMoves, Move{srcSq, destSq, NoType, tag})
			setCheckSet(checkSet, destSq)
		} else if piece.Color() == color { // Bishop & Queen cannot jump
			break
		} else if piece.Type() != King { // This is a capture
			tag := bishopLikeCheckTag(b, color, destSq)
			if tag != Check && p.Type() == Queen {
				tag = rookLikeCheckTag(b, color, destSq)
			}
			allMoves = append(allMoves, Move{srcSq, destSq, piece.Type(), tag | Capture})
			break
		} else if piece.Type() == King {
			// This is a check? what is that? don't add anything yet
			// allMoves = append(allMoves, Move{srcSq, &destSq, piece.Type(), Capture})
			break
		}
	}

	// up-right
	for destSq := srcSq; destSq < 8; destSq += 9 {
		piece := b.PieceAt(destSq)
		if piece == NoPiece {
			tag := bishopLikeCheckTag(b, color, destSq)
			if tag != Check && p.Type() == Queen {
				tag = rookLikeCheckTag(b, color, destSq)
			}
			allMoves = append(allMoves, Move{srcSq, destSq, NoType, tag})
			setCheckSet(checkSet, destSq)
		} else if piece.Color() == color { // Bishop & Queen cannot jump
			break
		} else if piece.Type() != King { // This is a capture
			tag := bishopLikeCheckTag(b, color, destSq)
			if tag != Check && p.Type() == Queen {
				tag = rookLikeCheckTag(b, color, destSq)
			}
			allMoves = append(allMoves, Move{srcSq, destSq, piece.Type(), Capture | tag})
			break
		} else if piece.Type() == King {
			// This is a check? what is that? don't add anything yet
			// allMoves = append(allMoves, Move{srcSq, &destSq, piece.Type(), Capture})
			break
		}
	}
	return checkSet, allMoves
}

func knightMoves(b *Bitboard, color Color, srcSq Square,
	allMoves *[]Move) (uint64, []Move) {
	moves := *allMoves
	var checkSet uint64 = 0
	destinations := []Square{srcSq + 15, srcSq + 17, srcSq + 6, srcSq + 10, srcSq - 15, srcSq - 17, srcSq - 6, srcSq - 10}

	for i := 0; i < len(destinations); i++ {
		destSq := destinations[i]
		if destSq >= A1 && destSq <= H8 {
			piece := b.PieceAt(destSq)
			if piece == NoPiece {
				tag := knightCheckTag(b, color, destSq)
				moves = append(moves, Move{srcSq, destSq, NoType, tag})
				setCheckSet(checkSet, destSq)
			} else if piece.Color() == color { // Knight cannot land on own pieces
				continue
			} else if piece.Type() != King { // This is a capture
				tag := knightCheckTag(b, color, destSq)
				moves = append(moves, Move{srcSq, destSq, piece.Type(), tag | Capture})
			}
		}
	}

	return checkSet, moves
}

func knightCheckTag(b *Bitboard, color Color, srcSq Square) MoveTag {
	destinations := []Square{srcSq + 15, srcSq + 17, srcSq + 6, srcSq + 10, srcSq - 15, srcSq - 17, srcSq - 6, srcSq - 10}

	for i := 0; i < len(destinations); i++ {
		destSq := destinations[i]
		if destSq >= 0 && destSq < 64 {
			piece := b.PieceAt(destSq)
			if piece.Color() != color && piece.Type() == King {
				return Check
			}
		}
	}

	return 0
}

func rookLikeCheckTag(b *Bitboard, color Color, src Square) MoveTag {
	// vertical up moves
	for destSq := src; destSq >= 0; destSq -= 8 {
		piece := b.PieceAt(destSq)
		if piece.Color() != color && piece.Type() == King {
			return Check
		}
	}

	// vertical down moves
	for destSq := src; destSq < 64; destSq += 8 {
		piece := b.PieceAt(destSq)
		if piece.Color() != color && piece.Type() == King {
			return Check
		}
	}

	// horizontal up moves
	for destSq := src; destSq >= 0; destSq-- {
		piece := b.PieceAt(destSq)
		if piece.Color() != color && piece.Type() == King {
			return Check
		}
	}

	// horizontal down moves
	for destSq := src; destSq < 8; destSq++ {
		piece := b.PieceAt(destSq)
		if piece.Color() != color && piece.Type() == King {
			return Check
		}
	}
	return 0
}

func bishopLikeCheckTag(b *Bitboard, color Color, src Square) MoveTag {

	// up-right
	for destSq := src; destSq >= 0; destSq -= 7 {
		piece := b.PieceAt(destSq)
		if piece.Color() != color && piece.Type() == King {
			return Check
		}
	}

	// down-left
	for destSq := src; destSq < 64; destSq += 7 {
		piece := b.PieceAt(destSq)
		if piece.Color() != color && piece.Type() == King {
			return Check
		}
	}

	// up-left
	for destSq := src; destSq >= 0; destSq -= 9 {
		piece := b.PieceAt(destSq)
		if piece.Color() != color && piece.Type() == King {
			return Check
		}
	}

	// up-right
	for destSq := src; destSq < 8; destSq += 9 {
		piece := b.PieceAt(destSq)
		if piece.Color() != color && piece.Type() == King {
			return Check
		}
	}
	return 0
}

func setCheckSet(checkBitSet uint64, bitIndex Square) {
	checkBitSet |= (1 << bitIndex)
}

func findEnPassantSquare(move Move, movingPiece Piece) Square {
	if !move.HasTag(EnPassant) {
		return NoSquare
	} else if movingPiece == WhitePawn &&
		move.source.Rank() == Rank2 && move.destination.Rank() == Rank4 {
		return SquareOf(move.source.File(), Rank3)
	} else if movingPiece == BlackPawn &&
		move.source.Rank() == Rank7 && move.destination.Rank() == Rank5 {
		return SquareOf(move.source.File(), Rank6)
	}
	return NoSquare
}
