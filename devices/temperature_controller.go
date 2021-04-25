package devices

import (
	"fmt"
	"log"
	"time"

	"github.com/dougedey/elsinore/database"
	"gorm.io/gorm"
	"periph.io/x/periph/conn/physic"
)

var controllers []*TemperatureController

// TemperatureController defines a mapping of temperature probes to their control settings
type TemperatureController struct {
	gorm.Model
	Name                    string
	LastReadings            []physic.Temperature `gorm:"-"`
	TemperatureProbes       []*TemperatureProbe
	CoolSettings            PidSettings
	HeatSettings            PidSettings
	HysteriaSettings        HysteriaSettings
	ManualSettings          ManualSettings
	Mode                    ControllerMode // Mode of this controller
	DutyCycle               int64
	CalculatedDuty          int64
	SetPointRaw             physic.Temperature
	PreviousCalculationTime time.Time `gorm:"-"`
	TotalDiff               float64   `gorm:"-"` // Always in Fahrenheit (internal calculation)
	integralError           float64   `gorm:"-"`
	derivativeFactor        float64   `gorm:"-"`
	prevDiff                float64   `gorm:"-"` // Always in Fahrenheit (internal calculaiton)
	TemperatureControllerID uint
}

// PidSettings define the actual values for heating/cooling as persisted
type PidSettings struct {
	gorm.Model
	Proportional            float64
	Integral                float64
	Derivative              float64
	CycleTime               int64
	Delay                   int64
	Configured              bool
	TemperatureControllerID uint
}

// HysteriaSettings are used for Hysteria mode
type HysteriaSettings struct {
	gorm.Model
	MaxTempRaw              physic.Temperature
	MinTempRaw              physic.Temperature
	MinTime                 int64 // In seconds
	Configured              bool
	TemperatureControllerID uint
}

// ManualSettings are used for manually controlling the output
type ManualSettings struct {
	gorm.Model
	DutyCycle               int64
	CycleTime               int64
	Configured              bool
	TemperatureControllerID uint
}

// ControllerMode Auto, Manual, Off, Hysteria
type ControllerMode string

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
		return nil, fmt.Errorf("temperature Controller (%v) exists for this probe, trying removing it first", existingControllerForProbe)
	}

	if controller == nil {
		controller = &TemperatureController{Name: name, TotalDiff: 0, integralError: 0, derivativeFactor: 0, prevDiff: 0}
		database.Create(&controller)
		controllers = append(controllers, controller)
	}

	controller.TemperatureProbes = append(controller.TemperatureProbes, probe)
	database.Save(&controller)
	return controller, nil
}

// RemoveProbe removes a temperature probe from this controller
func (c *TemperatureController) RemoveProbe(physAddr string) error {
	for i, probe := range c.TemperatureProbes {
		if probe.PhysAddr == physAddr {
			c.TemperatureProbes[i] = c.TemperatureProbes[len(c.TemperatureProbes)-1]
			c.TemperatureProbes = c.TemperatureProbes[:len(c.TemperatureProbes)-1]
			return nil
		}
	}
	return fmt.Errorf("could not find a probe with address %v", physAddr)
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
		log.Printf("%v: %v", totalTemp, probe.ReadingRaw.Celsius())
		totalTemp += (int64)(probe.ReadingRaw)
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

	var targetDiff = c.SetPointRaw.Fahrenheit() - averageTemperature.Fahrenheit()
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

// SetPoint -> THe target Setpoint for this controller
func (c *TemperatureController) SetPoint() string {
	return c.SetPointRaw.String()
}

// MaxTemp -> For hysteria, this is the string for the max temp to turn off
func (h *HysteriaSettings) MaxTemp() string {
	return h.MaxTempRaw.String()
}

// MinTemp -> For hysteria, this is the string for the min temp to turn on
func (h *HysteriaSettings) MinTemp() string {
	return h.MinTempRaw.String()
}
