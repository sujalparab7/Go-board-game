package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"reflect"
	"strconv"
	"time"
)

const (
	EMPTY      = 0
	BLACK      = 1
	WHITE      = 2
	BOARD_SIZE = 9
	MAX_DEPTH  = 2     
	KOMI       = 6.5 
)



type MoveRequest struct {
	Board          []int          `json:"Board"`
	Player         int            `json:"Player"`
	Row            int            `json:"Row"`
	Col            int            `json:"Col"`
	Captures       map[string]int `json:"Captures"`
	LastBoardState []int          `json:"LastBoardState"`
}

type MoveResponse struct {
	Board    []int          `json:"Board"`
	Captures map[string]int `json:"Captures"`
	Message  string         `json:"Message"`
	MoveType string         `json:"MoveType"` 
}

type ScoreRequest struct {
	Board    []int          `json:"Board"`
	Captures map[string]int `json:"Captures"`
}

type ScoreResponse struct {
	BlackScore float64 `json:"BlackScore"`
	WhiteScore float64 `json:"WhiteScore"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}


type Move struct {
	Row int
	Col int
}


func toIndex(r, c int) int {
	return r*BOARD_SIZE + c
}

func toCoords(idx int) (int, int) {
	return idx / BOARD_SIZE, idx % BOARD_SIZE
}

func isValid(r, c int) bool {
	return r >= 0 && r < BOARD_SIZE && c >= 0 && c < BOARD_SIZE
}

func getOpponent(player int) int {
	if player == BLACK {
		return WHITE
	}
	return BLACK
}


func getGroup(board []int, r, c int) (map[int]bool, int) {
	player := board[toIndex(r, c)]
	if player == EMPTY {
		return make(map[int]bool), 0
	}

	group := make(map[int]bool)     
	libertiesSet := make(map[int]bool) 
	liberties := 0
	
	queue := []int{toIndex(r, c)}
	group[toIndex(r, c)] = true

	for len(queue) > 0 {
		currentIdx := queue[0]
		queue = queue[1:]

		cr, cc := toCoords(currentIdx)
		neighbors := [][]int{{cr - 1, cc}, {cr + 1, cc}, {cr, cc - 1}, {cr, cc + 1}}

		for _, n := range neighbors {
			nr, nc := n[0], n[1]
			if !isValid(nr, nc) {
				continue
			}
			nIdx := toIndex(nr, nc)

			if board[nIdx] == player && !group[nIdx] {
				
				group[nIdx] = true
				queue = append(queue, nIdx)
			} else if board[nIdx] == EMPTY && !libertiesSet[nIdx] {
				
				libertiesSet[nIdx] = true
				liberties++
			}
		}
	}
	return group, liberties
}

func removeCapturedStones(board *[]int, player int) int {
	captured := 0
	opponent := getOpponent(player)
	checked := make(map[int]bool) 

	for r := 0; r < BOARD_SIZE; r++ {
		for c := 0; c < BOARD_SIZE; c++ {
			idx := toIndex(r, c)
			
			if (*board)[idx] == opponent && !checked[idx] {
				group, liberties := getGroup(*board, r, c)

				for gIdx := range group {
					checked[gIdx] = true
				}

				if liberties == 0 {
					
					for gIdx := range group {
						(*board)[gIdx] = EMPTY
						captured++
					}
				}
			}
		}
	}
	return captured
}


func isMoveLegal(board []int, r, c int, player int, lastBoardState []int) bool {
	if r == -1 && c == -1 { 
		return true
	}
	if !isValid(r, c) || board[toIndex(r, c)] != EMPTY {
		return false
	}

	simBoard := make([]int, len(board))
	copy(simBoard, board)
	simBoard[toIndex(r, c)] = player

	removeCapturedStones(&simBoard, player)

	_, newLiberties := getGroup(simBoard, r, c)
	if newLiberties == 0 {
		return false // Suicide
	}

	if reflect.DeepEqual(simBoard, lastBoardState) {
		return false // Ko
	}

	return true
}

func generateMoves(board []int, player int, lastBoardState []int) []Move {
	moves := []Move{}
	candidateSquares := make(map[int]bool)
	hasStones := false

	for r := 0; r < BOARD_SIZE; r++ {
		for c := 0; c < BOARD_SIZE; c++ {
			if board[toIndex(r, c)] != EMPTY {
				hasStones = true
				neighbors := [][]int{{r - 1, c}, {r + 1, c}, {r, c - 1}, {r, c + 1}}
				for _, n := range neighbors {
					nr, nc := n[0], n[1]
					if isValid(nr, nc) && board[toIndex(nr, nc)] == EMPTY {
						candidateSquares[toIndex(nr, nc)] = true
					}
				}
			}
		}
	}

	if !hasStones {
		for r := 0; r < BOARD_SIZE; r++ {
			for c := 0; c < BOARD_SIZE; c++ {
				moves = append(moves, Move{Row: r, Col: c})
			}
		}
	} else {
		for idx := range candidateSquares {
			r, c := toCoords(idx)
			if isMoveLegal(board, r, c, player, lastBoardState) {
				moves = append(moves, Move{Row: r, Col: c})
			}
		}
	}

	moves = append(moves, Move{Row: -1, Col: -1})
	return moves
}

func evaluate(board []int, captures map[string]int) float64 {
	blackScore := 0.0 
	whiteScore := 0.0 

	blackTerritory, whiteTerritory := countTerritory(board)
	blackScore += float64(blackTerritory) 
	whiteScore += float64(whiteTerritory) 

	blackScore += float64(captures[strconv.Itoa(WHITE)]) 
	whiteScore += float64(captures[strconv.Itoa(BLACK)]) 

	whiteScore += KOMI 

	return whiteScore - blackScore
}

func countTerritory(board []int) (int, int) {
	blackTerritory := 0
	whiteTerritory := 0
	visited := make(map[int]bool)

	for r := 0; r < BOARD_SIZE; r++ {
		for c := 0; c < BOARD_SIZE; c++ {
			idx := toIndex(r, c)
			if board[idx] == EMPTY && !visited[idx] {
				territory := 0
				queue := []int{idx}
				visited[idx] = true
				
				touchesBlack := false
				touchesWhite := false
				
				currentTerritoryGroup := []int{}

				for len(queue) > 0 {
					currentIdx := queue[0]
					queue = queue[1:]
					territory++
					currentTerritoryGroup = append(currentTerritoryGroup, currentIdx)

					cr, cc := toCoords(currentIdx)
					neighbors := [][]int{{cr - 1, cc}, {cr + 1, cc}, {cr, cc - 1}, {cr, cc + 1}}

					for _, n := range neighbors {
						nr, nc := n[0], n[1]
						if !isValid(nr, nc) {
							continue
						}
						nIdx := toIndex(nr, nc)
						if board[nIdx] == EMPTY && !visited[nIdx] {
							visited[nIdx] = true
							queue = append(queue, nIdx)
						} else if board[nIdx] == BLACK {
							touchesBlack = true
						} else if board[nIdx] == WHITE {
							touchesWhite = true
						}
					}
				}
				
				if touchesBlack && !touchesWhite {
					blackTerritory += territory
				} else if !touchesBlack && touchesWhite {
					whiteTerritory += territory
				}
			}
		}
	}
	return blackTerritory, whiteTerritory
}


func alphaBeta(board []int, depth int, alpha, beta float64, maximizingPlayer bool, captures map[string]int, lastBoardState []int) float64 {
	if depth == 0 {
		return evaluate(board, captures)
	}

	player := WHITE 
	if !maximizingPlayer {
		player = BLACK 
	}

	moves := generateMoves(board, player, lastBoardState)

	if maximizingPlayer {
		maxEval := -math.MaxFloat64 
		for _, move := range moves {
			simBoard := make([]int, len(board))
			copy(simBoard, board)
			simCaptures := copyCaptures(captures)

			if move.Row != -1 {
				simBoard[toIndex(move.Row, move.Col)] = player
				captured := removeCapturedStones(&simBoard, player)
				simCaptures[strconv.Itoa(getOpponent(player))] += captured
			}

			eval := alphaBeta(simBoard, depth-1, alpha, beta, false, simCaptures, board)
			maxEval = math.Max(maxEval, eval) 
			alpha = math.Max(alpha, eval)     
			if beta <= alpha {
				break
			}
		}
		return maxEval
	} else {
		minEval := math.MaxFloat64 
		for _, move := range moves {
			simBoard := make([]int, len(board))
			copy(simBoard, board)
			simCaptures := copyCaptures(captures)

			if move.Row != -1 { 
				simBoard[toIndex(move.Row, move.Col)] = player
				captured := removeCapturedStones(&simBoard, player)
				simCaptures[strconv.Itoa(getOpponent(player))] += captured
			}
			
			eval := alphaBeta(simBoard, depth-1, alpha, beta, true, simCaptures, board)
			minEval = math.Min(minEval, eval) 
			beta = math.Min(beta, eval)       
			if beta <= alpha {
				break
			}
		}
		return minEval
	}
}

func findBestMove(board []int, captures map[string]int, lastBoardState []int) Move {
	bestScore := -math.MaxFloat64 // FIX: Use float64 min
	bestMove := Move{Row: -1, Col: -1} // Default to pass

	moves := generateMoves(board, WHITE, lastBoardState)

	for _, move := range moves {
		simBoard := make([]int, len(board))
		copy(simBoard, board)
		simCaptures := copyCaptures(captures)

		if move.Row != -1 {
			simBoard[toIndex(move.Row, move.Col)] = WHITE
			captured := removeCapturedStones(&simBoard, WHITE)
			simCaptures[strconv.Itoa(BLACK)] += captured
		}

		moveScore := alphaBeta(simBoard, MAX_DEPTH-1, bestScore, math.MaxFloat64, false, simCaptures, board)

		if moveScore > bestScore {
			bestScore = moveScore
			bestMove = move
		}
	}
	return bestMove
}

func copyCaptures(captures map[string]int) map[string]int {
	newCaptures := make(map[string]int)
	for k, v := range captures {
		newCaptures[k] = v
	}
	return newCaptures
}


func enableCORS(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func moveHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var req MoveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Player == BLACK {
		if req.Row == -1 && req.Col == -1 {
			resp := MoveResponse{
				Board:    req.Board,
				Captures: req.Captures,
				Message:  "You passed. AI's turn.",
				MoveType: "PASS",
			}
			json.NewEncoder(w).Encode(resp)
			return
		}

		if !isMoveLegal(req.Board, req.Row, req.Col, req.Player, req.LastBoardState) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Illegal move. (Suicide or Ko)"})
			return
		}

		req.Board[toIndex(req.Row, req.Col)] = req.Player
		captured := removeCapturedStones(&req.Board, req.Player)
		req.Captures[strconv.Itoa(WHITE)] += captured 
	}

	time.Sleep(500 * time.Millisecond) 
	
	aiMove := findBestMove(req.Board, req.Captures, req.Board) 
	
	aiMoveType := "PLAY"
	aiMessage := fmt.Sprintf("AI played at (%d, %d).", aiMove.Row, aiMove.Col)

	if aiMove.Row == -1 && aiMove.Col == -1 {
		aiMoveType = "PASS"
		aiMessage = "AI passes. Your turn."
	} else {
		if !isMoveLegal(req.Board, aiMove.Row, aiMove.Col, WHITE, req.Board) {
			aiMoveType = "PASS"
			aiMessage = "AI tried illegal move and passes."
		} else {
			req.Board[toIndex(aiMove.Row, aiMove.Col)] = WHITE
			captured := removeCapturedStones(&req.Board, WHITE)
			req.Captures[strconv.Itoa(BLACK)] += captured 
		}
	}

	resp := MoveResponse{
		Board:    req.Board,
		Captures: req.Captures,
		Message:  aiMessage,
		MoveType: aiMoveType,
	}
	json.NewEncoder(w).Encode(resp)
}

func scoreHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var req ScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	blackTerritory, whiteTerritory := countTerritory(req.Board)
	blackCaptures := req.Captures[strconv.Itoa(WHITE)]
	whiteCaptures := req.Captures[strconv.Itoa(BLACK)]

	resp := ScoreResponse{
		BlackScore: float64(blackTerritory + blackCaptures),
		WhiteScore: float64(whiteTerritory+whiteCaptures) + KOMI,
	}
	json.NewEncoder(w).Encode(resp)
}


func main() {
	http.HandleFunc("/move", moveHandler)
	http.HandleFunc("/score", scoreHandler)

	fmt.Println("Go AI Server starting on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

