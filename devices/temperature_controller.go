package devices

import (
	"fmt"
	"log"
	"time"

	"periph.io/x/periph/conn/physic"
)

var controllers []*TemperatureController

// TemperatureController defines a mapping of temperature probes to their control settings
type TemperatureController struct {
	Name                    string
	LastReadings            []physic.Temperature
	TemperatureProbes       []*TemperatureProbe
	CoolSettings            PidSettings
	HeatSettings            PidSettings
	HysteriaSettings        HysteriaSettings
	ManualSettings          ManualSettings
	Mode                    string // Mode of this controller
	DutyCycle               int64
	CalculatedDuty          int64
	SetPoint                physic.Temperature
	PreviousCalculationTime time.Time
	TotalDiff               float64 // Always in Fahrenheit (internal calculation)
	integralError           float64
	derivativeFactor        float64
	prevDiff                float64 // Always in Fahrenheit (internal calculaiton)
}

// PidSettings define the actual values for heating/cooling as persisted
type PidSettings struct {
	Proportional float64
	Integral     float64
	Derivative   float64
	CycleTime    int64
	Delay        int64
	Configured   bool
}

// HysteriaSettings are used for Hysteria mode
type HysteriaSettings struct {
	MaxTemp    physic.Temperature
	MinTemp    physic.Temperature
	MinTime    int64 // In seconds
	Configured bool
}

type ManualSettings struct {
	DutyCycle  int64
	CycleTime  int64
	Configured bool
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
	for i := range controllers {
		if controllers[i].Name == name {
			return controllers[i]
		}
	}
	return nil
}

// ClearControllers reset the map of controllers
func ClearControllers() {
	controllers = nil
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
		controller = &TemperatureController{Name: name, TotalDiff: 0, integralError: 0, derivativeFactor: 0, prevDiff: 0}
		controllers = append(controllers, controller)
	}

	controller.TemperatureProbes = append(controller.TemperatureProbes, probe)
	return controller, nil
}

// UpdateOutput updates the temperatures and decides how to control the outputs
func (c *TemperatureController) UpdateOutput() {
	if len(c.LastReadings) >= 5 {
		c.LastReadings = c.LastReadings[1:5]
	}
	averageTemp := c.AverageTemperature()
	c.LastReadings = append(c.LastReadings, averageTemp)
	switch c.Mode {
	case "auto":
		c.CalculatedDuty = c.Calculate(averageTemp, nil)
	case "manual":
	case "off":
	case "hysteria":

	}
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

// Calculate does the calculation for the probe
func (c *TemperatureController) Calculate(averageTemperature physic.Temperature, now func() time.Time) int64 {
	if now == nil {
		now = time.Now
	}
	calculationTime := now()

	if (c.PreviousCalculationTime == time.Time{}) {
		c.PreviousCalculationTime = calculationTime
		return c.DutyCycle
	}

	delta := calculationTime.Sub(c.PreviousCalculationTime)
	// only caculate updates if we're over 100ms (0.1s)
	if delta.Milliseconds() < 100 {
		return c.DutyCycle
	}

	var targetDiff = c.SetPoint.Fahrenheit() - averageTemperature.Fahrenheit()
	var msDiff = float64(delta.Milliseconds())
	c.TotalDiff = (c.TotalDiff + targetDiff) * msDiff
	var currErr = (targetDiff - c.prevDiff) / msDiff

	var output = (c.HeatSettings.Proportional * targetDiff) +
		(c.HeatSettings.Integral * c.TotalDiff) +
		(c.HeatSettings.Derivative * currErr)

	c.prevDiff = targetDiff
	c.PreviousCalculationTime = calculationTime
	if output > 100 {
		output = 100
	} else if output < 0 {
		output = 0
	}
	return int64(output)
}
