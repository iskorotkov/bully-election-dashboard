package ui

import (
	"html/template"
	"net/http"
	"os"

	"go.uber.org/zap"
)

type Server struct {
	hostname  string
	namespace string
	template  *template.Template
	logger    *zap.Logger
}

func NewServer(namespace string, logger *zap.Logger) (*Server, error) {
	hostname, err := os.Hostname()
	if err != nil {
		logger.Error("couldn't get hostname",
			zap.Error(err))
		return nil, err
	}

	tpl, err := template.ParseFiles("./web/template/ui.html")
	if err != nil {
		logger.Error("couldn't parse files",
			zap.Error(err))
		return nil, err
	}

	return &Server{
		hostname:  hostname,
		namespace: namespace,
		template:  tpl,
		logger:    logger,
	}, nil
}

func (s *Server) Handle(rw http.ResponseWriter, r *http.Request) {
	err := s.template.Execute(rw, struct {
		Hostname  string
		Namespace string
	}{
		Hostname:  s.hostname,
		Namespace: s.namespace,
	})
	if err != nil {
		msg := "couldn't execute template"
		s.logger.Error(msg,
			zap.Error(err))
		http.Error(rw, msg, http.StatusInternalServerError)
		return
	}
}
