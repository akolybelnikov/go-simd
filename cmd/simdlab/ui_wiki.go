package main

import (
	"bytes"
	"errors"
	"html/template"
	"io/fs"
	"net/http"
	"strings"

	"github.com/akolybelnikov/go-simd/docs/wiki"
)

var wikiPage = template.Must(template.New("wiki").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{.Title}} · SIMD Lab Wiki</title>
  <link rel="stylesheet" href="/style.css">
</head>
<body>
  <header class="top">
    <div class="top-inner">
      <a class="brand wiki-home" href="/wiki/">SIMD Lab Wiki</a>
      <span class="local">Local documentation</span>
    </div>
  </header>
  <main class="wiki">
    <nav aria-label="Wiki">
      <a href="/wiki/">Wiki home</a>
      <a href="/wiki/00-What-is-SIMD">What is SIMD?</a>
      <a href="/wiki/{{.Guide}}">Measurement guide</a>
      <a href="/wiki/06-Understanding-results">Understanding results</a>
      <a href="/wiki/34-AI-setup-and-model-recommendations">AI setup</a>
    </nav>
    <article class="card">{{.Body}}</article>
  </main>
</body>
</html>
`))

func serveWiki(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "Use GET or HEAD.", http.StatusMethodNotAllowed)
		return
	}
	page := strings.TrimPrefix(r.URL.Path, "/wiki")
	title, body, err := wiki.Render(strings.TrimPrefix(page, "/"))
	if errors.Is(err, fs.ErrNotExist) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "Could not render this wiki page.", http.StatusInternalServerError)
		return
	}
	var output bytes.Buffer
	view := struct {
		Title string
		Guide string
		Body  template.HTML
	}{Title: title, Guide: wiki.Guide, Body: template.HTML(body)} // Markdown renderer strips raw HTML and unsafe URLs.
	if err := wikiPage.Execute(&output, view); err != nil {
		http.Error(w, "Could not display this wiki page.", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if r.Method == http.MethodGet {
		_, _ = output.WriteTo(w)
	}
}
