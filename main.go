package main

import (
	"context"
	"flag"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/deciphernow/gm-fabric-go/middleware"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

const (
	wgKey key = iota
)

var (
	useHTTP = flag.Bool("http", true, "Use HTTP server")
	config  = flag.String("config", "./services.toml", "Path to services.toml")
)

func main() {
	flag.Parse()
	log := zerolog.New(os.Stderr).With().Timestamp().Logger().
		Output(zerolog.ConsoleWriter{Out: os.Stderr})

	var s services
	var wg sync.WaitGroup
	rand.Seed(time.Now().UnixNano())

	if *config != "" {
		if _, err := toml.DecodeFile(*config, &s); err != nil {
			log.Error().Err(err).Msg("Failed to decode config file")
			return
		}

		log.Info().Msg("simulating load...")
		for _, service := range s.Service {
			wg.Add(1)
			finished := make(chan bool, 1)
			go hitService(service, finished, log)
			go isFinished(finished, service, &wg, log)
		}
	}

	if *useHTTP {
		router := mux.NewRouter()
		router.HandleFunc("/load", Load).Methods("POST")
		router.HandleFunc("/loads", Load).Methods("POST")

		stack := middleware.Chain(
			middleware.MiddlewareFunc(hlog.NewHandler(log)),
			middleware.MiddlewareFunc(hlog.AccessHandler(func(r *http.Request, status int, size int, duration time.Duration) {
				hlog.FromRequest(r).Info().
					Str("method", r.Method).
					Str("path", r.URL.String()).
					Int("status", status).
					Int("size", size).
					Dur("duration", duration).
					Msg("Access")
			})),
			withServiceWG(&wg),
		)

		s := http.Server{
			Addr:    os.Getenv("SERVICE_ADDR"),
			Handler: stack.Wrap(router),
		}

		wg.Add(1)
		go func() {
			log.Info().Msg("Service listening on: " + os.Getenv("SERVICE_ADDR"))
			log.Fatal().Err(s.ListenAndServe()).Msg("Error when starting service")
			wg.Done()
		}()
	}

	wg.Wait()
	log.Info().Msg("Exiting. . .")
}

func withServiceWG(wg *sync.WaitGroup) middleware.Middleware {
	return middleware.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), wgKey, wg))
			next.ServeHTTP(w, r)
		})
	})
}

func getServiceWG(ctx context.Context) *sync.WaitGroup {
	wg, ok := ctx.Value(wgKey).(*sync.WaitGroup)
	if !ok {
		log.Fatal("Failed to retrieve waitgroup from context")
	}

	return wg
}
