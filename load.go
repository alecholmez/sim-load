package main

import (
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

func hitService(s service, finished chan bool, log zerolog.Logger) {
	var rwg sync.WaitGroup
	client := &http.Client{}

	for _, route := range s.Routes {
		rwg.Add(1)
		go func(route, location, load string) {
			req, err := http.NewRequest("GET", location+route, nil)
			if err != nil {
				log.Error().Err(err).Msg("Failed to create new http request")
			}

			req.Close = true
			req.Header.Add("x-request-id", uuid.New().String())

			for {
				resp, err := client.Do(req)
				if err != nil {
					log.Error().Err(err).Msg("Failed to hit service: " + location + route)
					break
				}
				resp.Body.Close()

				switch {
				case load == "light":
					time.Sleep(time.Microsecond * time.Duration(randomInt(1000, 2000)))
				case load == "heavy":
					time.Sleep(time.Microsecond * time.Duration(randomInt(100, 500)))
				default:
					time.Sleep(time.Millisecond * time.Duration(randomInt(1, 1000)))
				}
			}
			rwg.Done()
		}(route, s.Location, s.Load)
	}

	rwg.Wait()
	finished <- true
}

func randomInt(min, max int) int {
	return rand.Intn(max-min) + min
}

func isFinished(finished chan bool, service service, wg *sync.WaitGroup, log zerolog.Logger) {
	for {
		select {
		case fin := <-finished:
			if fin {
				wg.Done()
				log.Info().Msg("Finished pinging " + service.Name)
			}
		}
	}
}
