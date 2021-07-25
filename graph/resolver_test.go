package graph_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/dougedey/elsinore/database"
	"github.com/dougedey/elsinore/devices"
	"github.com/dougedey/elsinore/graph"
	"github.com/dougedey/elsinore/graph/generated"
	"github.com/dougedey/elsinore/graph/model"
	"github.com/dougedey/elsinore/hardware"
	"github.com/stretchr/testify/require"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/gpio/gpiotest"
	"periph.io/x/periph/conn/onewire"
)

func setupTestDb(t *testing.T) {
	dbName := "test"
	database.InitDatabase(&dbName,
		&devices.TempProbeDetail{}, &devices.PidSettings{}, &devices.HysteriaSettings{},
		&devices.ManualSettings{}, &devices.TemperatureController{}, &devices.Switch{},
	)
	devices.ClearControllers()

	t.Cleanup(func() {
		database.Close()
		e := os.Remove("test.db")
		if e != nil {
			t.Fatal(e)
		}
		devices.ClearControllers()
	})
}

func TestQuery(t *testing.T) {
	c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}})))
	realAddress := "ARealAddress"
	hardware.SetProbe(&hardware.TemperatureProbe{
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
	hardware.SetProbe(&hardware.TemperatureProbe{
		PhysAddr: realAddress,
		Address:  onewire.Address(12345),
	})
	aRealAddress := "RealAddress"
	hardware.SetProbe(&hardware.TemperatureProbe{
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
		RemoveProbeFromTemperatureController struct {
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

	t.Run("removeProbeFromTemperatureController removes a probe from the controller", func(t *testing.T) {
		c.MustPost(`
		mutation {
			removeProbeFromTemperatureController(address: "ARealAddress") {
				id
				name
			}
		}
		`, &removeResp)

		require.Equal(t, "1", removeResp.RemoveProbeFromTemperatureController.ID)
	})
}

func TestUpdateTemperatureControllerMutations(t *testing.T) {
	setupTestDb(t)
	c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}})))

	var updateResp struct {
		UpdateTemperatureController struct {
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

	t.Run("updateTemperatureController with an invalid ID returns an error", func(t *testing.T) {
		err := c.Post(`
		mutation {
			updateTemperatureController(controllerSettings: { id: "1" }) {
				id
				name
			}
		}
		`, &updateResp)

		require.NotNil(t, err)
		require.Equal(t,
			`[{"message":"no controller could be found for: 1","path":["updateTemperatureController"]}]`,
			err.Error(),
		)
	})

	realAddress := "ARealAddress"
	devices.CreateTemperatureController("Test", &devices.TempProbeDetail{
		PhysAddr: realAddress,
	})

	t.Run("updateTemperatureController updates the name and makes no other changes", func(t *testing.T) {
		c.MustPost(`
		mutation {
			updateTemperatureController(controllerSettings: { id: "1", name: "Updated name" }) {
				id
				name
			}
		}
		`, &updateResp)

		require.Equal(t, "1", updateResp.UpdateTemperatureController.ID)
		require.Equal(t, "Updated name", updateResp.UpdateTemperatureController.Name)
		require.Nil(t, updateResp.UpdateTemperatureController.CoolSettings)
		require.Nil(t, updateResp.UpdateTemperatureController.HeatSettings)
		require.Nil(t, updateResp.UpdateTemperatureController.ManualSettings)
		require.Nil(t, updateResp.UpdateTemperatureController.HysteriaSettings)
	})

	t.Run("updateTemperatureController creates Cool settings", func(t *testing.T) {
		err := c.Post(`
		mutation {
			updateTemperatureController(controllerSettings: { id: "1", name: "Updated name", coolSettings: { configured: true, cycleTime: 1 } }) {
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
		require.Equal(t, "1", updateResp.UpdateTemperatureController.ID)
		require.Equal(t, "Updated name", updateResp.UpdateTemperatureController.Name)
		require.NotNil(t, updateResp.UpdateTemperatureController.CoolSettings)
		require.Equal(t, "1", *updateResp.UpdateTemperatureController.CoolSettings.Id)
		require.Equal(t, 1, *updateResp.UpdateTemperatureController.CoolSettings.CycleTime)
		require.Equal(t, 0, *updateResp.UpdateTemperatureController.CoolSettings.Delay)
		require.Nil(t, updateResp.UpdateTemperatureController.HeatSettings)
		require.Nil(t, updateResp.UpdateTemperatureController.ManualSettings)
		require.Nil(t, updateResp.UpdateTemperatureController.HysteriaSettings)
	})

	t.Run("updateTemperatureController creates Heat settings", func(t *testing.T) {
		err := c.Post(`
		mutation {
			updateTemperatureController(controllerSettings: { id: "1", name: "Updated name", heatSettings: { configured: true, cycleTime: 4 } }) {
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
		require.Equal(t, "1", updateResp.UpdateTemperatureController.ID)
		require.Equal(t, "Updated name", updateResp.UpdateTemperatureController.Name)
		require.NotNil(t, updateResp.UpdateTemperatureController.CoolSettings)
		require.Equal(t, "1", *updateResp.UpdateTemperatureController.CoolSettings.Id)
		require.Equal(t, 1, *updateResp.UpdateTemperatureController.CoolSettings.CycleTime)
		require.Equal(t, 0, *updateResp.UpdateTemperatureController.CoolSettings.Delay)
		require.Equal(t, "2", *updateResp.UpdateTemperatureController.HeatSettings.Id)
		require.Equal(t, 4, *updateResp.UpdateTemperatureController.HeatSettings.CycleTime)
		require.Equal(t, 0, *updateResp.UpdateTemperatureController.HeatSettings.Delay)
		require.Nil(t, updateResp.UpdateTemperatureController.ManualSettings)
		require.Nil(t, updateResp.UpdateTemperatureController.HysteriaSettings)
	})

	t.Run("updateTemperatureController creates Manual settings", func(t *testing.T) {
		err := c.Post(`
		mutation {
			updateTemperatureController(controllerSettings: { id: "1",  name: "Updated name", manualSettings: { configured: true, cycleTime: 5, dutyCycle: 45 } }) {
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
		require.Equal(t, "1", updateResp.UpdateTemperatureController.ID)
		require.Equal(t, "Updated name", updateResp.UpdateTemperatureController.Name)
		require.NotNil(t, updateResp.UpdateTemperatureController.CoolSettings)
		require.Equal(t, "1", *updateResp.UpdateTemperatureController.CoolSettings.Id)
		require.Equal(t, 1, *updateResp.UpdateTemperatureController.CoolSettings.CycleTime)
		require.Equal(t, 0, *updateResp.UpdateTemperatureController.CoolSettings.Delay)
		require.Equal(t, "2", *updateResp.UpdateTemperatureController.HeatSettings.Id)
		require.Equal(t, 4, *updateResp.UpdateTemperatureController.HeatSettings.CycleTime)
		require.Equal(t, 0, *updateResp.UpdateTemperatureController.HeatSettings.Delay)
		require.Equal(t, "1", *updateResp.UpdateTemperatureController.ManualSettings.Id)
		require.Equal(t, 5, *updateResp.UpdateTemperatureController.ManualSettings.CycleTime)
		require.Equal(t, 45, *updateResp.UpdateTemperatureController.ManualSettings.DutyCycle)
		require.Nil(t, updateResp.UpdateTemperatureController.HysteriaSettings)
	})

	t.Run("updateTemperatureController creates Manual settings", func(t *testing.T) {
		err := c.Post(`
		mutation {
			updateTemperatureController(controllerSettings: { id: "1", name: "Updated name", hysteriaSettings: { configured: true, maxTemp: "103C" } }) {
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
		require.Equal(t, "1", updateResp.UpdateTemperatureController.ID)
		require.Equal(t, "Updated name", updateResp.UpdateTemperatureController.Name)
		require.NotNil(t, updateResp.UpdateTemperatureController.CoolSettings)
		require.Equal(t, "1", *updateResp.UpdateTemperatureController.CoolSettings.Id)
		require.Equal(t, 1, *updateResp.UpdateTemperatureController.CoolSettings.CycleTime)
		require.Equal(t, 0, *updateResp.UpdateTemperatureController.CoolSettings.Delay)
		require.Equal(t, "2", *updateResp.UpdateTemperatureController.HeatSettings.Id)
		require.Equal(t, 4, *updateResp.UpdateTemperatureController.HeatSettings.CycleTime)
		require.Equal(t, 0, *updateResp.UpdateTemperatureController.HeatSettings.Delay)
		require.Equal(t, "1", *updateResp.UpdateTemperatureController.ManualSettings.Id)
		require.Equal(t, 5, *updateResp.UpdateTemperatureController.ManualSettings.CycleTime)
		require.Equal(t, 45, *updateResp.UpdateTemperatureController.ManualSettings.DutyCycle)
		require.Equal(t, "1", *updateResp.UpdateTemperatureController.HysteriaSettings.Id)
		require.Equal(t, "103°C", *updateResp.UpdateTemperatureController.HysteriaSettings.MaxTemp)
	})
}

func TestDeleteTemperatureControllerMutations(t *testing.T) {
	setupTestDb(t)
	c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}})))

	var deleteRespFail struct {
		DeleteTemperatureController struct {
			ID                string
			TemperatureProbes []string
		}
		Errors []struct {
			Message   string
			Locations []struct {
				Line   int
				Column int
			}
		}
	}

	t.Run("deleteTemperature with an invalid ID returns an error", func(t *testing.T) {
		err := c.Post(`
		mutation {
			deleteTemperatureController(id: "1") {
				id
				temperatureProbes
			}
		}
		`, &deleteRespFail)

		require.Equal(t,
			`[{"message":"failed to find a controller to delete for: 1","path":["deleteTemperatureController"]}]`,
			err.Error(),
		)
	})

	realAddress := "ARealAddress"
	controller, err := devices.CreateTemperatureController("Test", &devices.TempProbeDetail{
		PhysAddr: realAddress,
	})

	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("Inserted %v", controller.ID)
	var deleteResp struct {
		DeleteTemperatureController struct {
			ID                string
			TemperatureProbes []string
		}
		Errors []struct {
			Message   string
			Locations []struct {
				Line   int
				Column int
			}
		}
	}

	t.Run("deleteTemperature with a valid ID returns the list of addresses", func(t *testing.T) {
		c.MustPost(fmt.Sprintf(`
		mutation {
			deleteTemperatureController(id: "%v") {
				id
				temperatureProbes
			}
		}
		`, controller.ID), &deleteResp)

		// require.Nil(t, err.Error())
		require.Equal(t, "1", deleteResp.DeleteTemperatureController.ID)
		require.Equal(t, 1, len(deleteResp.DeleteTemperatureController.TemperatureProbes))
	})
}

func TestSystemSettings(t *testing.T) {
	setupTestDb(t)
	c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}})))

	var settingsResp struct {
		Settings struct {
			BreweryName string
		}
		Errors []struct {
			Message   string
			Locations []struct {
				Line   int
				Column int
			}
		}
	}

	var updateSettingsResp struct {
		UpdateSettings struct {
			BreweryName string
		}
		Errors []struct {
			Message   string
			Locations []struct {
				Line   int
				Column int
			}
		}
	}

	t.Run("Name defaults to empty", func(t *testing.T) {
		c.MustPost(`
					query {
						settings {
							breweryName
						}
					}
				`, &settingsResp)

		require.Equal(t,
			``,
			settingsResp.Settings.BreweryName,
		)
	})

	t.Run("Name can be updated", func(t *testing.T) {
		c.MustPost(`
					mutation {
						updateSettings(settings: {breweryName: "Some Name"}) {
							breweryName
						}
					}
				`, &updateSettingsResp)

		require.Equal(t,
			`Some Name`,
			updateSettingsResp.UpdateSettings.BreweryName,
		)
	})

	t.Run("No Name does not change anything", func(t *testing.T) {
		c.MustPost(`
					mutation {
						updateSettings(settings: {}) {
							breweryName
						}
					}
				`, &updateSettingsResp)

		require.Equal(t,
			`Some Name`,
			updateSettingsResp.UpdateSettings.BreweryName,
		)
	})
}

