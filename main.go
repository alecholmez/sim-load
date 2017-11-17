package main

import (
	"flag"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type service struct {
	Name     string   `toml:"name"`
	Routes   []string `toml:"routes"`
	Location string   `toml:"location"`
	Load     string   `toml:"load"`
}

type services struct {
	Service []service
}

var (
	settings = flag.String("services", "./services.toml", "Path to services.toml")
)

func main() {
	flag.Parse()

	var s services
	var wg sync.WaitGroup
	rand.Seed(time.Now().UnixNano())

	if _, err := toml.DecodeFile(*settings, &s); err != nil {
		log.Print(err)
		return
	}

	log.Print("simulating load...")
	for _, service := range s.Service {
		wg.Add(1)
		finished := make(chan bool, 1)
		go hitService(service, finished, &wg)
		go isFinished(finished, service, &wg)
	}

	wg.Wait()
	log.Print("Exiting. . .")
}

func hitService(s service, finished chan bool, wg *sync.WaitGroup) {
	var rwg sync.WaitGroup
	client := &http.Client{}

	for _, route := range s.Routes {
		rwg.Add(1)
		go func(route, location, load string) {
			req, err := http.NewRequest("GET", location+route, nil)
			if err != nil {
				log.Print(err)
			}

			req.Close = true
			req.Header.Add("x-request-id", uuid.New().String())

			for {
				resp, err := client.Do(req)
				if err != nil {
					log.Print(err)
					break
				}
				resp.Body.Close()

				switch {
				case load == "light":
					time.Sleep(time.Millisecond * time.Duration(randomInt(500, 1500)))
				case load == "heavy":
					time.Sleep(time.Millisecond * time.Duration(randomInt(5, 350)))
				default:
					time.Sleep(time.Second * 2)
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

func isFinished(finished chan bool, service service, wg *sync.WaitGroup) {
	for {
		select {
		case fin := <-finished:
			if fin {
				wg.Done()
				log.Print("Finished pinging " + service.Name)
			}
		}
	}
}
