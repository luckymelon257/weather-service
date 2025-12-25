package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-co-op/gocron/v2"
	"github.com/luckymelon257/weather-service/internal/client/http/geocoding"
	"github.com/luckymelon257/weather-service/internal/client/http/open-meteo"
)

const httpPort = ":3000"

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}
	geocodingClient := geocoding.NewClient(httpClient)
	openMeteoClient := open_meteo.NewClient(httpClient)
	r.Get("/{city}", func(w http.ResponseWriter, r *http.Request) {
		city := chi.URLParam(r, "city")

		geoRes, err := geocodingClient.GetCoords(city)
		if err != nil {
			log.Println(err)
			return
		}
		openMeteoRes, err := openMeteoClient.GetTemperature(geoRes.Latitude, geoRes.Longitude)
		if err != nil {
			log.Println(err)
			return
		}

		raw, err := json.Marshal(openMeteoRes)
		_, err = w.Write([]byte(raw))

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
		}
	})

	s, err := gocron.NewScheduler()
	if err != nil {
		panic(err)
	}

	jobs, err := initJobs(s)
	if err != nil {
		panic(err)
	}

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()

		fmt.Printf("Server listening on %s", httpPort)
		err := http.ListenAndServe(httpPort, r)
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		defer wg.Done()

		fmt.Printf("Starting job: %v", jobs[0].ID())
		s.Start()
	}()

	wg.Wait()
}
func initJobs(scheduler gocron.Scheduler) ([]gocron.Job, error) {
	// add a job to the scheduler
	j, err := scheduler.NewJob(
		gocron.DurationJob(
			10*time.Second,
		),
		gocron.NewTask(
			func() {
				log.Println("Hello")
			},
		),
	)

	if err != nil {
		return nil, err
	}
	return []gocron.Job{j}, nil
}