func TestSwitches(t *testing.T) {
	setupTestDb(t)
	c := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}})))

	var switchesListResp struct {
		Switches []struct {
			Id   string
			Name string
			Gpio string
		}
		Errors []struct {
			Message   string
			Locations []struct {
				Line   int
				Column int
			}
		}
	}

	var switchResp struct {
		ModifySwitch struct {
			Id    string
			Name  string
			Gpio  string
			State string
		}
		Errors []struct {
			Message   string
			Locations []struct {
				Line   int
				Column int
			}
		}
	}

	t.Run(("Empty switch list is returned by default"), func(t *testing.T) {
		c.MustPost(`
			query {
				switches {
					id
					name
					gpio
				}
			}
		`, &switchesListResp)

		require.Empty(t, switchesListResp.Switches, "There should not be any switches configured")
	})

	gpioreg.Register(&gpiotest.Pin{N: "GPIO1", Num: 10, Fn: "I2C1_SDA"})
	gpioreg.Register(&gpiotest.Pin{N: "GPIO2", Num: 11, Fn: "I2C1_SDC"})

	t.Run("Can create a new switch", func(t *testing.T) {
		c.MustPost(`
			mutation {
				modifySwitch(switchSettings: {name: "Test", gpio: "GPIO1"} ) {
					id
					name
					gpio
					state
				}
			}
		`, &switchResp)

		require.Empty(t, switchResp.Errors, "no Errors should be returned")
		require.Equal(t, "Test", switchResp.ModifySwitch.Name)
		require.Equal(t, "GPIO1", switchResp.ModifySwitch.Gpio)
		require.Equal(t, "off", switchResp.ModifySwitch.State)
	})

	t.Run("Cannot create a new switch with a GPIO in use already", func(t *testing.T) {
		err := c.Post(`
			mutation {
				modifySwitch(switchSettings: {name: "Test2", gpio: "GPIO1"} ) {
					id
					name
					gpio
				}
			}
		`, &switchResp)

		require.NotNil(t, err)
		require.Equal(t,
			`[{"message":"GPIO 'GPIO1' is already in use","path":["modifySwitch"]}]`,
			err.Error(),
		)
	})

	t.Run("Cannot create a new switch with an existing name", func(t *testing.T) {
		err := c.Post(`
			mutation {
				modifySwitch(switchSettings: {name: "Test", gpio: "GPIO2"} ) {
					id
					name
					gpio
				}
			}
		`, &switchResp)

		require.NotNil(t, err)
		require.Equal(t,
			`[{"message":"switch 'Test' already exists","path":["modifySwitch"]}]`,
			err.Error(),
		)
	})

	t.Run(("The switch list is returned when switches are configured"), func(t *testing.T) {
		c.MustPost(`
			query {
				switches {
					id
					name
					gpio
				}
			}
		`, &switchesListResp)

		require.Len(t, switchesListResp.Switches, 1, "One switch should exist")
	})

	t.Run("Can update an existing switch", func(t *testing.T) {
		c.MustPost(`
			mutation {
				modifySwitch(switchSettings: {id: 1, name: "TestUpdate", gpio: "GPIO2"} ) {
					id
					name
					gpio
				}
			}
		`, &switchResp)

		require.Empty(t, switchResp.Errors, "no Errors should be returned")
		require.Equal(t, "TestUpdate", switchResp.ModifySwitch.Name)
		require.Equal(t, "GPIO2", switchResp.ModifySwitch.Gpio)
	})

	var toggleSwitchResp struct {
		ToggleSwitch struct {
			Id    string
			Name  string
			Gpio  string
			State string
		}
		Errors []struct {
			Message   string
			Locations []struct {
				Line   int
				Column int
			}
		}
	}

	t.Run("Can turn a switch on", func(t *testing.T) {
		c.MustPost(`
			mutation {
				toggleSwitch(id: 1, mode: on ) {
					id
					name
					gpio
					state
				}
			}
		`, &toggleSwitchResp)

		require.Empty(t, toggleSwitchResp.Errors, "no Errors should be returned")
		require.Equal(t, "TestUpdate", toggleSwitchResp.ToggleSwitch.Name)
		require.Equal(t, "GPIO2", toggleSwitchResp.ToggleSwitch.Gpio)
		require.Equal(t, "on", toggleSwitchResp.ToggleSwitch.State)
	})

	t.Run("Can turn a switch off", func(t *testing.T) {
		c.MustPost(`
			mutation {
				toggleSwitch(id: 1, mode: off ) {
					id
					name
					gpio
					state
				}
			}
		`, &toggleSwitchResp)

		require.Empty(t, toggleSwitchResp.Errors, "no Errors should be returned")
		require.Equal(t, "TestUpdate", toggleSwitchResp.ToggleSwitch.Name)
		require.Equal(t, "GPIO2", toggleSwitchResp.ToggleSwitch.Gpio)
		require.Equal(t, "off", toggleSwitchResp.ToggleSwitch.State)
	})

	var deleteSwitchResp struct {
		DeleteSwitch struct {
			Id    string
			Name  string
			Gpio  string
			State string
		}
		Errors []struct {
			Message   string
			Locations []struct {
				Line   int
				Column int
			}
		}
	}
	t.Run("Can delete a switch", func(t *testing.T) {
		c.MustPost(`
			mutation {
				deleteSwitch(id: 1) {
					id
					name
					gpio
					state
				}
			}
		`, &deleteSwitchResp)

		require.Empty(t, toggleSwitchResp.Errors, "no Errors should be returned")
		require.Equal(t, "1", deleteSwitchResp.DeleteSwitch.Id)
		require.Equal(t, "TestUpdate", deleteSwitchResp.DeleteSwitch.Name)
		require.Equal(t, "GPIO2", deleteSwitchResp.DeleteSwitch.Gpio)
		require.Equal(t, "off", deleteSwitchResp.DeleteSwitch.State)
	})

	t.Run("Can reuse a deleted switches GPIO", func(t *testing.T) {
		c.MustPost(`
			mutation {
				modifySwitch(switchSettings: {name: "Test2", gpio: "GPIO1"} ) {
					id
					name
					gpio
				}
			}
		`, &switchResp)

		require.Empty(t, switchResp.Errors, "no Errors should be returned")
		require.Equal(t, "Test2", switchResp.ModifySwitch.Name)
		require.Equal(t, "GPIO1", switchResp.ModifySwitch.Gpio)
		require.Equal(t, "off", switchResp.ModifySwitch.State)
	})
}
