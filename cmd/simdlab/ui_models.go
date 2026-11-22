package main

import (
	"net/http"
	"slices"
	"strings"

	"github.com/akolybelnikov/go-simd/internal/modelcatalog"
)

type localChatModel struct {
	Name      string `json:"name"`
	Installed bool   `json:"installed"`
	modelcatalog.Assessment
}

// chatModels reads only the configured local server. It never pulls a model.
func (app *localUI) chatModels(w http.ResponseWriter, r *http.Request) {
	response := struct {
		Enabled            bool             `json:"enabled"`
		Initial            string           `json:"initial_model"`
		Models             []localChatModel `json:"models"`
		Error              string           `json:"error,omitempty"`
		HasRecommendations bool             `json:"has_recommendations"`
	}{Enabled: app.localChat != nil, Initial: app.initialChatModel, Models: []localChatModel{}, HasRecommendations: len(modelcatalog.BrowserRecommendations()) > 0}
	if app.localChat != nil {
		names, err := app.localChat.localModels(r.Context())
		if err != nil {
			response.Error = err.Error()
		}
		for _, name := range names {
			response.Models = append(response.Models, localChatModel{name, true, modelcatalog.Browser(name)})
		}
		for _, recommended := range modelcatalog.BrowserRecommendations() {
			if !slices.Contains(names, recommended) {
				response.Models = append(response.Models, localChatModel{recommended, false, modelcatalog.Browser(recommended)})
			}
		}
		if app.initialChatModel != "" && !slices.Contains(names, app.initialChatModel) && !slices.Contains(modelcatalog.BrowserRecommendations(), app.initialChatModel) {
			name := app.initialChatModel
			if !strings.Contains(name[strings.LastIndex(name, "/")+1:], ":") {
				name += ":latest"
			}
			if !slices.Contains(names, name) {
				response.Models = append(response.Models, localChatModel{app.initialChatModel, false, modelcatalog.Browser(app.initialChatModel)})
			}
		}
	}
	uiJSON(w, http.StatusOK, response)
}
