package devices_test

import (
	"testing"

	"github.com/dougedey/elsinore/devices"
	"periph.io/x/periph/conn/onewire"
)

func TestCreatePidController(t *testing.T) {
	probe := devices.TemperatureProbe{
		PhysAddr: "ARealAddress",
		Address:  onewire.Address(12345),
	}

	t.Run("A new Pid Controller is created if no existing device with the same name exists", func(t *testing.T) {
		pidController, err := devices.CreatePidController("sample", probe)
		if err != nil {
			t.Fatalf("Failed to create the Pid Controller: %v", err)
		}

		if pidController == nil {
			t.Fatalf("No Pid Controller returned for sample")
		}
	})

	t.Run("A Pid controller cannot be created if the probe is already associated with a controller", func(t *testing.T) {
		pidController, err := devices.CreatePidController("sample_2", probe)

		if err == nil {
			t.Fatalf("Created a duplicate PID controller: %v", pidController)
		}
	})

	t.Run("Re-adding a probe to the same pid controller is a no-op", func(t *testing.T) {
		pidController, err := devices.CreatePidController("sample", probe)

		if err != nil {
			t.Fatalf("Failed to do nothing: %v", err)
		}

		existingController := devices.FindPidControllerByName("sample")
		if pidController != existingController {
			t.Fatalf("Expected %v, but got %v", existingController, pidController)
		}
	})
}
