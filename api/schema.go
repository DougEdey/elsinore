package api

import (
	"fmt"

	"github.com/dougedey/elsinore/devices"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"
)

var temperatureProbe *graphql.Object

// Schema is the generated GraphQL schema
var Schema graphql.Schema

func init() {
	// Define the basic temperature probe
	temperatureProbe = graphql.NewObject(graphql.ObjectConfig{
		Name:        "TemperatureProbe",
		Description: "A device that reads a temperature",
		Fields: graphql.Fields{
			"id": relay.GlobalIDField("Temperature", nil),
			"reading": &graphql.Field{
				Type:        graphql.String,
				Description: "The value of the reading",
			},
			"physAddr": &graphql.Field{
				Type:        graphql.String,
				Description: "The physical address of this probe",
			},
		},
	})

	/**
	* This is the type that will be the root of our query, and the entry point to the Schema

	* This implements
	* type Query {
	* 	probe: TemperatureProbe
	* 	node(id: String!): Node
	* }
	 */

	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"probe": &graphql.Field{
				Type: temperatureProbe,
				Args: graphql.FieldConfigArgument{
					"address": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if id, ok := p.Args["address"].(string); ok {
						device := devices.GetTemperature(id)
						if device != nil {
							return device, nil
						}
					}
					return nil, fmt.Errorf("No device found for address %v", p.Args["address"])
				},
			},
			"probeList": &graphql.Field{
				Type:        graphql.NewList(graphql.String),
				Description: "Get the list of device addresses",
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					return devices.GetAddresses(), nil
				},
			},
		},
	})

	// Construct the schema
	var err error
	Schema, err = graphql.NewSchema(graphql.SchemaConfig{
		Query: queryType,
	})

	if err != nil {
		panic(err)
	}
}
