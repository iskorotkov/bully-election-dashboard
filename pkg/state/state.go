package state

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"go.uber.org/zap"
)

type ReplicaState struct {
	Name   string `json:"name"`
	Leader string `json:"leader,omitempty"`
	State  string `json:"state,omitempty"`
}

func NewReplicaState(name string, leader string, state string) ReplicaState {
	return ReplicaState{
		Name:   name,
		Leader: leader,
		State:  state,
	}
}

func NewUnknownReplicaState(name string) ReplicaState {
	return ReplicaState{
		Name:   name,
		Leader: "",
		State:  "",
	}
}

type State struct {
	Leader   string         `json:"leader,omitempty"`
	Replicas []ReplicaState `json:"replicas,omitempty"`
}

func (s State) Empty() bool {
	return s.Replicas == nil || len(s.Replicas) == 0
}

type dto struct {
	Data  State  `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

type Server struct {
	state  State
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
		zap.Any("data", s.state))

	state := s.State()

	var resp dto
	if state.Empty() {
		resp = dto{
			Data:  state,
			Error: "No data available",
		}
	} else {
		resp = dto{
			Data:  state,
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

func (s *Server) Update(replicas []ReplicaState) {
	s.logger.Debug("data updated",
		zap.Any("data", replicas))

	leader := ""
	for _, replica := range replicas {
		if leader == "" || replica.Name > leader {
			leader = replica.Name
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.state = State{
		Leader:   leader,
		Replicas: replicas,
	}
}

func (s *Server) State() State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}
