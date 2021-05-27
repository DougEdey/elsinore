package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/dougedey/elsinore/database"
	"github.com/dougedey/elsinore/devices"
	"github.com/dougedey/elsinore/graph"
	"github.com/dougedey/elsinore/graph/generated"
	"github.com/dougedey/elsinore/hardware"
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
	flag.Parse()

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
		&devices.ManualSettings{}, &devices.TemperatureController{},
	)

	_, err := host.Init()
	if err != nil {
		log.Fatalf("failed to initialize periph: %v", err)
	}

	fmt.Println("Loaded and looking for temperatures")
	messages := make(chan string)
	go hardware.ReadTemperatures(messages, quit)

	httpServerExitDone := &sync.WaitGroup{}

	httpServerExitDone.Add(1)
	srv := startHTTPServer(portPtr, graphiqlFlag, httpServerExitDone)

	shutdown.Add(func() {
		devices.CancelFunc()
		shutdownErr := srv.Shutdown(devices.Context)
		if shutdownErr != nil {
			log.Println(shutdownErr)
		}
	})

	shutdown.Listen()
}
func startHTTPServer(portPtr *string, graphiqlFlag *bool, wg *sync.WaitGroup) *http.Server {
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}}))
	httpSrv := &http.Server{Addr: ":" + *portPtr}

	if *graphiqlFlag {
		http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	}
	http.Handle("/query", srv)

	go func() {
		defer wg.Done()
		log.Fatal(httpSrv.ListenAndServe())
	}()

	fullPort := fmt.Sprintf(":%v", *portPtr)

	name, err := os.Hostname()
	if err != nil {
		fmt.Printf("Failed to get hostname: %v\n", err)
	} else {
		fmt.Printf("CORS API Listening on: http://%v%v/graphql \n", name, fullPort)
		if *graphiqlFlag {
			fmt.Printf("GraphiQL interface: http://%v%v/graphiql \n", name, fullPort)
		}
	}
	return httpSrv
}
