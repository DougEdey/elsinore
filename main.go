package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/dougedey/elsinore/api"
	"github.com/dougedey/elsinore/devices"

	"github.com/graphql-go/handler"

	"net/http"
)

func main() {
	portPtr := flag.String("port", "8080", "The port to listen on")
	graphiqlFlag := flag.Bool("graphiql", true, "Disable GraphiQL web UI")
	flag.Parse()

	fmt.Println("Loaded and looking for temperatures")
	messages := make(chan string)
	go devices.ReadTemperatures(messages)
	go devices.LogTemperatures(messages)

	http.Handle("/graphql", CorsMiddleware(handler.New(
		&handler.Config{
			Schema: &api.Schema,
			Pretty: true,
		}),
	))

	if *graphiqlFlag {
		http.Handle("/graphiql", handler.New(
			&handler.Config{
				Schema:   &api.Schema,
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
		fmt.Printf("CORS API Listening on: http://%v%v/graphql \n", name, fullPort)
		if *graphiqlFlag {
			fmt.Printf("GraphiQL interface: http://%v%v/graphiql \n", name, fullPort)
		}
	}
	log.Fatal(http.ListenAndServe(fullPort, nil))
}

// CorsMiddleware used to allow all origins to access the GraphQL API
func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// allow cross domain AJAX requests
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST,OPTIONS")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
