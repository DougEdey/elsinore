package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/dougedey/elsinore/github.com/dougedey/elsinore"
	"github.com/graphql-go/handler"

	"net/http"
)

func main() {
	portPtr := flag.String("port", "8080", "The port to listen on")
	graphiqlFlag := flag.Bool("graphiql", true, "Disable GraphiQL web UI")
	flag.Parse()

	fmt.Println("Loaded and looking for temperatures")
	messages := make(chan string)
	go elsinore.ReadTemperatures(messages)
	go elsinore.LogTemperatures(messages)

	http.Handle("/graphql", handler.New(
		&handler.Config{
			Schema: &elsinore.Schema,
			Pretty: true,
		}),
	)

	if *graphiqlFlag {
		http.Handle("/graphiql", handler.New(
			&handler.Config{
				Schema:   &elsinore.Schema,
				GraphiQL: true,
				Pretty:   true,
			}),
		)
	}

	fmt.Printf("Server on %v\n", *portPtr)
	fullPort := fmt.Sprintf(":%v", *portPtr)

	name, err := os.Hostname()
	if err != nil {
		fmt.Printf("Failed to get hostname: %v\n", err)
	} else {
		fmt.Printf("API Listening on: http://%v%v/graphql \n", name, fullPort)
		if *graphiqlFlag {
			fmt.Printf("GraphiQL interface: http://%v%v/graphiql \n", name, fullPort)
		}
	}
	log.Fatal(http.ListenAndServe(fullPort, nil))
}
