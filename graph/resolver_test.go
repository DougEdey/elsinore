package graph_test

import (
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

	var fetchProbesResp struct {
		FetchProbes []struct {
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
		ProbeList []struct {
			PhysAddr string
		}
	}

	t.Run("Invalid Probe address", func(t *testing.T) {
		err := c.Post(`
					query InvalidTemperatureProbe {
						probe(address: "Invalid") {
							physAddr
						}
					}
				`, &probeResp)

		require.Equal(t,
			`[{"message":"No device found for address Invalid","path":["probe"]}]`,
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

	t.Run("AllProbes lists all temperature probes", func(t *testing.T) {
		c.MustPost(`
					query AllProbes {
						probeList {
							physAddr
						}
					}
				`, &probeListResp)

		require.Equal(t, "ARealAddress", probeListResp.ProbeList[0].PhysAddr)
	})

	t.Run("fetchProbes returns the matching probes", func(t *testing.T) {
		c.MustPost(`
					query FetchValidProbes {
						fetchProbes(addresses: ["ARealAddress"]) {
							physAddr
						}
					}
				`, &fetchProbesResp)

		require.Equal(t, "ARealAddress", fetchProbesResp.FetchProbes[0].PhysAddr)
	})

	t.Run("fetchProbes returns the matching probes and errors for invalid ones", func(t *testing.T) {
		err := c.Post(`
					query FetchInvalidProbes {
						fetchProbes(addresses: ["ARealAddress", "Invalid"]) {
							physAddr
						}
					}
				`, &fetchProbesResp)

		require.Equal(t, "ARealAddress", fetchProbesResp.FetchProbes[0].PhysAddr)
		require.Equal(t,
			`[{"message":"No device(s) found for address(es): [Invalid]","path":["fetchProbes"]}]`,
			err.Error(),
		)
	})
}
