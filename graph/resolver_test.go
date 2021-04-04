package graph_test

import (
	"regexp"
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/dougedey/elsinore/devices"
	"github.com/dougedey/elsinore/graph"
	"github.com/dougedey/elsinore/graph/generated"
	"github.com/stretchr/testify/require"
	"periph.io/x/periph/conn/onewire"
)

func TestQuery(t *testing.T) {
	c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}})))
	realAddress := "ARealAddress"
	devices.SetProbe(&devices.TemperatureProbe{
		PhysAddr: realAddress,
		Address:  onewire.Address(12345),
	},
	)
	var probeResp struct {
		Probe struct {
			PhysAddr string
		}
		Errors []struct {
			Message   string
			Locations []struct {
				Line   int
				Column int
			}
		}
	}

	var probeListResp struct {
		ProbeList []*string
	}

	t.Run("Invalid Probe address", func(t *testing.T) {
		err := c.Post(`
					query InvalidTemperatureProbe {
						probe(address: "Invalid") {
							physAddr
						}
					}
				`, &probeResp)

		require.Regexp(t,
			regexp.MustCompile("\\[{\"message\":\"No device found for address .+\",\"path\":\\[\"probe\"\\]}\\]"),
			err.Error(),
		)
	})

	t.Run("Valid Probe address", func(t *testing.T) {
		c.MustPost(`
					query ValidTemperatureProbe {
						probe(address: "ARealAddress") {
							physAddr
						}
					}
				`, &probeResp)

		require.Equal(t, "ARealAddress", probeResp.Probe.PhysAddr)
	})

	t.Run("AllProbes lists Probe address", func(t *testing.T) {
		c.MustPost(`
					query AllProbes {
						probeList
					}
				`, &probeListResp)

		require.Equal(t, "ARealAddress", *probeListResp.ProbeList[0])
	})
}
