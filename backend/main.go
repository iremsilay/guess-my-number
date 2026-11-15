package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

type api struct {
	mu    sync.Mutex
	games map[string]*game
}

func newAPI() *api {
	return &api{games: make(map[string]*game)}
}

func (a *api) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /games", a.createGame)
	mux.HandleFunc("POST /games/{gameId}/guesses", a.submitGuess)
	return mux
}

func (a *api) createGame(w http.ResponseWriter, r *http.Request) {
	secret, err := generateCode()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not generate a code")
		return
	}

	var idBytes [16]byte
	if _, err := rand.Read(idBytes[:]); err != nil {
		writeError(w, http.StatusInternalServerError, "could not generate a game ID")
		return
	}
	id := hex.EncodeToString(idBytes[:])

	a.mu.Lock()
	_, exists := a.games[id]
	if !exists {
		a.games[id] = &game{secret: secret, status: inProgress}
	}
	a.mu.Unlock()
	if exists {
		writeError(w, http.StatusInternalServerError, "could not allocate a game ID")
		return
	}

	writeJSON(w, http.StatusCreated, struct {
		GameID            string `json:"gameId"`
		RemainingAttempts int    `json:"remainingAttempts"`
		Status            string `json:"status"`
	}{id, maxAttempts, inProgress})
}

func (a *api) submitGuess(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Guess string `json:"guess"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024))
	if err := decoder.Decode(&input); err != nil || decoder.Decode(new(json.RawMessage)) != io.EOF {
		writeError(w, http.StatusBadRequest, "expected one JSON object with a guess string")
		return
	}
	if !validCode(input.Guess) {
		writeError(w, http.StatusBadRequest, "guess must contain four distinct digits")
		return
	}

	a.mu.Lock()
	g, exists := a.games[r.PathValue("gameId")]
	switch {
	case !exists:
		a.mu.Unlock()
		writeError(w, http.StatusNotFound, "game not found")
	case g.status != inProgress:
		a.mu.Unlock()
		writeError(w, http.StatusConflict, "game is already over")
	default:
		result := g.submit(input.Guess)
		a.mu.Unlock()
		writeJSON(w, http.StatusOK, result)
	}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("write response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func main() {
	server := &http.Server{
		Addr:              "localhost:8080",
		Handler:           newAPI().routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
	}
	log.Printf("Listening on http://%s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
