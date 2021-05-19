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
	"github.com/dougedey/elsinore/hardware"
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
		hardware.SetProbe(&hardware.TemperatureProbe{
			PhysAddr: realAddress,
			Address:  onewire.Address(12345),
		})
	}

	database.InitDatabase(dbName,
		&devices.TempProbeDetail{}, &devices.PidSettings{}, &devices.HysteriaSettings{},
		&devices.ManualSettings{}, &devices.TemperatureController{},
	)

	fmt.Println("Loaded and looking for temperatures")
	messages := make(chan string)
	go hardware.ReadTemperatures(messages)
	go hardware.LogTemperatures(messages)

	// fmt.Println("Starting")
	// out21 := devices.OutPin{Identifier: "GPIO21", FriendlyName: "GPIO21"}
	// o := devices.OutputControl{HeatOutput: &out21, DutyCycle: 50, CycleTime: 4}
	// go o.RunControl()

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
