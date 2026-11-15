package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGameAPI(t *testing.T) {
	for _, outcome := range []string{won, lost} {
		t.Run(outcome, func(t *testing.T) {
			a := newAPI()
			handler := a.routes()
			post := func(path, body string, wantStatus int) *httptest.ResponseRecorder {
				t.Helper()
				response := httptest.NewRecorder()
				handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, path, strings.NewReader(body)))
				if response.Code != wantStatus {
					t.Fatalf("%s returned %d, want %d. Body: %s", path, response.Code, wantStatus, response.Body)
				}
				return response
			}

			created := post("/games", "", http.StatusCreated)
			var start struct {
				GameID            string `json:"gameId"`
				RemainingAttempts int    `json:"remainingAttempts"`
				Status            string `json:"status"`
			}
			if err := json.Unmarshal(created.Body.Bytes(), &start); err != nil {
				t.Fatal(err)
			}
			g, exists := a.games[start.GameID]
			if !exists || !validCode(g.secret) || start.RemainingAttempts != maxAttempts || start.Status != inProgress {
				t.Fatalf("invalid game creation response: %+v", start)
			}
			if strings.Contains(created.Body.String(), `"secret"`) {
				t.Fatal("creation response exposes the secret")
			}
			g.secret = "0123"

			path := "/games/" + start.GameID + "/guesses"
			post("/games/missing/guesses", `{"guess":"0123"}`, http.StatusNotFound)
			for _, invalid := range []string{`{"guess":"0012"}`, `{"guess":"123"}`, `{"guess":1234}`, `{"guess":"12a3"}`, `{"guess":"0123"} {}`} {
				post(path, invalid, http.StatusBadRequest)
			}

			for attempt := 1; attempt <= maxAttempts; attempt++ {
				guess, status, matches := "4567", inProgress, 0
				if attempt == maxAttempts {
					status = outcome
					if outcome == won {
						guess, matches = "0123", 4
					}
				}
				response := post(path, `{"guess":"`+guess+`"}`, http.StatusOK)
				var result guessResult
				if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
					t.Fatal(err)
				}
				want := guessResult{matches, matches, maxAttempts - attempt, status}
				if result != want {
					t.Errorf("attempt %d: got %+v, want %+v", attempt, result, want)
				}
			}
			post(path, `{"guess":"0123"}`, http.StatusConflict)
		})
	}
}
