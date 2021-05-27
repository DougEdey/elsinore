package devices

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/dougedey/elsinore/database"
	"github.com/dougedey/elsinore/graph/model"
	"github.com/dougedey/elsinore/hardware"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"periph.io/x/periph/conn/physic"
)

var controllers []*TemperatureController

// TempProbeDetail is the persisted model for TemperatureProbe
// this is to simplify loading data so that TemperatureProbe represents the physical state and this represents the cached state
type TempProbeDetail struct {
	gorm.Model
	TemperatureControllerID uint
	PhysAddr                string
	FriendlyName            string
	ReadingRaw              physic.Temperature `gorm:"-"`
	Updated                 time.Time
}

// TemperatureController defines a mapping of temperature probes to their control settings
type TemperatureController struct {
	gorm.Model
	Name             string
	LastReadings     []physic.Temperature `gorm:"-"`
	TempProbeDetails []*TempProbeDetail
	CoolSettings     PidSettings `gorm:"polymorphic:TemperatureController;polymorphicValue:coolSettings"`
	// CoolSettingsID					uint
	HeatSettings PidSettings `gorm:"polymorphic:TemperatureController;polymorphicValue:heatSettings"`
	// HeatSettingsID					uint
	HysteriaSettings HysteriaSettings
	// HysteriaSettingsID			uint
	ManualSettings          ManualSettings
	Mode                    model.ControllerMode // Mode of this controller
	DutyCycle               int64
	CalculatedDuty          int64
	SetPointRaw             physic.Temperature
	PreviousCalculationTime time.Time      `gorm:"-"`
	TotalDiff               float64        `gorm:"-"` // Always in Fahrenheit (internal calculation)
	integralError           float64        `gorm:"-"`
	derivativeFactor        float64        `gorm:"-"`
	prevDiff                float64        `gorm:"-"` // Always in Fahrenheit (internal calculaiton)
	OutputControl           *OutputControl `gorm:"-"`
}

