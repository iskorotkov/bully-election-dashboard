package state

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"go.uber.org/zap"
)

var (
	StateNone = State{}
)

type State struct {
	Name       string `json:"name"`
	LeaderName string `json:"leaderName"`
	State      string `json:"state"`
}

func NewState(name string, leaderName string, state string) State {
	return State{
		Name:       name,
		LeaderName: leaderName,
		State:      state,
	}
}

func NewUnknownState(name string) State {
	return State{
		Name:       name,
		LeaderName: "unknown",
		State:      "unknown",
	}
}

type dto struct {
	Data  []State `json:"data"`
	Error string  `json:"error"`
}

type Server struct {
	data   []State
	mu     sync.RWMutex
	logger *zap.Logger
}

func NewServer(logger *zap.Logger) *Server {
	return &Server{
		logger: logger,
	}
}

func (s *Server) Handle(rw http.ResponseWriter, r *http.Request) {
	logger := s.logger.Named("handle")
	logger.Debug("incoming request for fetching data",
		zap.Any("request", r),
		zap.Any("data", s.data))

	data := s.Data()

	var resp dto
	if data == nil {
		resp = dto{
			Data:  nil,
			Error: "No data available",
		}
	} else {
		resp = dto{
			Data:  data,
			Error: "",
		}
	}

	logger.Debug("sending response",
		zap.Any("response", resp))

	b, err := json.Marshal(resp)
	if err != nil {
		msg := "couldn't marshal response to json"
		logger.Error(msg,
			zap.Any("response", resp),
			zap.Error(err))
		http.Error(rw, msg, http.StatusInternalServerError)
	}

	rw.Header().Add("Content-Type", "application/json")
	fmt.Fprint(rw, string(b))
}

func (s *Server) Update(data []State) {
	s.logger.Debug("data updated",
		zap.Any("data", data))

	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = data
}

func (s *Server) Data() []State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data
}
