package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/dougedey/elsinore/database"
	"github.com/dougedey/elsinore/devices"
	"github.com/dougedey/elsinore/graph"
	"github.com/dougedey/elsinore/graph/generated"
	"github.com/dougedey/elsinore/hardware"
	"github.com/dougedey/elsinore/system"
	"github.com/go-chi/chi"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"github.com/ztrue/shutdown"
	"periph.io/x/periph/conn/onewire"
	"periph.io/x/periph/host"

	"net/http"
)

func main() {
	portPtr := flag.String("port", "8080", "The port to listen on")
	graphiqlFlag := flag.Bool("graphiql", true, "Disable GraphiQL web UI")
	dbName := flag.String("db_name", "elsinore", "The path/name of the local database")
	testDeviceFlag := flag.Bool("test_device", false, "Create a test device")
	autostartFlag := flag.Bool("autostart", false, "Autostart controllers from their previous state on startup")
	flag.Parse()

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	quit := make(chan struct{})

	if *testDeviceFlag {
		realAddress := "ARealAddress"
		hardware.SetProbe(&hardware.TemperatureProbe{
			PhysAddr: realAddress,
			Address:  onewire.Address(12345),
		})
	}

	database.InitDatabase(dbName,
		&devices.TempProbeDetail{}, &devices.PidSettings{}, &devices.HysteriaSettings{},
		&devices.ManualSettings{}, &devices.TemperatureController{}, &system.Settings{},
		&devices.Switch{},
	)

	if len(strings.TrimSpace(system.CurrentSettings().BreweryName)) == 0 {
		system.CurrentSettings().BreweryName = "Elsinore"
		system.CurrentSettings().Save()
	}
	log.Printf("Starting %v", system.CurrentSettings().BreweryName)

	_, err := host.Init()
	if err != nil {
		log.Fatal().
			Msgf("failed to initialize periph: %v", err)
	}

	log.Print("Loaded and looking for temperatures")
	// messages := make(chan string)
	go hardware.ReadTemperatures(nil, quit)
	for _, controller := range devices.AllTemperatureControllers() {
		if *autostartFlag {
			continue
		}
		controller.Mode = "off"
	}
	go temperatureControllerRunner()

	log.Printf("Loaded %v switches.", len(devices.AllSwitches()))

	httpServerExitDone := &sync.WaitGroup{}

	httpServerExitDone.Add(1)
	srv := startHTTPServer(portPtr, graphiqlFlag, httpServerExitDone)

	shutdown.Add(func() {
		devices.CancelFunc()
		shutdownErr := srv.Shutdown(devices.Context)
		if shutdownErr != nil {
			log.Print(shutdownErr)
		}
	})

	shutdown.Listen()
}

func startHTTPServer(portPtr *string, graphiqlFlag *bool, wg *sync.WaitGroup) *http.Server {
	router := chi.NewRouter()

	// Add CORS middleware around every request
	// See https://github.com/rs/cors for full option listing
	router.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: false,
		Debug:            true,
	}).Handler)

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}}))

	srv.SetErrorPresenter(func(ctx context.Context, e error) *gqlerror.Error {
		err := graphql.DefaultErrorPresenter(ctx, e)
		log.Error().Err(err).Msg("GraphQL Error")
		return err
	})
	httpSrv := &http.Server{Addr: ":" + *portPtr, Handler: router}

	if *graphiqlFlag {
		router.Handle("/", playground.Handler("GraphQL playground", "/graphiql"))
	}
	router.Handle("/graphql", srv)

	go func() {
		defer wg.Done()
		err := httpSrv.ListenAndServe()
		log.Fatal().Err(err).Msg("Failed to start server")
	}()

	fullPort := fmt.Sprintf(":%v", *portPtr)

	name, err := os.Hostname()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get hostname: %v")
	} else {
		fmt.Printf("CORS API Listening on: http://%v%v/graphql \n", name, fullPort)
		if *graphiqlFlag {
			fmt.Printf("GraphiQL interface: http://%v%v/graphiql \n", name, fullPort)
		}
	}
	return httpSrv
}

func temperatureControllerRunner() {
	fmt.Println("Monitoring for temperature controller changes...")
	duration, err := time.ParseDuration("1000ms")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse 1000ms as a duration")
	}

	ticker := time.NewTicker(duration)

	for {
		select {
		case <-ticker.C:
			for _, controller := range devices.AllTemperatureControllers() {
				if !controller.Running {
					log.Info().Msgf("Starting %v!\n", controller.Name)
					go controller.RunControl()
				}
			}
		case <-devices.Context.Done():
			ticker.Stop()
			return
		}
	}
}
