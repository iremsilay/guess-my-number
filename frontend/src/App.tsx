import { useCallback, useEffect, useState } from "react";

const API_URL = import.meta.env.VITE_API_URL ?? "";
const MAX_ATTEMPTS = 10;

type GameStatus = "in_progress" | "won" | "lost";

type GuessResult = {
  guess: string;
  correctNumbers: number;
  correctPositions: number;
};

type GameResponse = {
  gameId: string;
  remainingAttempts: number;
  status: GameStatus;
};

type GuessResponse = Omit<GuessResult, "guess"> & {
  remainingAttempts: number;
  status: GameStatus;
};

const digits = ["1", "2", "3", "4", "5", "6", "7", "8", "9", "0"];

async function parseResponse<T>(response: Response): Promise<T> {
  const body = await response.json();
  if (!response.ok) {
    throw new Error(body.error ?? "Something went wrong. Please try again.");
  }
  return body as T;
}

function App() {
  const [gameId, setGameId] = useState("");
  const [guess, setGuess] = useState("");
  const [history, setHistory] = useState<GuessResult[]>([]);
  const [remainingAttempts, setRemainingAttempts] = useState(MAX_ATTEMPTS);
  const [status, setStatus] = useState<GameStatus>("in_progress");
  const [message, setMessage] = useState("Starting a new game…");
  const [loading, setLoading] = useState(true);

  const startGame = useCallback(async () => {
    setLoading(true);
    setMessage("Starting a new game…");
    try {
      const response = await fetch(`${API_URL}/games`, { method: "POST" });
      const game = await parseResponse<GameResponse>(response);
      setGameId(game.gameId);
      setGuess("");
      setHistory([]);
      setRemainingAttempts(game.remainingAttempts);
      setStatus(game.status);
      setMessage("Enter four different digits.");
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Could not start the game.");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void startGame();
  }, [startGame]);

  const addDigit = useCallback(
    (digit: string) => {
      if (loading || status !== "in_progress" || guess.length === 4) return;
      if (guess.includes(digit)) {
        setMessage("Each digit can be used only once.");
        return;
      }
      setGuess((current) => current + digit);
      setMessage("Enter four different digits.");
    },
    [guess, loading, status],
  );

  const removeDigit = useCallback(() => {
    if (loading || status !== "in_progress") return;
    setGuess((current) => current.slice(0, -1));
    setMessage("Enter four different digits.");
  }, [loading, status]);

  const submitGuess = useCallback(async () => {
    if (loading || status !== "in_progress") return;
    if (guess.length !== 4) {
      setMessage("Your guess must contain four digits.");
      return;
    }

    setLoading(true);
    try {
      const response = await fetch(`${API_URL}/games/${gameId}/guesses`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ guess }),
      });
      const result = await parseResponse<GuessResponse>(response);
      setHistory((current) => [
        ...current,
        {
          guess,
          correctNumbers: result.correctNumbers,
          correctPositions: result.correctPositions,
        },
      ]);
      setRemainingAttempts(result.remainingAttempts);
      setStatus(result.status);
      setGuess("");

      if (result.status === "won") setMessage("You cracked the code!");
      else if (result.status === "lost") setMessage("No attempts left. Start a new game.");
      else setMessage(`${result.remainingAttempts} attempts remaining.`);
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Could not submit the guess.");
    } finally {
      setLoading(false);
    }
  }, [gameId, guess, loading, status]);

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (/^[0-9]$/.test(event.key)) {
        event.preventDefault();
        addDigit(event.key);
      } else if (event.key === "Backspace" || event.key === "Delete") {
        event.preventDefault();
        removeDigit();
      } else if (event.key === "Enter") {
        event.preventDefault();
        void submitGuess();
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [addDigit, removeDigit, submitGuess]);

  const rows = Array.from({ length: MAX_ATTEMPTS }, (_, index) => history[index]);
  const statusClass = status === "won" ? "success" : status === "lost" ? "error" : "";

  return (
    <main className="page-shell">
      <header className="page-header">
        <p className="eyebrow">Code-breaking game</p>
        <h1>Guess My Number</h1>
        <p className="attempts" aria-live="polite">
          {remainingAttempts} of {MAX_ATTEMPTS} attempts left
        </p>
      </header>

      <div className="game-layout">
        <section className="game-area" aria-label="Guessing game">
          <div className="score-card">
            <div className="score-grid score-header" aria-hidden="true">
              <span>Your guess</span>
              <span>Correct numbers</span>
              <span>Correct positions</span>
            </div>
            <div className="score-body">
              {rows.map((row, index) => (
                <div className="score-grid score-row" key={index}>
                  <span className="round-number">{index + 1}</span>
                  <span className="guess-value">{row?.guess ?? "— — — —"}</span>
                  <span>{row?.correctNumbers ?? "—"}</span>
                  <span>{row?.correctPositions ?? "—"}</span>
                </div>
              ))}
            </div>
          </div>

          <div className="controls-card">
            <div className="digit-display" aria-label={`Current guess: ${guess || "empty"}`}>
              {Array.from({ length: 4 }, (_, index) => (
                <span className={guess[index] ? "filled" : ""} key={index}>
                  {guess[index] ?? ""}
                </span>
              ))}
            </div>

            <p className={`game-message ${statusClass}`} role="status">
              {message}
            </p>

            <div className="keypad" aria-label="Number keypad">
              {digits.map((digit) => (
                <button
                  className="key"
                  disabled={loading || status !== "in_progress" || guess.includes(digit)}
                  key={digit}
                  onClick={() => addDigit(digit)}
                  type="button"
                >
                  {digit}
                </button>
              ))}
              <button
                className="key delete-key"
                disabled={loading || status !== "in_progress" || guess.length === 0}
                onClick={removeDigit}
                type="button"
                aria-label="Delete last digit"
              >
                Delete
              </button>
              <button
                className="key submit-key"
                disabled={loading || status !== "in_progress" || guess.length !== 4}
                onClick={() => void submitGuess()}
                type="button"
              >
                GO
              </button>
            </div>

            {status !== "in_progress" && (
              <button className="new-game-button" onClick={() => void startGame()} type="button">
                New game
              </button>
            )}
          </div>
        </section>

        <aside className="rules-card">
          <p className="rules-label">How to play</p>
          <h2>Crack the four-digit code.</h2>
          <p>You have 10 attempts. The code may start with 0, and each digit is used only once.</p>
          <dl>
            <div>
              <dt>Correct numbers</dt>
              <dd>Digits that exist anywhere in the secret code.</dd>
            </div>
            <div>
              <dt>Correct positions</dt>
              <dd>Digits that are also in the exact right place.</dd>
            </div>
          </dl>
          <p className="rules-note">A correctly positioned digit is included in both results.</p>
        </aside>
      </div>
    </main>
  );
}

export default App;
