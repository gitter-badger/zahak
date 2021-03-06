package engine

import (
	"fmt"
	"testing"
)

func TestBishopMoves(t *testing.T) {
	fen := "rnbqkbnr/pPp1pppp/4P3/3pP3/3p4/4B1N1/PP2BPPP/1NRQK2R w Kkq - 0 1"
	g := FromFen(fen, true)
	board := g.position.Board
	moves := make([]Move, 0, 8)
	g.position.bbSlidingMoves(board.whiteBishop, board.whitePieces, board.blackPieces,
		White, bishopAttacks, false, false, false, false, &moves)
	expectedMoves := []Move{
		Move{E2, F1, NoType, 0},
		Move{E2, F3, NoType, 0},
		Move{E2, G4, NoType, 0},
		Move{E2, H5, NoType, 0},
		Move{E2, D3, NoType, 0},
		Move{E2, C4, NoType, 0},
		Move{E2, B5, NoType, Check},
		Move{E2, A6, NoType, 0},
		Move{E3, D2, NoType, 0},
		Move{E3, F4, NoType, 0},
		Move{E3, G5, NoType, 0},
		Move{E3, H6, NoType, 0},
		Move{E3, D4, NoType, Capture},
	}
	expectedLen := len(expectedMoves)
	if len(moves) != expectedLen || !equalMoves(expectedMoves, moves) {
		fmt.Println("Got:")
		for _, i := range moves {
			fmt.Println(i.ToString(), i.PromoType, i.Tag)
		}
		fmt.Println("Expected:")
		for _, i := range expectedMoves {
			fmt.Println(i.ToString(), i.PromoType, i.Tag)
		}
		t.Errorf("Expected different number of moves to be generated%s",
			fmt.Sprintf("\nExpected: %d\nGot: %d\n", expectedLen, len(moves)))
	}
}

func TestRookMoves(t *testing.T) {
	fen := "rnkqbbnr/ppp1pppp/4P3/3pP3/3P4/4B1N1/PP2BPPP/1NRQK2R w Kkq - 0 1"
	g := FromFen(fen, true)
	board := g.position.Board
	moves := make([]Move, 0, 8)
	g.position.bbSlidingMoves(board.whiteRook, board.whitePieces, board.blackPieces,
		White, rookAttacks, false, false, false, false, &moves)
	expectedMoves := []Move{
		Move{H1, G1, NoType, 0},
		Move{H1, F1, NoType, 0},
		Move{C1, C2, NoType, 0},
		Move{C1, C3, NoType, 0},
		Move{C1, C4, NoType, 0},
		Move{C1, C5, NoType, 0},
		Move{C1, C6, NoType, 0},
		Move{C1, C7, NoType, Capture | Check},
	}
	expectedLen := len(expectedMoves)
	if len(moves) != expectedLen || !equalMoves(expectedMoves, moves) {
		fmt.Println("Got:")
		for _, i := range moves {
			fmt.Println(i.ToString(), i.PromoType, i.Tag)
		}
		fmt.Println("Expected:")
		for _, i := range expectedMoves {
			fmt.Println(i.ToString(), i.PromoType, i.Tag)
		}
		t.Errorf("Expected different number of moves to be generated%s",
			fmt.Sprintf("\nExpected: %d\nGot: %d\n", expectedLen, len(moves)))
	}
}

func TestQueenMoves(t *testing.T) {
	fen := "rnbqkbnr/pPp1pppp/4P3/3pP3/3p4/4B1N1/PP2BPPP/1NRQK2R w Kkq - 0 1"
	g := FromFen(fen, true)
	board := g.position.Board
	moves := make([]Move, 0, 8)
	g.position.bbSlidingMoves(board.whiteQueen, board.whitePieces, board.blackPieces,
		White, queenAttacks, false, false, false, false, &moves)
	expectedMoves := []Move{
		Move{D1, D2, NoType, 0},
		Move{D1, D3, NoType, 0},
		Move{D1, D4, NoType, Capture},
		Move{D1, C2, NoType, 0},
		Move{D1, B3, NoType, 0},
		Move{D1, A4, NoType, Check},
	}
	expectedLen := len(expectedMoves)
	if len(moves) != expectedLen || !equalMoves(expectedMoves, moves) {
		fmt.Println("Got:")
		for _, i := range moves {
			fmt.Println(i.ToString(), i.PromoType, i.Tag)
		}
		fmt.Println("Expected:")
		for _, i := range expectedMoves {
			fmt.Println(i.ToString(), i.PromoType, i.Tag)
		}
		t.Errorf("Expected different number of moves to be generated%s",
			fmt.Sprintf("\nExpected: %d\nGot: %d\n", expectedLen, len(moves)))
	}
}

