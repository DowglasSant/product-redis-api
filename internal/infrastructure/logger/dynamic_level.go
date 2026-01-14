package logger

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type AtomicLevelServer struct {
	level *zap.AtomicLevel
}

func NewAtomicLevelServer(level *zap.AtomicLevel) *AtomicLevelServer {
	return &AtomicLevelServer{
		level: level,
	}
}

func (s *AtomicLevelServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGet(w, r)
	case http.MethodPut:
		s.handlePut(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *AtomicLevelServer) handleGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"level": s.level.Level().String(),
	})
}

func (s *AtomicLevelServer) handlePut(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Level string `json:"level"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var level zapcore.Level
	if err := level.UnmarshalText([]byte(req.Level)); err != nil {
		http.Error(w, "Invalid log level. Valid levels: debug, info, warn, error, dpanic, panic, fatal", http.StatusBadRequest)
		return
	}

	s.level.SetLevel(level)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Log level updated successfully",
		"level":   level.String(),
	})
}

func (s *AtomicLevelServer) GetCurrentLevel() string {
	return s.level.Level().String()
}

func (s *AtomicLevelServer) SetLevel(level zapcore.Level) {
	s.level.SetLevel(level)
}

type LogLevelGauge struct {
	level *zap.AtomicLevel
}

func NewLogLevelGauge(level *zap.AtomicLevel) *LogLevelGauge {
	return &LogLevelGauge{level: level}
}

func (g *LogLevelGauge) Describe(ch chan<- interface{}) {
}

func (g *LogLevelGauge) Collect(ch chan<- interface{}) {
}

func (g *LogLevelGauge) LevelAsInt() int {
	return int(g.level.Level())
}
