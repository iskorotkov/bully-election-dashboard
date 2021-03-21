package ui

import (
	"embed"
	"html/template"
	"net/http"
	"os"

	"go.uber.org/zap"
)

//go:embed template
var files embed.FS

type Server struct {
	hostname  string
	namespace string
	templates *template.Template
	logger    *zap.Logger
}

func NewServer(namespace string, logger *zap.Logger) (*Server, error) {
	hostname, err := os.Hostname()
	if err != nil {
		logger.Error("couldn't get hostname",
			zap.Error(err))
		return nil, err
	}

	templates, err := template.ParseFS(files, "template/*")
	if err != nil {
		logger.Error("couldn't parse files",
			zap.Error(err))
		return nil, err
	}

	return &Server{
		hostname:  hostname,
		namespace: namespace,
		templates: templates,
		logger:    logger,
	}, nil
}

func (s *Server) Handle(rw http.ResponseWriter, r *http.Request) {
	data := struct {
		Hostname  string
		Namespace string
	}{
		Hostname:  s.hostname,
		Namespace: s.namespace,
	}

	if err := s.templates.ExecuteTemplate(rw, "ui.html", data); err != nil {
		msg := "couldn't execute template"
		s.logger.Error(msg,
			zap.Error(err))
		http.Error(rw, msg, http.StatusInternalServerError)
		return
	}
}
