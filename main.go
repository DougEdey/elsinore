package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/dougedey/elsinore/database"
	"github.com/dougedey/elsinore/devices"
	"github.com/dougedey/elsinore/graph"
	"github.com/dougedey/elsinore/graph/generated"
	"periph.io/x/periph/conn/onewire"

	"net/http"
)

func main() {
	portPtr := flag.String("port", "8080", "The port to listen on")
	graphiqlFlag := flag.Bool("graphiql", true, "Disable GraphiQL web UI")
	dbName := flag.String("db_name", "elsinore", "The path/name of the local database")
	testDeviceFlag := flag.Bool("test_device", false, "Create a test device")
	flag.Parse()

	if *testDeviceFlag {
		realAddress := "ARealAddress"
		devices.SetProbe(&devices.TemperatureProbe{
			PhysAddr: realAddress,
			Address:  onewire.Address(12345),
		})
	}

	database.InitDatabase(dbName,
		&devices.TemperatureProbe{}, &devices.PidSettings{}, &devices.HysteriaSettings{},
		&devices.ManualSettings{}, &devices.TemperatureController{},
	)

	fmt.Println("Loaded and looking for temperatures")
	messages := make(chan string)
	go devices.ReadTemperatures(messages)
	go devices.LogTemperatures(messages)

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}}))

	if *graphiqlFlag {
		http.Handle("/", playground.Handler("GraphQL playground", "/query"))
		log.Printf("connect to http://localhost:%s/ for GraphQL playground", *portPtr)
	}
	http.Handle("/query", srv)

	log.Fatal(http.ListenAndServe(":"+*portPtr, nil))

	fmt.Printf("Server on %v\n", *portPtr)
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
}