func TestKingMoves(t *testing.T) {
	fen := "rnbqkbn1/pPp1pppp/4P3/3pP3/3p4/4B1N1/PP1rBPPP/R3K2R w Kkq - 0 1"
	g := FromFen(fen, true)
	p := g.position
	board := g.position.Board
	moves := make([]Move, 0, 8)
	color := White
	taboo := tabooSquares(board, color)
	g.position.bbKingMoves(board.whiteKing, board.whitePieces, board.blackPieces, board.blackKing,
		taboo, color, p.HasTag(WhiteCanCastleKingSide), p.HasTag(WhiteCanCastleQueenSide), false, false, false, false, &moves)
	expectedMoves := []Move{
		Move{E1, D2, NoType, Capture},
		Move{E1, F1, NoType, 0},
		Move{E1, G1, NoType, KingSideCastle},
	}
	expectedLen := len(expectedMoves)
	if len(moves) != expectedLen || !equalMoves(expectedMoves, moves) {
		fmt.Println("Got:")
		for _, i := range moves {
			fmt.Println(i.ToString(), i.PromoType, i.Tag)
		}
		fmt.Println("Expected:")
		for _, i := range expectedMoves {
			fmt.Println(i.ToString(), i.PromoType, i.Tag)
		}
		t.Errorf("Expected different number of moves to be generated%s",
			fmt.Sprintf("\nExpected: %d\nGot: %d\n", expectedLen, len(moves)))
	}
}

func TestKingCastlingWithOccupiedSquares(t *testing.T) {
	fen := "rnbqkbnr/1p6/p1p3Pp/1B1pp2Q/1P6/B7/P1PP1PPP/RN2K1NR w KQkq - 0 1"
	g := FromFen(fen, true)
	p := g.position
	board := g.position.Board
	moves := make([]Move, 0, 8)
	color := White
	taboo := tabooSquares(board, color)
	g.position.bbKingMoves(board.whiteKing, board.whitePieces, board.blackPieces, board.blackKing,
		taboo, color, p.HasTag(WhiteCanCastleKingSide), p.HasTag(WhiteCanCastleQueenSide), false, false, false, false, &moves)
	expectedMoves := []Move{
		Move{E1, E2, NoType, 0},
		Move{E1, F1, NoType, 0},
		Move{E1, D1, NoType, 0},
	}
	expectedLen := len(expectedMoves)
	if len(moves) != expectedLen || !equalMoves(expectedMoves, moves) {
		fmt.Println("Got:")
		for _, i := range moves {
			fmt.Println(i.ToString(), i.PromoType, i.Tag)
		}
		fmt.Println("Expected:")
		for _, i := range expectedMoves {
			fmt.Println(i.ToString(), i.PromoType, i.Tag)
		}
		t.Errorf("Expected different number of moves to be generated%s",
			fmt.Sprintf("\nExpected: %d\nGot: %d\n", expectedLen, len(moves)))
	}
}

func TestKingQueenSideCastling(t *testing.T) {
	fen := "rnbqkbnr/1p6/p1p3Pp/1B1pp2Q/1P6/B7/P1PP1PPP/R3K1NR w KQkq - 0 1"
	g := FromFen(fen, true)
	p := g.position
	board := g.position.Board
	moves := make([]Move, 0, 8)
	color := White
	taboo := tabooSquares(board, color)
	g.position.bbKingMoves(board.whiteKing, board.whitePieces, board.blackPieces, board.blackKing,
		taboo, color, p.HasTag(WhiteCanCastleKingSide), p.HasTag(WhiteCanCastleQueenSide), false, false, false, false, &moves)
	expectedMoves := []Move{
		Move{E1, E2, NoType, 0},
		Move{E1, F1, NoType, 0},
		Move{E1, D1, NoType, 0},
		Move{E1, C1, NoType, QueenSideCastle},
	}
	expectedLen := len(expectedMoves)
	if len(moves) != expectedLen || !equalMoves(expectedMoves, moves) {
		fmt.Println("Got:")
		for _, i := range moves {
			fmt.Println(i.ToString(), i.PromoType, i.Tag)
		}
		fmt.Println("Expected:")
		for _, i := range expectedMoves {
			fmt.Println(i.ToString(), i.PromoType, i.Tag)
		}
		t.Errorf("Expected different number of moves to be generated%s",
			fmt.Sprintf("\nExpected: %d\nGot: %d\n", expectedLen, len(moves)))
	}
}

