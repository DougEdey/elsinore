package devices

import (
	"fmt"
	"log"
	"time"

	"periph.io/x/periph/conn/physic"

	"math/big"
)

var controllers = make(map[string]*TemperatureController)

// TemperatureController defines a mapping of temperature probes to their control settings
type TemperatureController struct {
	Name              string
	LastReadings      map[time.Time]physic.Temperature
	TemperatureProbes []*TemperatureProbe
	CoolSettings      PidSettings
	HeatSettings      PidSettings
	HysteriaSettings  HysteriaSettings
	Mode              string // Mode of this controller
	DutyCycle         big.Int
	CalculatedDuty    big.Int
	SetPoint          physic.Temperature
	ManualDuty        big.Int
	ManualTime        big.Int
}

// PidSettings define the actual values for heating/cooling as persisted
type PidSettings struct {
	Proportional big.Float
	Integral     big.Float
	Differential big.Float
	CycleTime    big.Int
	Delay        big.Int
}

// HysteriaSettings are used for Hysteria mode
type HysteriaSettings struct {
	MaxTemp physic.Temperature
	MinTemp physic.Temperature
	MinTime big.Int
}

// FindTemperatureControllerForProbe returns the pid controller associated with the TemperatureProbe
func FindTemperatureControllerForProbe(physAddr string) *TemperatureController {
	for _, controller := range controllers {
		for _, probe := range controller.TemperatureProbes {
			if probe.PhysAddr == physAddr {
				return controller
			}
		}
	}
	return nil
}

// FindTemperatureControllerByName returns the pid controller with a specific name
func FindTemperatureControllerByName(name string) *TemperatureController {
	controller := controllers[name]
	if controller != nil {
		log.Printf("Found PID Controller for %v: %v", name, controller)
	}
	return controller
}

// ClearControllers reset the map of controllers
func ClearControllers() {
	controllers = make(map[string]*TemperatureController)
}

// CreateTemperatureController Create a new PID controller for the Temperature probe
// name -> The name of the PID Controller
// probe -> The probe to associate with the controller
// If a PID controller exists with the same name, the probe will be added to it (or no-op if it is already assigned)
// If the passed in TemperatureProbe is associated with a different PID Controller, an error will be returned and no PID controller will be returned
func CreateTemperatureController(name string, probe *TemperatureProbe) (*TemperatureController, error) {
	existingControllerForProbe := FindTemperatureControllerForProbe(probe.PhysAddr)
	controller := FindTemperatureControllerByName(name)

	if existingControllerForProbe != nil {
		if existingControllerForProbe == controller {
			return controller, nil
		}
		return nil, fmt.Errorf("Temperature Controller (%v) exists for this probe, trying removing it first", existingControllerForProbe)
	}

	if controller == nil {
		controller = &TemperatureController{Name: name}
		controllers[name] = controller
	}

	controller.TemperatureProbes = append(controller.TemperatureProbes, probe)
	return controller, nil
}

// UpdateOutput updates the temperatures and decides how to control the outputs
func (c *TemperatureController) UpdateOutput() {
	if len(c.LastReadings) >= 5 {
		var oldestKey *time.Time
		for key := range c.LastReadings {
			if oldestKey == nil || key.Before(*oldestKey) {
				oldestKey = &key
			}
		}
		delete(c.LastReadings, *oldestKey)
	}
	averageTemp := c.AverageTemperature()
	c.LastReadings[time.Now()] = averageTemp
}

// AverageTemperature Calculate the average temperature for a temperature controller over all the probes
func (c *TemperatureController) AverageTemperature() physic.Temperature {
	var totalTemp int64
	for _, probe := range c.TemperatureProbes {
		log.Printf("%v: %v", totalTemp, probe.Reading.Celsius())
		totalTemp += (int64)(probe.Reading)
	}

	return (physic.Temperature)(totalTemp / (int64)(len(c.TemperatureProbes)))
}
