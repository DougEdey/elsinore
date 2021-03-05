package devices_test

import (
	"testing"

	"github.com/dougedey/elsinore/devices"
	"periph.io/x/periph/conn/onewire"
)

func TestCreateTemperatureController(t *testing.T) {
	probe := devices.TemperatureProbe{
		PhysAddr: "ARealAddress",
		Address:  onewire.Address(12345),
	}

	t.Run("A new Temperature controller is created if no existing device with the same name exists", func(t *testing.T) {
		temperatureController, err := devices.CreateTemperatureController("sample", probe)
		if err != nil {
			t.Fatalf("Failed to create the Temperature Controller: %v", err)
		}

		if temperatureController == nil {
			t.Fatalf("No Pid Controller returned for sample")
		}

		if temperatureController.Name != "sample" {
			t.Fatalf("Expected the temperature controller to be called sample, but got %v", temperatureController.Name)
		}
	})

	t.Run("A Temperature controller cannot be created if the probe is already associated with a controller", func(t *testing.T) {
		temperatureController, err := devices.CreateTemperatureController("sample_2", probe)

		if err == nil {
			t.Fatalf("Created a duplicate Temperature controller: %v", temperatureController)
		}
	})

	t.Run("Re-adding a probe to the same temperature controller is a no-op", func(t *testing.T) {
		temperatureController, err := devices.CreateTemperatureController("sample", probe)

		if err != nil {
			t.Fatalf("Failed to do nothing: %v", err)
		}

		existingController := devices.FindTemperatureControllerByName("sample")
		if temperatureController != existingController {
			t.Fatalf("Expected %v, but got %v", existingController, temperatureController)
		}
	})
}