func TestPawnMovesForWhite(t *testing.T) {
	fen := "rnbqkbn1/pPp1pppp/4P3/3pP3/3p4/4B1N1/PP1rBPPP/R3K2R w Kkq d6 0 1"
	g := FromFen(fen, true)
	p := g.position
	board := g.position.Board
	moves := make([]Move, 0, 8)
	color := White
	g.position.bbPawnMoves(board.whitePawn, board.whitePieces, board.blackPieces,
		color, p.EnPassant, false, false, false, false, &moves)
	expectedMoves := []Move{
		Move{H2, H4, NoType, 0},
		Move{H2, H3, NoType, 0},
		Move{F2, F4, NoType, 0},
		Move{F2, F3, NoType, 0},
		Move{A2, A4, NoType, 0},
		Move{A2, A3, NoType, 0},
		Move{B2, B4, NoType, 0},
		Move{B2, B3, NoType, 0},
		Move{E5, D6, NoType, EnPassant | Capture},
		Move{E6, F7, NoType, Capture | Check},
		Move{B7, A8, Queen, Capture},
		Move{B7, A8, Rook, Capture},
		Move{B7, A8, Bishop, Capture},
		Move{B7, A8, Knight, Capture},
		Move{B7, C8, Queen, Capture},
		Move{B7, C8, Rook, Capture},
		Move{B7, C8, Bishop, Capture},
		Move{B7, C8, Knight, Capture},
	}
	expectedLen := len(expectedMoves)
	if len(moves) != expectedLen || !equalMoves(expectedMoves, moves) {
		fmt.Println("Got:")
		for _, i := range moves {
			fmt.Println(i.ToString(), i.PromoType, i.Tag)
		}
		fmt.Println("Expected:")
		for _, i := range expectedMoves {
			fmt.Println(i.ToString(), i.PromoType, i.Tag)
		}
		t.Errorf("Expected different number of moves to be generated%s",
			fmt.Sprintf("\nExpected: %d\nGot: %d\n", expectedLen, len(moves)))
	}
}

func TestPawnMovesForBlack(t *testing.T) {
	fen := "rnbqkbnr/ppp3pp/3p1p2/1P4P1/4pP2/N6N/P1PPP2P/R1BQKB1R b KQkq f3 0 1"
	g := FromFen(fen, true)
	p := g.position
	board := g.position.Board
	moves := make([]Move, 0, 8)
	color := Black
	g.position.bbPawnMoves(board.blackPawn, board.blackPieces, board.whitePieces,
		color, p.EnPassant, false, false, false, false, &moves)
	expectedMoves := []Move{
		Move{H7, H6, NoType, 0},
		Move{H7, H5, NoType, 0},
		Move{G7, G6, NoType, 0},
		Move{F6, F5, NoType, 0},
		Move{F6, G5, NoType, Capture},
		Move{E4, E3, NoType, 0},
		Move{E4, F3, NoType, EnPassant | Capture},
		Move{D6, D5, NoType, 0},
		Move{C7, C6, NoType, 0},
		Move{C7, C5, NoType, 0},
		Move{B7, B6, NoType, 0},
		Move{A7, A6, NoType, 0},
		Move{A7, A5, NoType, 0},
	}
	expectedLen := len(expectedMoves)
	if len(moves) != expectedLen || !equalMoves(expectedMoves, moves) {
		fmt.Println("Got:")
		for _, i := range moves {
			fmt.Println(i.ToString(), i.PromoType, i.Tag)
		}
		fmt.Println("Expected:")
		for _, i := range expectedMoves {
			fmt.Println(i.ToString(), i.PromoType, i.Tag)
		}
		t.Errorf("Expected different number of moves to be generated%s",
			fmt.Sprintf("\nExpected: %d\nGot: %d\n", expectedLen, len(moves)))
	}
}

