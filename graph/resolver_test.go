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
	"github.com/dougedey/elsinore/graph/model"
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
			`[{"message":"no device found for address Invalid","path":["probe"]}]`,
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
			`[{"message":"no device(s) found for address(es): [Invalid]","path":["fetchProbes"]}]`,
			err.Error(),
		)
	})
}

func TestAssignProbeMutations(t *testing.T) {
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
			assignProbe(address: "ARealAddress", name: "TEST PROBE") {
				id
				name
			}
		}
		`, &assignResp)
		require.Equal(t, "1", assignResp.AssignProbe.ID)

		c.MustPost(`
		mutation {
			assignProbe(address: "RealAddress", name: "TEST PROBE 2") {
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

func TestUpdateProbeMutations(t *testing.T) {
	setupTestDb(t)
	c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}})))

	var updateResp struct {
		UpdateProbe struct {
			ID           string
			Name         string
			Mode         model.ControllerMode
			CoolSettings *struct {
				Id        *string
				CycleTime *int
				Delay     *int
			}
			HeatSettings *struct {
				Id        *string
				CycleTime *int
				Delay     *int
			}
			ManualSettings *struct {
				Id        *string
				CycleTime *int
				DutyCycle *int
			}
			HysteriaSettings *struct {
				Id         *string
				Configured *bool
				MaxTemp    *string
			}
		}
		Errors []struct {
			Message   string
			Locations []struct {
				Line   int
				Column int
			}
		}
	}

	t.Run("updateProbe with an invalid ID returns an error", func(t *testing.T) {
		err := c.Post(`
		mutation {
			updateProbe(id: "1", controllerSettings: {}) {
				id
				name
			}
		}
		`, &updateResp)

		require.Equal(t,
			`[{"message":"no controller could be found for: 1","path":["updateProbe"]}]`,
			err.Error(),
		)
	})

	realAddress := "ARealAddress"
	devices.CreateTemperatureController("Test", &devices.TemperatureProbe{
		PhysAddr: realAddress,
		Address:  onewire.Address(12345),
	})

	t.Run("updateProbe updates the name and makes no other changes", func(t *testing.T) {
		c.MustPost(`
		mutation {
			updateProbe(id: "1", controllerSettings: { name: "Updated name" }) {
				id
				name
			}
		}
		`, &updateResp)

		require.Equal(t, "1", updateResp.UpdateProbe.ID)
		require.Equal(t, "Updated name", updateResp.UpdateProbe.Name)
		require.Nil(t, updateResp.UpdateProbe.CoolSettings)
		require.Nil(t, updateResp.UpdateProbe.HeatSettings)
		require.Nil(t, updateResp.UpdateProbe.ManualSettings)
		require.Nil(t, updateResp.UpdateProbe.HysteriaSettings)
	})

	t.Run("updateProbe creates Cool settings", func(t *testing.T) {
		err := c.Post(`
		mutation {
			updateProbe(id: "1", controllerSettings: { name: "Updated name", coolSettings: { configured: true, cycleTime: 1 } }) {
				id
				name
				coolSettings {
					id
					cycleTime
					delay
				}
			}
		}
		`, &updateResp)

		require.Nil(t, err)
		require.Equal(t, "1", updateResp.UpdateProbe.ID)
		require.Equal(t, "Updated name", updateResp.UpdateProbe.Name)
		require.NotNil(t, updateResp.UpdateProbe.CoolSettings)
		require.Equal(t, "1", *updateResp.UpdateProbe.CoolSettings.Id)
		require.Equal(t, 1, *updateResp.UpdateProbe.CoolSettings.CycleTime)
		require.Equal(t, 0, *updateResp.UpdateProbe.CoolSettings.Delay)
		require.Nil(t, updateResp.UpdateProbe.HeatSettings)
		require.Nil(t, updateResp.UpdateProbe.ManualSettings)
		require.Nil(t, updateResp.UpdateProbe.HysteriaSettings)
	})

	t.Run("updateProbe creates Heat settings", func(t *testing.T) {
		err := c.Post(`
		mutation {
			updateProbe(id: "1", controllerSettings: { name: "Updated name", heatSettings: { configured: true, cycleTime: 4 } }) {
				id
				name
				coolSettings {
					id
					cycleTime
					delay
				}
				heatSettings {
					id
					cycleTime
					delay
				}
			}
		}
		`, &updateResp)

		require.Nil(t, err)
		require.Equal(t, "1", updateResp.UpdateProbe.ID)
		require.Equal(t, "Updated name", updateResp.UpdateProbe.Name)
		require.NotNil(t, updateResp.UpdateProbe.CoolSettings)
		require.Equal(t, "1", *updateResp.UpdateProbe.CoolSettings.Id)
		require.Equal(t, 1, *updateResp.UpdateProbe.CoolSettings.CycleTime)
		require.Equal(t, 0, *updateResp.UpdateProbe.CoolSettings.Delay)
		require.Equal(t, "2", *updateResp.UpdateProbe.HeatSettings.Id)
		require.Equal(t, 4, *updateResp.UpdateProbe.HeatSettings.CycleTime)
		require.Equal(t, 0, *updateResp.UpdateProbe.HeatSettings.Delay)
		require.Nil(t, updateResp.UpdateProbe.ManualSettings)
		require.Nil(t, updateResp.UpdateProbe.HysteriaSettings)
	})

	t.Run("updateProbe creates Manual settings", func(t *testing.T) {
		err := c.Post(`
		mutation {
			updateProbe(id: "1", controllerSettings: { name: "Updated name", manualSettings: { configured: true, cycleTime: 5, dutyCycle: 45 } }) {
				id
				name
				coolSettings {
					id
					cycleTime
					delay
				}
				heatSettings {
					id
					cycleTime
					delay
				}
				manualSettings {
					id
					cycleTime
					dutyCycle
				}
			}
		}
		`, &updateResp)

		require.Nil(t, err)
		require.Equal(t, "1", updateResp.UpdateProbe.ID)
		require.Equal(t, "Updated name", updateResp.UpdateProbe.Name)
		require.NotNil(t, updateResp.UpdateProbe.CoolSettings)
		require.Equal(t, "1", *updateResp.UpdateProbe.CoolSettings.Id)
		require.Equal(t, 1, *updateResp.UpdateProbe.CoolSettings.CycleTime)
		require.Equal(t, 0, *updateResp.UpdateProbe.CoolSettings.Delay)
		require.Equal(t, "2", *updateResp.UpdateProbe.HeatSettings.Id)
		require.Equal(t, 4, *updateResp.UpdateProbe.HeatSettings.CycleTime)
		require.Equal(t, 0, *updateResp.UpdateProbe.HeatSettings.Delay)
		require.Equal(t, "1", *updateResp.UpdateProbe.ManualSettings.Id)
		require.Equal(t, 5, *updateResp.UpdateProbe.ManualSettings.CycleTime)
		require.Equal(t, 45, *updateResp.UpdateProbe.ManualSettings.DutyCycle)
		require.Nil(t, updateResp.UpdateProbe.HysteriaSettings)
	})

	t.Run("updateProbe creates Manual settings", func(t *testing.T) {
		err := c.Post(`
		mutation {
			updateProbe(id: "1", controllerSettings: { name: "Updated name", hysteriaSettings: { configured: true, maxTemp: "103C" } }) {
				id
				name
				coolSettings {
					id
					cycleTime
					delay
				}
				heatSettings {
					id
					cycleTime
					delay
				}
				manualSettings {
					id
					cycleTime
					dutyCycle
				}
				hysteriaSettings {
					id
					maxTemp
				}
			}
		}
		`, &updateResp)

		require.Nil(t, err)
		require.Equal(t, "1", updateResp.UpdateProbe.ID)
		require.Equal(t, "Updated name", updateResp.UpdateProbe.Name)
		require.NotNil(t, updateResp.UpdateProbe.CoolSettings)
		require.Equal(t, "1", *updateResp.UpdateProbe.CoolSettings.Id)
		require.Equal(t, 1, *updateResp.UpdateProbe.CoolSettings.CycleTime)
		require.Equal(t, 0, *updateResp.UpdateProbe.CoolSettings.Delay)
		require.Equal(t, "2", *updateResp.UpdateProbe.HeatSettings.Id)
		require.Equal(t, 4, *updateResp.UpdateProbe.HeatSettings.CycleTime)
		require.Equal(t, 0, *updateResp.UpdateProbe.HeatSettings.Delay)
		require.Equal(t, "1", *updateResp.UpdateProbe.ManualSettings.Id)
		require.Equal(t, 5, *updateResp.UpdateProbe.ManualSettings.CycleTime)
		require.Equal(t, 45, *updateResp.UpdateProbe.ManualSettings.DutyCycle)
		require.Equal(t, "1", *updateResp.UpdateProbe.HysteriaSettings.Id)
		require.Equal(t, "103°C", *updateResp.UpdateProbe.HysteriaSettings.MaxTemp)
	})
}
