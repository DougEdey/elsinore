package api

import (
	"fmt"

	"github.com/dougedey/elsinore/devices"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"
)

var temperatureProbe *graphql.Object
var pidSettings *graphql.Object
var hysteriaSettings *graphql.Object
var manualSettings *graphql.Object
var temperatureController *graphql.Object
var probeSettings *graphql.InputObject
var modeEnum *graphql.Enum

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

	pidSettings = graphql.NewObject(graphql.ObjectConfig{
		Name:        "PIDSettings",
		Description: "The settings for heating or cooling on a temperature controller",
		Fields: graphql.Fields{
			"id": relay.GlobalIDField("PidSettings", nil),
			"proportional": &graphql.Field{
				Type:        graphql.Float,
				Description: "The proportional calculation value",
			},
			"integral": &graphql.Field{
				Type:        graphql.Float,
				Description: "The integral calculation value",
			},
			"differential": &graphql.Field{
				Type:        graphql.Float,
				Description: "The differental calculation value",
			},
			"cycleTime": &graphql.Field{
				Type:        graphql.Int,
				Description: "The automatic cycle time in seconds",
			},
			"delay": &graphql.Field{
				Type:        graphql.Int,
				Description: "The minimum delay between turning an output on and off in seconds",
			},
			"configured": &graphql.Field{
				Type:        graphql.Boolean,
				Description: "Indicates if these settings have been configured yet",
			},
		},
	})

	hysteriaSettings = graphql.NewObject(graphql.ObjectConfig{
		Name:        "HysteriaSettings",
		Description: "The settings for hysteria mode",
		Fields: graphql.Fields{
			"id": relay.GlobalIDField("HysteriaSettings", nil),
			"maxTemp": &graphql.Field{
				Type:        graphql.Float,
				Description: "When this temperature is hit, turn on the cooling output",
			},
			"minTemp": &graphql.Field{
				Type:        graphql.Float,
				Description: "When this temperature is hit, turn on the heating output",
			},
			"minTime": &graphql.Field{
				Type:        graphql.Int,
				Description: "The minimum amount of time to turn the outputs on for.",
			},
			"configured": &graphql.Field{
				Type:        graphql.Boolean,
				Description: "Indicates if these settings have been configrued yet.",
			},
		},
	})

	manualSettings = graphql.NewObject(graphql.ObjectConfig{
		Name:        "ManualSettings",
		Description: "The manual settings for this controller",
		Fields: graphql.Fields{
			"id": relay.GlobalIDField("ManualSettings", nil),
			"dutyCycle": &graphql.Field{
				Type:        graphql.Int,
				Description: "The manual duty cycle percentage for this controller",
			},
			"cycleTime": &graphql.Field{
				Type:        graphql.Int,
				Description: "The time for one duty cycle in seconds",
			},
			"configured": &graphql.Field{
				Type:        graphql.Boolean,
				Description: "Indicates if these settings have been configrued yet.",
			},
		},
	})

	probeSettings = graphql.NewInputObject(graphql.InputObjectConfig{
		Name:        "probeSettings",
		Description: "Used to configure a probe to a controller",
		Fields: graphql.InputObjectConfigFieldMap{
			"address": &graphql.InputObjectFieldConfig{
				Type:        graphql.String,
				Description: "The address of the probe to add to a controller",
			},
			"name": &graphql.InputObjectFieldConfig{
				Type:        graphql.String,
				Description: "The name of the controller to add the probe to.",
			},
		},
	})

	modeEnum = graphql.NewEnum(graphql.EnumConfig{
		Name:        "ControllerMode",
		Description: "The temperature controller mode",
		Values: graphql.EnumValueConfigMap{
			"manual": &graphql.EnumValueConfig{
				Description: "Use the manual settings",
			},
			"off": &graphql.EnumValueConfig{
				Description: "This controller is off",
			},
			"auto": &graphql.EnumValueConfig{
				Description: "Use the PID settings",
			},
			"hysteria": &graphql.EnumValueConfig{
				Description: "Use the hysteria settings",
			},
		},
	})

	temperatureController = graphql.NewObject(graphql.ObjectConfig{
		Name: "TemperatureController",
		Fields: graphql.Fields{
			"id": relay.GlobalIDField("TemperatureController", nil),
			"name": &graphql.Field{
				Type:        graphql.String,
				Description: "The assigned name of this controller",
			},
			"temperatureProbes": &graphql.Field{
				Type:        graphql.NewList(temperatureProbe),
				Description: "The probes assigned to this controller",
			},
			"coolSettings": &graphql.Field{
				Type:        pidSettings,
				Description: "The cooling settings for this controller",
			},
			"heatSettings": &graphql.Field{
				Type:        pidSettings,
				Description: "The heating settings for this controller",
			},
			"hysteriaSettings": &graphql.Field{
				Type:        hysteriaSettings,
				Description: "The hysteria mode settings for this controller",
			},
			"manualSettings": &graphql.Field{
				Type:        manualSettings,
				Description: "The manual settings for this controller",
			},
			"mode": &graphql.Field{
				Type:        modeEnum,
				Description: "The controller mode",
			},
			"dutyCycle": &graphql.Field{
				Type:        graphql.Int,
				Description: "The percentage of time this controller is on",
			},
			"calculatedDuty": &graphql.Field{
				Type:        graphql.Int,
				Description: "The PID calculated duty cycle, this can be overriden by the ManualDuty in manual mode",
			},
			"setPoint": &graphql.Field{
				Type:        graphql.Float,
				Description: "The target temperature when in auto mode",
			},
			"previousCalculationTime": &graphql.Field{
				Type:        graphql.DateTime,
				Description: "The last time that the duty cycle was calculated",
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

	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"assignProbe": &graphql.Field{
				Type: temperatureController,
				Args: graphql.FieldConfigArgument{
					"settings": &graphql.ArgumentConfig{
						Type: probeSettings,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					input, ok := p.Args["settings"].(map[string]interface{})
					if !ok {
						return nil, fmt.Errorf("Failed to parse input")
					}

					probe := devices.GetTemperature(input["address"].(string))
					if probe != nil {
						return devices.CreateTemperatureController(input["name"].(string), probe)
					}
					return nil, fmt.Errorf("Coud not find a probe for %v", input["address"])
				},
			},
		},
	})

	// Construct the schema
	var err error
	Schema, err = graphql.NewSchema(graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType,
	})

	if err != nil {
		panic(err)
	}
}