func TestKnightMoves(t *testing.T) {
	fen := "rnbqkbn1/pPp1pppp/4P3/1N1pP3/3p4/4B1N1/PP1rBPPP/R3K2R w Kkq d6 0 1"
	g := FromFen(fen, true)
	p := g.position
	b := p.Board
	moves := make([]Move, 0, 8)
	g.position.bbKnightMoves(b.whiteKnight, b.whitePieces, b.blackPieces, false, false, false, false, &moves)
	expectedMoves := []Move{
		Move{G3, F1, NoType, 0},
		Move{G3, E4, NoType, 0},
		Move{G3, F5, NoType, 0},
		Move{G3, H5, NoType, 0},
		Move{B5, A7, NoType, Capture},
		Move{B5, A3, NoType, 0},
		Move{B5, C7, NoType, Capture | Check},
		Move{B5, C3, NoType, 0},
		Move{B5, D4, NoType, Capture},
		Move{B5, D6, NoType, Check},
	}
	expectedLen := len(expectedMoves)
	if len(moves) != expectedLen || !equalMoves(expectedMoves, moves) {
		fmt.Println("Got:")
		for _, i := range moves {
			fmt.Println(i.ToString(), i.PromoType, i.Tag)
		}
		fmt.Println("Expected:")
		for _, i := range expectedMoves {
			fmt.Println(i.ToString(), i.PromoType, i.Tag)
		}
		t.Errorf("Expected different number of moves to be generated%s",
			fmt.Sprintf("\nExpected: %d\nGot: %d\n", expectedLen, len(moves)))
	}
}

func TestCastleAndDiscoveredChecks(t *testing.T) {
	fen := "rnbq1bn1/pPp1pppp/4P3/3pP3/3p4/4B1N1/PP1rBPPP/k3K2R w Kkq - 0 1"
	g := FromFen(fen, true)
	p := g.position
	legalMoves := p.LegalMoves()
	move := Move{E1, G1, NoType, Check | KingSideCastle}
	if !containsMove(legalMoves, move) {
		fmt.Println("Got:")
		for _, i := range legalMoves {
			fmt.Println(i.ToString(), i.PromoType, i.Tag)
		}
		t.Errorf("Expected to see %s", fmt.Sprintf("%s %d", move.ToString(), move.Tag))
	}
	move = Move{E1, D2, NoType, Check | Capture}
	if !containsMove(legalMoves, move) {
		fmt.Println("Got:")
		for _, i := range legalMoves {
			fmt.Println(i.ToString(), i.PromoType, i.Tag)
		}
		t.Errorf("Expected to see %s", fmt.Sprintf("%s %d", move.ToString(), move.Tag))
	}
}

func TestCastleAndPawnAttack(t *testing.T) {
	fen := "r3k2r/p1ppqpb1/1n2pnp1/1b1PN3/1p2P3/P1N2Q2/1PPBBPpP/1R2K2R w Kkq - 0 1"
	g := FromFen(fen, true)
	p := g.position
	board := g.position.Board
	moves := make([]Move, 0, 8)
	color := White
	taboo := tabooSquares(board, color)
	g.position.bbKingMoves(board.whiteKing, board.whitePieces, board.blackPieces, board.blackKing,
		taboo, color, p.HasTag(WhiteCanCastleKingSide), p.HasTag(WhiteCanCastleQueenSide), false, false, false, false, &moves)
	expectedMoves := []Move{
		Move{E1, D1, NoType, 0},
	}
	expectedLen := len(expectedMoves)
	if !equalMoves(expectedMoves, moves) {
		fmt.Println(g.position.Board.Draw())
		fmt.Println("Got:")
		for _, i := range moves {
			fmt.Println(i.ToString(), i.PromoType, i.Tag)
		}
		fmt.Println("Expected:")
		for _, i := range expectedMoves {
			fmt.Println(i.ToString(), i.PromoType, i.Tag)
		}
		t.Errorf("Expected different number of moves to be generated%s",
			fmt.Sprintf("\nExpected: %d\nGot: %d\n", expectedLen, len(moves)))
	}
}

