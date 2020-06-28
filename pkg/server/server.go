package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"bean/pkg/core"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

type Server struct {
	core core.Core

	mdRenderer *html.Renderer
	mdParser   *parser.Parser

	tmplList *template.Template
	tmplCard *template.Template
}

func New(core core.Core) (*Server, error) {
	tmplList, err := template.New("topics").Parse("./static/topics.html")
	if err != nil {
		return nil, errors.Wrap(err, "parse template \"topics\"")
	}
	tmplCard, err := template.New("card").Parse("./static/card.html")
	if err != nil {
		return nil, errors.Wrap(err, "parse template \"card\"")
	}

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.HardLineBreak
	parser := parser.NewWithExtensions(extensions)

	return &Server{core, renderer, parser, tmplList, tmplCard}, nil
}

func (s *Server) HandleListTopics(w http.ResponseWriter, r *http.Request) {
	list, err := s.core.ListTopics()
	if err != nil {
		fmt.Fprint(w, "list topics: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := s.tmplList.Execute(w, list); err != nil {
		fmt.Fprint(w, "list topics: exec template: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) HandleNextCard(w http.ResponseWriter, r *http.Request) {
	topic := mux.Vars(r)["topic"]
	key, title, content, err := s.core.Get(topic)
	if err != nil {
		fmt.Fprintf(w, "get from topic %q: %v", topic, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	contentHtml := markdown.ToHTML(content, s.mdParser, s.mdRenderer)
	if err := s.tmplCard.Execute(w, struct {
		Key     string
		Topic   string
		Title   string
		Content []byte
	}{
		Key:     key,
		Topic:   topic,
		Title:   title,
		Content: contentHtml,
	}); err != nil {
		fmt.Fprintf(w, "execute card template: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *Server) HandleReflect(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		Topic   string `json:"topic"`
		Key     string `json:"key"`
		Quality int    `json:"quality"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		fmt.Fprintf(w, "decode request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := s.core.Reflect(reqBody.Topic, reqBody.Key, reqBody.Quality); err != nil {
		fmt.Fprintf(w, "reflect: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
