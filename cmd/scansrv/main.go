package main

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/pkg/sleuth/scan"
)

type scanReq struct {
	Domain string `json:"domain"`
}

func domainHandler(w http.ResponseWriter, r *http.Request) {
	var req scanReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Domain == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	report, err := scan.ScanDomain(r.Context(), req.Domain)
	if err != nil {
		log.Error().Err(err).Msg("domain scan failed")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(report)
}

func vulnHandler(w http.ResponseWriter, r *http.Request) {
	var req scanReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Domain == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	report, err := scan.ScanVulnerabilities(r.Context(), req.Domain)
	if err != nil {
		log.Error().Err(err).Msg("vulnerability scan failed")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(report)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/scan/domain", domainHandler)
	mux.HandleFunc("/scan/vulnerability", vulnHandler)
	log.Info().Msg("starting scanner service on :17700")
	if err := http.ListenAndServe(":17701", mux); err != nil {
		log.Fatal().Err(err).Msg("scanner shutdown")
	}
}