func TestLegalMoves(t *testing.T) {
	fen := "rn1q1bn1/pPp1pppp/4P3/1N1pP2Q/3p3b/4B3/PP1rBPPP/k3K2R w Kkq d6 0 1"
	g := FromFen(fen, true)
	p := g.position
	legalMoves := p.LegalMoves()
	expectedMoves := []Move{
		Move{H1, G1, NoType, 0},
		Move{H1, F1, NoType, 0},
		Move{E1, F1, NoType, 0},
		Move{E1, G1, NoType, Check | KingSideCastle},
		Move{E1, D2, NoType, Check | Capture},
		Move{H2, H3, NoType, 0},
		Move{G2, G3, NoType, 0},
		Move{G2, G4, NoType, 0},
		Move{E2, F1, NoType, 0},
		Move{E2, D1, NoType, 0},
		Move{E2, F3, NoType, 0},
		Move{E2, G4, NoType, 0},
		Move{E2, D3, NoType, 0},
		Move{E2, C4, NoType, 0},
		Move{B2, B3, NoType, 0},
		Move{B2, B4, NoType, 0},
		Move{A2, A3, NoType, 0},
		Move{A2, A4, NoType, 0},
		Move{E3, D4, NoType, Capture},
		Move{E3, F4, NoType, 0},
		Move{E3, G5, NoType, 0},
		Move{E3, H6, NoType, 0},
		Move{H5, H7, NoType, Capture},
		Move{H5, H6, NoType, 0},
		Move{H5, F7, NoType, Capture},
		Move{H5, G6, NoType, 0},
		Move{H5, G5, NoType, 0},
		Move{H5, F5, NoType, 0},
		Move{H5, G4, NoType, 0},
		Move{H5, F3, NoType, 0},
		Move{H5, H4, NoType, Capture},
		Move{E5, D6, NoType, Capture | EnPassant},
		Move{B5, A3, NoType, 0},
		Move{B5, C3, NoType, 0},
		Move{B5, A7, NoType, Capture},
		Move{B5, C7, NoType, Capture},
		Move{B5, D4, NoType, Capture},
		Move{B5, D6, NoType, 0},
		Move{B5, D6, NoType, 0},
		Move{E6, F7, NoType, Capture},
		Move{B7, A8, Queen, Capture},
		Move{B7, A8, Rook, Capture},
		Move{B7, A8, Bishop, Capture},
		Move{B7, A8, Knight, Capture},
	}
	expectedLen := len(expectedMoves)
	if expectedLen != len(legalMoves) || !equalMoves(expectedMoves, legalMoves) {
		fmt.Println("Got:")
		for _, i := range legalMoves {
			fmt.Println(i.ToString(), i.PromoType, i.Tag)
		}
		fmt.Println("Expected:")
		for _, i := range expectedMoves {
			fmt.Println(i.ToString(), i.PromoType, i.Tag)
		}
		t.Errorf("Expected different number of moves to be generated%s",
			fmt.Sprintf("\nExpected: %d\nGot: %d\n", expectedLen, len(legalMoves)))
	}
}

func TestDoubleCheckResponses(t *testing.T) {
	fen := "5Q2/8/1q5P/8/6k1/5R2/6P1/2r3K1 w - - 0 1"
	g := FromFen(fen, true)
	p := g.position
	legalMoves := p.LegalMoves()
	expectedMoves := []Move{
		Move{G1, H2, NoType, 0},
	}
	if !p.IsInCheck() {
		t.Errorf("Position is wrongfully considered not check for: %s", fen)
	}
	if !isDoubleCheck(p.Board, White) {
		t.Errorf("Position is wrongfully considered not double-check for: %s", fen)
	}
	if p.Status() != Unknown {
		t.Errorf("Position is wrongfully considered ended: %b", p.Status())
	}
	expectedLen := len(expectedMoves)
	if expectedLen != len(legalMoves) || !equalMoves(expectedMoves, legalMoves) {
		fmt.Println("Got:")
		for _, i := range legalMoves {
			fmt.Println(i.ToString(), i.PromoType, i.Tag)
		}
		fmt.Println("Expected:")
		for _, i := range expectedMoves {
			fmt.Println(i.ToString(), i.PromoType, i.Tag)
		}
		t.Errorf("Expected different number of moves to be generated%s",
			fmt.Sprintf("\nExpected: %d\nGot: %d\n", expectedLen, len(legalMoves)))
	}
}