// PidSettings define the actual values for heating/cooling as persisted
type PidSettings struct {
	gorm.Model
	Proportional              float64
	Integral                  float64
	Derivative                float64
	CycleTime                 int64
	Delay                     int64
	Configured                bool
	TemperatureControllerID   uint
	TemperatureControllerType string
	Gpio                      string
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

// FindTemperatureControllerForProbe returns the pid controller associated with the TemperatureProbe
func FindTemperatureControllerForProbe(physAddr string) *TemperatureController {
	for _, controller := range controllers {
		for _, probe := range controller.TempProbeDetails {
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

	if database.FetchDatabase() == nil {
		log.Printf("No Database configured, cannot lookup by Name")
		return nil
	}

	var controller *TemperatureController = nil
	result := database.FetchDatabase().Debug().Preload(clause.Associations).First(&controller, "name = ?", name)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil
	}

	database.FetchDatabase().Debug().
		Where("temperature_controller_type = ? AND temperature_controller_id = ?", "coolSettings", controller.ID).
		First(&controller.CoolSettings)
	database.FetchDatabase().Debug().
		Where("temperature_controller_type = ? AND temperature_controller_id = ?", "heatSettings", controller.ID).
		First(&controller.HeatSettings)
	controllers = append(controllers, controller)
	return controller
}

// AllTemperatureControllers returns all the temperature controllers
func AllTemperatureControllers() []*TemperatureController {
	return controllers
}

// FindTemperatureControllerByID - Find a temperature controller by id, preloading everything
func FindTemperatureControllerByID(id string) *TemperatureController {
	intID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil
	}
	fmt.Printf("Converting %v to an int %v, %v\n", id, intID, len((controllers)))

	for i := range controllers {
		fmt.Printf("Comparing %v to %v\n", controllers[i].ID, uint(intID))
		if controllers[i].ID == uint(intID) {
			return controllers[i]
		}
	}
	var controller *TemperatureController = nil
	if database.FetchDatabase() == nil {
		log.Printf("No Database configured, cannot lookup by Id")
		return nil
	}
	result := database.FetchDatabase().Debug().Preload(clause.Associations).First(&controller, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil
	}
	database.FetchDatabase().Debug().
		Where("temperature_controller_type = ? AND temperature_controller_id = ?", "coolSettings", controller.ID).
		First(&controller.CoolSettings)
	database.FetchDatabase().Debug().
		Where("temperature_controller_type = ? AND temperature_controller_id = ?", "heatSettings", controller.ID).
		First(&controller.HeatSettings)
	controllers = append(controllers, controller)
	return controller
}

// DeleteTemperatureControllerByID - Delete a temperature controller with the ID specified and return the probes that are now freed
func DeleteTemperatureControllerByID(id string) []*string {
	var controller *TemperatureController
	if database.FetchDatabase() == nil {
		log.Printf("No Database configured, cannot lookup by Id")
		return nil
	}
	controller = FindTemperatureControllerByID(id)
	if controller == nil {
		fmt.Printf("Could not find the recordx for %v", id)
		return nil
	}

	probeList := []*string{}
	for _, t := range controller.TempProbeDetails {
		probeList = append(probeList, &t.PhysAddr)
	}

	database.FetchDatabase().Delete(&controller, id)

	return probeList
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
func CreateTemperatureController(name string, probe *TempProbeDetail) (*TemperatureController, error) {
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
	} else {
		fmt.Printf("Found controller for %v", controller.ID)
	}

	controller.TempProbeDetails = append(controller.TempProbeDetails, probe)
	database.Save(&controller)
	return controller, nil
}

// RemoveProbe removes a temperature probe from this controller
func (c *TemperatureController) RemoveProbe(physAddr string) error {
	for i, probe := range c.TempProbeDetails {
		if probe.PhysAddr == physAddr {
			c.TempProbeDetails[i] = c.TempProbeDetails[len(c.TempProbeDetails)-1]
			c.TempProbeDetails = c.TempProbeDetails[:len(c.TempProbeDetails)-1]
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
	for _, probe := range c.TempProbeDetails {
		log.Printf("%v: %v", totalTemp, probe.ReadingRaw.Celsius())
		totalTemp += (int64)(probe.ReadingRaw)
	}

	return (physic.Temperature)(totalTemp / (int64)(len(c.TempProbeDetails)))
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

// ApplySettings - Update the current temperature controller settings
func (c *TemperatureController) ApplySettings(newSettings model.TemperatureControllerSettingsInput) error {
	err := c.CoolSettings.ApplySettings(newSettings.CoolSettings)
	if err != nil {
		return err
	}

	err = c.HeatSettings.ApplySettings(newSettings.HeatSettings)
	if err != nil {
		return err
	}

	err = c.ManualSettings.ApplySettings(newSettings.ManualSettings)
	if err != nil {
		return err
	}

	err = c.HysteriaSettings.ApplySettings(newSettings.HysteriaSettings)
	if err != nil {
		return err
	}

	if newSettings.Name != nil {
		c.Name = *newSettings.Name
	}

	if newSettings.Mode != nil {
		c.Mode = *newSettings.Mode
	}
	database.Save(c)

	if c.HeatSettings.Gpio != "" || c.CoolSettings.Gpio != "" {
		if c.OutputControl == nil {
			log.Println("Turning on output control")
			heatingPin := OutPin{Identifier: c.HeatSettings.Gpio, FriendlyName: "Heating"}
			coolingPin := OutPin{Identifier: c.CoolSettings.Gpio, FriendlyName: "Cooling"}
			outputControl := OutputControl{HeatOutput: &heatingPin, CoolOutput: &coolingPin, DutyCycle: c.ManualSettings.DutyCycle, CycleTime: c.HeatSettings.CycleTime}
			c.OutputControl = &outputControl
			go c.OutputControl.RunControl()
		} else {
			log.Println("Updating output control")
			c.OutputControl.DutyCycle = c.ManualSettings.DutyCycle
			c.OutputControl.CycleTime = c.HeatSettings.CycleTime
		}
	} else {
		log.Println("Turning off output control")
		c.OutputControl.Reset()
		c.OutputControl = nil
	}
	return nil
}

// ApplySettings - Update the current pid settings
func (s *PidSettings) ApplySettings(newSettings *model.PidSettingsInput) error {
	if newSettings == nil {
		return nil
	}
	if newSettings.Configured != nil {
		s.Configured = *newSettings.Configured
	}
	if newSettings.CycleTime != nil {
		s.CycleTime = int64(*newSettings.CycleTime)
	}
	if newSettings.Delay != nil {
		s.Delay = int64(*newSettings.Delay)
	}
	if newSettings.CycleTime != nil {
		s.CycleTime = int64(*newSettings.CycleTime)
	}
	if newSettings.Proportional != nil {
		s.Proportional = *newSettings.Proportional
	}
	if newSettings.Integral != nil {
		s.Integral = *newSettings.Integral
	}
	if newSettings.Derivative != nil {
		s.Derivative = *newSettings.Derivative
	}
	if newSettings.Gpio != nil {
		s.Gpio = *newSettings.Gpio
	}
	return nil
}

// ApplySettings - Update the current manual settings
func (s *ManualSettings) ApplySettings(newSettings *model.ManualSettingsInput) error {
	if newSettings == nil {
		return nil
	}
	if newSettings.Configured != nil {
		s.Configured = *newSettings.Configured
	}
	if newSettings.CycleTime != nil {
		s.CycleTime = int64(*newSettings.CycleTime)
	}
	if newSettings.DutyCycle != nil {
		s.DutyCycle = int64(*newSettings.DutyCycle)
	}
	return nil
}

// ApplySettings - Update the current hysteria settings
func (h *HysteriaSettings) ApplySettings(newSettings *model.HysteriaSettingsInput) error {
	if newSettings == nil {
		return nil
	}
	if newSettings.Configured != nil {
		h.Configured = *newSettings.Configured
	}
	if newSettings.MaxTemp != nil {
		err := h.MaxTempRaw.Set(*newSettings.MaxTemp)
		if err != nil {
			return err
		}
	}
	if newSettings.MinTemp != nil {
		err := h.MinTempRaw.Set(*newSettings.MinTemp)
		if err != nil {
			return err
		}
	}
	if newSettings.MinTime != nil {
		h.MinTime = int64(*newSettings.MinTime)
	}
	return nil
}

// UpdateTemperature Set the temperature on the Temperature Probe from a string
func (t *TempProbeDetail) UpdateTemperature(newTemp string) error {
	return t.ReadingRaw.Set(newTemp)
}

// Reading The current temperature reading for the probe
func (t *TempProbeDetail) Reading() string {
	return t.ReadingRaw.String()
}

// UpdateReading -  Update the reading from the associated probe
func (t *TempProbeDetail) UpdateReading() {
	err := t.ReadingRaw.Set(hardware.GetTemperature(t.PhysAddr).Reading())
	if err != nil {
		log.Printf("Failed to update %v temperature details: %v", t.PhysAddr, err)
	}
}
