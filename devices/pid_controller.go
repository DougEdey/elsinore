package devices

import (
	"fmt"
	"log"
	"time"

	"periph.io/x/periph/conn/physic"

	"math/big"
)

var controllers = make(map[string]*PidController)

// PidController defines a mapping of temperature probes to their control settings
type PidController struct {
	LastReadings      map[time.Time]physic.Temperature
	TemperatureProbes []TemperatureProbe
	Proportional      big.Float
	Integral          big.Float
	Differential      big.Float
}

// FindPidControllerForProbe returns the pid controller associated with the TemperatureProbe
func FindPidControllerForProbe(physAddr string) *PidController {
	for _, controller := range controllers {
		for _, probe := range controller.TemperatureProbes {
			if probe.PhysAddr == physAddr {
				return controller
			}
		}
	}
	return nil
}

// FindPidControllerByName returns the pid controller with a specific name
func FindPidControllerByName(name string) *PidController {
	controller := controllers[name]
	if controller != nil {
		log.Printf("Found PID Controller for %v: %v", name, controller)
	}
	return controller
}

// CreatePidController Create a new PID controller for the Temperature probe
// name -> The name of the PID Controller
// probe -> The probe to associate with the controller
// If a PID controller exists with the same name, the probe will be added to it (or no-op if it is already assigned)
// If the passed in TemperatureProbe is associated with a different PID Controller, an error will be returned and no PID controller will be returned
func CreatePidController(name string, probe TemperatureProbe) (*PidController, error) {
	existingControllerForProbe := FindPidControllerForProbe(probe.PhysAddr)
	controller := FindPidControllerByName(name)

	if existingControllerForProbe != nil {
		if existingControllerForProbe == controller {
			return controller, nil
		}
		return nil, fmt.Errorf("PID Controller (%v) exists for this probe, trying removing it first", existingControllerForProbe)
	}

	if controller == nil {
		controller = &PidController{}
		controllers[name] = controller
	}

	controller.TemperatureProbes = append(controller.TemperatureProbes, probe)
	return controller, nil
}