func TestHasLegalMovesCheckmate(t *testing.T) {
	fen := "5Q2/8/1q5P/8/6k1/5R2/6PR/2r3K1 w - - 0 1"
	g := FromFen(fen, true)
	p := g.position
	hasMoves := p.HasLegalMoves()
	if hasMoves {
		for _, i := range p.LegalMoves() {
			fmt.Println(i.ToString(), i.PromoType, i.Tag)
		}
		t.Errorf("Position is wrongfully considered playable, %p", p.LegalMoves())
	}
}

func TestHasLegalMovesDraw(t *testing.T) {
	fen := "7k/5Q2/6K1/8/8/8/8/8 b - - 0 1"
	g := FromFen(fen, true)
	p := g.position
	hasMoves := p.HasLegalMoves()
	if hasMoves {
		t.Errorf("Position is wrongfully considered playable")
	}
}

func TestHasLegalMoves(t *testing.T) {
	fen := "5Q2/8/1q5P/8/6k1/5R2/6P1/2r3K1 w - - 0 1"
	g := FromFen(fen, true)
	p := g.position
	legalMoves1 := p.LegalMoves()
	hasMoves := p.HasLegalMoves()
	legalMoves2 := p.LegalMoves()
	if !hasMoves || !equalMoves(legalMoves1, legalMoves2) {
		fmt.Println("First call to LegalMoves")
		for _, i := range legalMoves1 {
			fmt.Println(i.ToString(), i.PromoType, i.Tag)
		}
		fmt.Println("Second call to LegalMoves")
		for _, i := range legalMoves2 {
			fmt.Println(i.ToString(), i.PromoType, i.Tag)
		}
		t.Errorf("Position is wrongfully considered lost, %p", p.LegalMoves())
	}
}

func TestLegalMovesInOpenning(t *testing.T) {
	fen := "rnbqkbnr/ppp3pp/3ppp2/1P6/6P1/N6N/P1PPPP1P/R1BQKB1R w KQkq - 0 1"
	g := FromFen(fen, true)
	p := g.position
	legalMoves := p.LegalMoves()
	expectedMoves := []Move{
		Move{H1, G1, NoType, 0},
		Move{G4, G5, NoType, 0},
		Move{F2, F3, NoType, 0},
		Move{F2, F4, NoType, 0},
		Move{E2, E3, NoType, 0},
		Move{E2, E4, NoType, 0},
		Move{D2, D3, NoType, 0},
		Move{D2, D4, NoType, 0},
		Move{C2, C3, NoType, 0},
		Move{C2, C4, NoType, 0},
		Move{B5, B6, NoType, 0},
		Move{A1, B1, NoType, 0},
		Move{A3, C4, NoType, 0},
		Move{A3, B1, NoType, 0},
		Move{C1, B2, NoType, 0},
		Move{F1, G2, NoType, 0},
		Move{H3, G5, NoType, 0},
		Move{H3, F4, NoType, 0},
		Move{H3, G1, NoType, 0},
	}
	expectedLen := len(expectedMoves)
	if expectedLen != len(legalMoves) || !equalMoves(expectedMoves, legalMoves) {
		fmt.Println("Got:")
		for _, i := range legalMoves {
			fmt.Println(i.ToString(), i.PromoType, i.Tag)
		}
		fmt.Println("Expected:")
		for _, i := range expectedMoves {
			fmt.Println(i.ToString(), i.PromoType, i.Tag)
		}
		t.Errorf("Expected different number of moves to be generated%s",
			fmt.Sprintf("\nExpected: %d\nGot: %d\n", expectedLen, len(legalMoves)))
	}
}

func equalMoves(moves1 []Move, moves2 []Move) bool {
	if len(moves1) != len(moves2) {
		return false
	}
	for _, m1 := range moves1 {
		exists := false
		for _, m2 := range moves2 {
			if m1 == m2 {
				exists = true
				break
			}
		}
		if !exists {
			fmt.Println("Missing", m1.ToString(), m1.Tag)
			return false
		}
	}
	return true
}

func containsMove(moves1 []Move, move Move) bool {
	exists := false
	for _, m := range moves1 {
		if m == move {
			exists = true
			break
		}
	}
	return exists
}
