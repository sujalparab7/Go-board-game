Go AI Game (9x9)

This project is a web-based 9x9 Go game where you (playing as Black) compete against an AI opponent (playing as White).

The frontend is a clean, single-page application built with HTML, Tailwind CSS, and vanilla JavaScript. The AI logic is powered by a robust backend server written entirely in Go (Golang), which uses the Alpha-Beta Pruning algorithm to determine its moves.

Features

Interactive 9x9 Go Board: Play stones directly in your browser.

AI Opponent: A Go backend featuring an Alpha-Beta Pruning search algorithm.

Optimized AI: The AI uses an optimized move generation strategy (checking only adjacent squares) to ensure fast response times.

Full Game Logic: Includes rules for:

Stone Captures

Suicide Moves (Illegal)

Ko Rule (Illegal)

Game Controls:

Pass Turn: Allows you to pass your move.

New Game: Resets the board and state.

Get Score: Calculates the final score (Territory + Captures + Komi).

Game Over: The game automatically ends and calculates the score if both players pass consecutively.

Tech Stack

Frontend: HTML, Tailwind CSS, Vanilla JavaScript (ES6+)

Backend: Go (Golang)

Algorithm: Minimax with Alpha-Beta Pruning

How to Run

You must run both the backend server and the frontend file.

1. Run the Backend (Go Server)

Prerequisites: You must have Go installed on your machine.

Navigate: Open your terminal and cd into the directory containing the backend code.

Run: Execute the following command:

go run go_backend.go



The server will start and listen on http://localhost:8080.

2. Run the Frontend (Web App)

Open: Simply open the index.html file in any modern web browser (like Chrome, Firefox, or Safari).

Play: The game will automatically connect to the running backend server. You can now place your first stone!

Project Structure

index.html: The complete frontend application (HTML, CSS, and JS).

go_backend.go: The complete Go backend server, including all AI logic and API handlers.

API Endpoints

The Go server exposes two simple API endpoints:

POST /move:

Accepts the current game state (board, captures, player, move, last board state).

Validates the player's move.

Runs the AI's findBestMove function.

Returns the new board state after both moves (player and AI) have been made.

POST /score:

Accepts the final board state and captures.

Returns the calculated final score for Black and White (including Komi).