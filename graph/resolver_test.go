package graph_test

import (
	"os"
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/dougedey/elsinore/database"
	"github.com/dougedey/elsinore/devices"
	"github.com/dougedey/elsinore/graph"
	"github.com/dougedey/elsinore/graph/generated"
	"github.com/stretchr/testify/require"
	"periph.io/x/periph/conn/onewire"
)

func setupTestDb(t *testing.T) {
	dbName := "test"
	database.InitDatabase(&dbName,
		&devices.TemperatureProbe{}, &devices.PidSettings{}, &devices.HysteriaSettings{},
		&devices.ManualSettings{}, &devices.TemperatureController{},
	)

	t.Cleanup(func() {
		database.Close()
		e := os.Remove("test.db")
		if e != nil {
			t.Fatal(e)
		}
	})
}

func TestQuery(t *testing.T) {
	c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}})))
	realAddress := "ARealAddress"
	devices.SetProbe(&devices.TemperatureProbe{
		PhysAddr: realAddress,
		Address:  onewire.Address(12345),
	})

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

func TestMutations(t *testing.T) {
	setupTestDb(t)

	c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}})))
	realAddress := "ARealAddress"
	devices.SetProbe(&devices.TemperatureProbe{
		PhysAddr: realAddress,
		Address:  onewire.Address(12345),
	})
	aRealAddress := "RealAddress"
	devices.SetProbe(&devices.TemperatureProbe{
		PhysAddr: aRealAddress,
		Address:  onewire.Address(123456),
	})
	var assignResp struct {
		AssignProbe struct {
			ID   string
			Name string
		}
		Errors []struct {
			Message   string
			Locations []struct {
				Line   int
				Column int
			}
		}
	}
	var assignRespTwo struct {
		AssignProbe struct {
			ID   string
			Name string
		}
		Errors []struct {
			Message   string
			Locations []struct {
				Line   int
				Column int
			}
		}
	}

	t.Run("assignProbe saves to the DB", func(t *testing.T) {
		c.MustPost(`
		mutation {
			assignProbe(settings: { address: "ARealAddress", name: "TEST PROBE"}) {
				id
				name
			}
		}
		`, &assignResp)
		require.Equal(t, "1", assignResp.AssignProbe.ID)

		c.MustPost(`
		mutation {
			assignProbe(settings: { address: "RealAddress", name: "TEST PROBE 2"}) {
				id
				name
			}
		}
		`, &assignRespTwo)

		require.Equal(t, "2", assignRespTwo.AssignProbe.ID)
	})

	var removeResp struct {
		RemoveProbeFromController struct {
			ID   string
			Name string
		}
		Errors []struct {
			Message   string
			Locations []struct {
				Line   int
				Column int
			}
		}
	}

	t.Run("removeProbeFromController removes a probe from the controller", func(t *testing.T) {
		c.MustPost(`
		mutation {
			removeProbeFromController(address: "ARealAddress") {
				id
				name
			}
		}
		`, &removeResp)

		require.Equal(t, "1", removeResp.RemoveProbeFromController.ID)
	})
}
