// Package wiki renders the lab's embedded Markdown documentation.
package wiki

import (
	"bytes"
	"embed"
	"io/fs"
	"net/url"
	"regexp"
	"strings"

	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/extension"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
	"github.com/yuin/goldmark/v2/text"
)

//go:embed *.md
var sources embed.FS

const Guide = "32-Operation-measurement-guide"

const repository = "/akolybelnikov/go-simd"

type OperationNote struct {
	Summary string
	URL     string
}

// Each marked section begins with the same plain-text introduction shown in Explore.
var operationSection = regexp.MustCompile(`(?m)^## ([A-Za-z0-9 ]+)\r?\n\r?\n<!-- operation: ([a-z0-9]+) -->\r?\n\r?\n((?:[^\r\n]+\r?\n)+)`)

var operationNotes = func() map[string]OperationNote {
	source, _ := sources.ReadFile(Guide + ".md")
	notes := make(map[string]OperationNote)
	for _, section := range operationSection.FindAllStringSubmatch(string(source), -1) {
		anchor := strings.Join(strings.Fields(strings.ToLower(section[1])), "-")
		notes[section[2]] = OperationNote{
			Summary: strings.Join(strings.Fields(section[3]), " "),
			URL:     "/wiki/" + Guide + "#" + anchor,
		}
	}
	return notes
}()

func Operation(group string) OperationNote {
	return operationNotes[group]
}

func isPage(name string) bool {
	if name == "" || strings.HasPrefix(name, "_") || strings.ContainsAny(name, "/\\") {
		return false
	}
	_, err := fs.Stat(sources, name+".md")
	return err == nil
}

// Render returns the title and rendered HTML of an embedded page. The HTML
// renderer leaves raw Markdown HTML and unsafe links disabled.
func Render(name string) (string, string, error) {
	if name == "" {
		name = "Home"
	}
	if !isPage(name) {
		return "", "", fs.ErrNotExist
	}
	source, err := sources.ReadFile(name + ".md")
	if err != nil {
		return "", "", err
	}
	p := parser.New(parser.WithAutoHeadingID(), parser.WithExtensions(extension.NewGFMParser()))
	doc := p.Parse(source)
	if err := ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if link, ok := node.(*ast.Link); entering && ok {
			destination := link.Destination.Value(source)
			if local := localDestination(destination); local != destination {
				link.Destination = text.NewSingleLineValueFromString(local, text.IdentityDecoder)
			}
		}
		return ast.WalkContinue, nil
	}); err != nil {
		return "", "", err
	}
	var body bytes.Buffer
	renderer := html.New(html.WithExtensions(extension.NewGFMHTMLRenderer()))
	if err := renderer.Render(&body, source, doc); err != nil {
		return "", "", err
	}
	firstLine, _, _ := strings.Cut(string(source), "\n")
	title := strings.TrimSpace(strings.TrimPrefix(firstLine, "# "))
	return title, body.String(), nil
}

func localDestination(destination string) string {
	link, err := url.Parse(destination)
	if err != nil || link.Path == "" {
		return destination
	}
	name := strings.TrimPrefix(link.Path, "./")
	if link.IsAbs() || link.Host != "" {
		if link.Scheme != "https" || link.Host != "github.com" {
			return destination
		}
		var local bool
		switch {
		case link.Path == repository+"/wiki" || link.Path == repository+"/wiki/":
			name, local = "Home", true
		default:
			for _, prefix := range []string{repository + "/wiki/", repository + "/blob/main/docs/wiki/"} {
				if page, ok := strings.CutPrefix(link.Path, prefix); ok {
					name, local = page, true
					break
				}
			}
		}
		if !local {
			return destination
		}
	}
	name = strings.TrimSuffix(name, ".md")
	if !isPage(name) {
		return destination
	}
	return (&url.URL{Path: "/wiki/" + name, Fragment: link.Fragment, RawQuery: link.RawQuery}).String()
}
