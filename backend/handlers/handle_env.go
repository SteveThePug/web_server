package handlers

import (
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	"adam-french.co.uk/backend/services"
	"github.com/gin-gonic/gin"
)

// This file backs the admin panel's Environment tab: a report on every
// variable in .env, and a way to change the ones that are safe to change.
//
// The handlers never read or write .env. State is this process's own
// environment, and a change is a request left for the host to apply (see
// services.EnvManager). REST rather than GraphQL for the same reason as the
// rest of the admin endpoints: this is process and host state, not the data
// model the schema describes.

// envSecretHintMinLen is the shortest secret that gets its last four
// characters shown. Below it, four characters is too much of the value.
const envSecretHintMinLen = 12

type envVarView struct {
	services.EnvVar
	Secret bool `json:"secret"`
	Set    bool `json:"set"`
	// Value is sent only for non-secret variables; Hint (the last four
	// characters) is all a secret ever reveals.
	Value string `json:"value,omitempty"`
	Hint  string `json:"hint,omitempty"`
}

type envGroupView struct {
	Name string       `json:"name"`
	Vars []envVarView `json:"vars"`
}

// GetEnv backs GET /admin/env: every registry variable grouped for display,
// plus the state of the latest change.
//
// Values are what this process was started with, not what .env holds now —
// the two differ between a save and the recreate that follows it, which is
// what the status's pending list is for.
func (store *Store) GetEnv(ctx *gin.Context) {
	var groups []envGroupView
	index := map[string]int{}
	for _, v := range store.Env.Vars() {
		value := os.Getenv(v.Name)
		view := envVarView{EnvVar: v, Secret: v.Secret(), Set: strings.TrimSpace(value) != ""}
		switch {
		case !view.Set:
		case !view.Secret:
			view.Value = value
		case len(value) >= envSecretHintMinLen:
			view.Hint = value[len(value)-4:]
		}

		i, seen := index[v.Group]
		if !seen {
			i = len(groups)
			index[v.Group] = i
			groups = append(groups, envGroupView{Name: v.Group})
		}
		groups[i].Vars = append(groups[i].Vars, view)
	}
	ctx.JSON(http.StatusOK, gin.H{"groups": groups, "status": store.Env.Status()})
}

type updateEnvRequest struct {
	// Set maps a variable to its new value; "" clears it.
	Set map[string]string `json:"set"`
	// Generate names "generated" variables to mint a random value for.
	Generate []string `json:"generate"`
}

// UpdateEnv backs PUT /admin/env. It queues the change for the host and
// returns straight away; the panel follows progress on GET /admin/env/status,
// through the backend's own restart.
//
// The response is the only time a generated value is sent anywhere.
func (store *Store) UpdateEnv(ctx *gin.Context) {
	var body updateEnvRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	generated, recreates, err := store.Env.Submit(body.Set, body.Generate)
	if errors.Is(err, services.ErrEnvApplying) {
		ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		// Validation messages name the variable, never the value.
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	names := make([]string, 0, len(body.Set)+len(body.Generate))
	for name := range body.Set {
		names = append(names, name)
	}
	names = append(names, body.Generate...)
	log.Printf("[Env] change queued for %s", strings.Join(names, ", "))

	ctx.Header("Cache-Control", "no-store")
	ctx.JSON(http.StatusAccepted, gin.H{
		"generated": generated,
		"services":  recreates,
		"status":    store.Env.Status(),
	})
}

// GetEnvStatus backs GET /admin/env/status.
func (store *Store) GetEnvStatus(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, store.Env.Status())
}

// CancelEnv backs DELETE /admin/env/pending: withdraw a request nothing on
// the host has picked up, so new secrets do not sit in the shared directory
// waiting for a watcher that is not there.
func (store *Store) CancelEnv(ctx *gin.Context) {
	if err := store.Env.Cancel(); err != nil {
		log.Println("Cancelling env request failed:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel the request"})
		return
	}
	ctx.JSON(http.StatusOK, store.Env.Status())
}
