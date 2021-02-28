package api_test

import (
	"testing"

	"github.com/dougedey/elsinore/api"
	"github.com/dougedey/elsinore/devices"
	"periph.io/x/periph/conn/onewire"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/gqlerrors"
	"github.com/graphql-go/graphql/language/location"
	"github.com/graphql-go/graphql/testutil"
)

type T struct {
	Query     string
	Schema    graphql.Schema
	Expected  *graphql.Result
	Variables map[string]interface{}
}

var Tests = []T{}

func init() {
	realAddress := "ARealAddress"
	devices.SetProbe(&devices.TemperatureProbe{
		PhysAddr: realAddress,
		Address:  onewire.Address(12345),
	},
	)

	Tests = []T{
		{
			Query: `
				query InvalidTemperatureProbe {
					probe(address: "Invalid") {
						id
					}
				}
			`,
			Schema: api.Schema,
			Expected: &graphql.Result{
				Data: map[string]interface{}{
					"probe": nil,
				},
				Errors: []gqlerrors.FormattedError{{
					Message: "No device found for address Invalid",
					Locations: []location.SourceLocation{
						{
							Line: 3, Column: 6,
						},
					},
					Path: []interface{}{
						"probe",
					},
				}},
			},
		},
		{
			Query: `
				query ValidTemperatureProbe {
					probe(address: "ARealAddress") {
						id
					}
				}
			`,
			Schema: api.Schema,
			Expected: &graphql.Result{
				Data: map[string]interface{}{
					"probe": map[string]interface{}{
						"id": "VGVtcGVyYXR1cmU6",
					},
				},
				Errors: nil,
			},
		},
		{
			Query: `
				query AllProbes {
					probeList
				}
			`,
			Schema: api.Schema,
			Expected: &graphql.Result{
				Data: map[string]interface{}{
					"probeList": []interface{}{
						"ARealAddress",
					},
				},
				Errors: nil,
			},
		},
	}
}

func TestQuery(t *testing.T) {
	for _, test := range Tests {
		params := graphql.Params{
			Schema:         test.Schema,
			RequestString:  test.Query,
			VariableValues: test.Variables,
		}
		testGraphql(test, params, t)
	}
}

func testGraphql(test T, p graphql.Params, t *testing.T) {
	result := graphql.Do(p)

	if len(result.Errors) > 0 && test.Expected.Errors == nil {
		t.Fatalf("wrong result, unexpected errors: %v", result.Errors)
	}

	if !testutil.EqualResults(result, test.Expected) {
		t.Fatalf("wrong result, query: %v, graphql result diff: %v", test.Query, testutil.Diff(test.Expected, result))
	}
}
