package main

import (
	"fmt"
	"log"

	"github.com/dougedey/elsinore/github.com/dougedey/elsinore"
	"github.com/graphql-go/handler"

	"net/http"
)


func main() {
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

	http.Handle("/graphiql", handler.New(
		&handler.Config{
			Schema: &elsinore.Schema,
			GraphiQL: true,
			Pretty: true,
		}),
	)
	
	fmt.Println("Server on 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}