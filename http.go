package main

import (
	"net/http"

	"github.com/deciphernow/gm-fabric-go/dbutil"
	"github.com/rs/zerolog/hlog"
)

// Load ...
func Load(w http.ResponseWriter, r *http.Request) {
	log := hlog.FromRequest(r)
	wg := getServiceWG(r.Context())
	var s service
	err := dbutil.ReadReqest(r, &s)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read request body")
		return
	}

	finished := make(chan bool, 1)
	// Maintain the waitgroup so we don't exit the ping unexpectedly
	wg.Add(1)
	go hitService(s, finished, *log)
	go isFinished(finished, s, wg, *log)
}

// Loads ...
func Loads(w http.ResponseWriter, r *http.Request) {
	log := hlog.FromRequest(r)
	wg := getServiceWG(r.Context())
	var s services
	err := dbutil.ReadReqest(r, &s)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read request body")
		return
	}

	for _, service := range s.Service {
		wg.Add(1)
		finished := make(chan bool, 1)
		go hitService(service, finished, *log)
		go isFinished(finished, service, wg, *log)
	}
}
