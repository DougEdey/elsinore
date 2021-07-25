package devices_test

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/dougedey/elsinore/database"
	"github.com/dougedey/elsinore/devices"
	"periph.io/x/periph/conn/gpio/gpiotest"
	"periph.io/x/periph/conn/physic"
)

func setupTestDb(t *testing.T) {
	dbName := "test"
	database.InitDatabase(&dbName,
		&devices.TempProbeDetail{}, &devices.PidSettings{}, &devices.HysteriaSettings{},
		&devices.ManualSettings{}, &devices.TemperatureController{}, &devices.Switch{},
	)

	t.Cleanup(func() {
		database.Close()
		e := os.Remove("test.db")
		if e != nil {
			t.Fatal(e)
		}
	})
}

func TestCreateTemperatureController(t *testing.T) {
	devices.ClearControllers()
	probe := devices.TempProbeDetail{
		PhysAddr: "ARealAddress",
	}

	t.Run("A new Temperature controller is created if no existing device with the same name exists", func(t *testing.T) {
		devices.ClearControllers()
		temperatureController, err := devices.CreateTemperatureController("sample", &probe)
		if err != nil {
			t.Fatalf("Failed to create the Temperature Controller: %v", err)
		}

		if temperatureController == nil {
			t.Fatal("No Pid Controller returned for sample")
		}

		if temperatureController.Name != "sample" {
			t.Fatalf("Expected the temperature controller to be called sample, but got %v", temperatureController.Name)
		}
	})

	t.Run("A Temperature controller cannot be created if the probe is already associated with a controller", func(t *testing.T) {
		temperatureController, err := devices.CreateTemperatureController("sample_2", &probe)

		if err == nil {
			t.Fatalf("Created a duplicate Temperature controller: %v", temperatureController)
		}
	})

	t.Run("Re-adding a probe to the same temperature controller is a no-op", func(t *testing.T) {
		temperatureController, err := devices.CreateTemperatureController("sample", &probe)

		if err != nil {
			t.Fatalf("Failed to do nothing: %v", err)
		}

		existingController := devices.FindTemperatureControllerByName("sample")
		if temperatureController != existingController {
			t.Fatalf("Expected %v, but got %v", existingController, temperatureController)
		}
	})

	t.Run("A Temperature controller is returned by reference", func(t *testing.T) {
		temperatureController, err := devices.CreateTemperatureController("sample_2", &probe)

		if err == nil {
			t.Fatalf("Created a duplicate Temperature controller: %v", temperatureController)
		}
	})

	t.Run("Updating a temperature controller name makes it findable by default", func(t *testing.T) {
		devices.ClearControllers()
		temperatureController, err := devices.CreateTemperatureController("sample", &probe)
		if err != nil {
			t.Fatalf("Failed to create the Temperature Controller: %v", err)
		}

		temperatureController.Name = "Some new name"

		existingController := devices.FindTemperatureControllerByName("Some new name")
		if temperatureController != existingController {
			t.Fatalf("Expected %v, but got %v", existingController, temperatureController)
		}
	})
}

func TestPersistenceTemperatureController(t *testing.T) {
	t.Run("When no DB is configured, nil is returned", func(t *testing.T) {
		var tempController *devices.TemperatureController = devices.FindTemperatureControllerByID("1")

		if tempController != nil {
			t.Fatalf("Expected the temperature controller to be called nil, but got %v", tempController)
		}
	})

	setupTestDb(t)
	devices.ClearControllers()
	probe := devices.TempProbeDetail{
		PhysAddr: "ARealAddress",
	}

	t.Run("A non existant ID returns nil", func(t *testing.T) {
		var tempController *devices.TemperatureController = devices.FindTemperatureControllerByID("10000")

		if tempController != nil {
			t.Fatalf("Expected the temperature controller to be called nil, but got %v", tempController)
		}
	})

	t.Run("An invalid ID returns nil", func(t *testing.T) {
		var tempController *devices.TemperatureController = devices.FindTemperatureControllerByID("f")

		if tempController != nil {
			t.Fatalf("Expected the temperature controller to be called nil, but got %v", tempController)
		}
	})

	t.Run("A new Temperature controller is persisted to the DB when configured", func(t *testing.T) {
		devices.ClearControllers()
		temperatureController, err := devices.CreateTemperatureController("sample", &probe)
		if err != nil {
			t.Fatalf("Failed to create the Temperature Controller: %v", err)
		}

		if temperatureController == nil {
			t.Fatal("No Pid Controller returned for sample")
		}

		var dbTempController devices.TemperatureController
		database.FetchDatabase().First(&dbTempController)
		if dbTempController.Name != "sample" {
			t.Fatalf("Expected the temperature controller to be called sample, but got %v", temperatureController.Name)
		}

		devices.ClearControllers()

		var tempController *devices.TemperatureController = devices.FindTemperatureControllerByID(fmt.Sprint(dbTempController.ID))

		if tempController.Name != "sample" {
			t.Fatalf("Expected the temperature controller to be called sample, but got %v", tempController.Name)
		}

		devices.ClearControllers()

		tempController = devices.FindTemperatureControllerByName(dbTempController.Name)

		if tempController.Name != "sample" {
			t.Fatalf("Expected the temperature controller to be called sample, but got %v", tempController.Name)
		}
	})

	t.Run("A temperature controller can be deleted by id", func(t *testing.T) {
		temperatureController, err := devices.CreateTemperatureController("sample", &probe)
		if err != nil {
			t.Fatalf("Failed to create the Temperature Controller: %v", err)
		}

		if temperatureController == nil {
			t.Fatal("No Pid Controller returned for sample")
		}
		if temperatureController.ID == 0 {
			t.Fatalf("No ID!")
		}

		p := devices.DeleteTemperatureControllerByID(fmt.Sprint(temperatureController.ID))
		if len(p) != 1 {
			t.Fatalf("Expected one probe, but got %v", len(p))
		}

		var dbTempController *devices.TemperatureController = devices.FindTemperatureControllerByName("sample")
		if dbTempController != nil {
			t.Fatalf("Expected nil but got %v", dbTempController)
		}
	})
}

func TestTemperatureControllerAverageTemperature(t *testing.T) {
	devices.ClearControllers()
	probe := devices.TempProbeDetail{
		PhysAddr:   "ARealAddress",
		ReadingRaw: physic.Temperature(0),
	}
	temperatureController, err := devices.CreateTemperatureController("sample", &probe)

	if err != nil {
		t.Fatalf("Failed to create the controller: %v", err)
	}

	t.Run("With a single probe, you get the current value", func(t *testing.T) {
		err = probe.UpdateTemperature("35C")
		if err != nil {
			log.Fatalf("Failed to update %v", err)
		}
		avgTemperature := temperatureController.AverageTemperature()
		if float64(35.0) != avgTemperature.Celsius() {
			t.Fatalf("Expected %v, but got %v", probe.ReadingRaw, avgTemperature)
		}
	})

	t.Run("With multiple probes, you get an average value", func(t *testing.T) {
		probe_two := devices.TempProbeDetail{
			PhysAddr:   "AnotherRealAddress",
			ReadingRaw: physic.Temperature(0),
		}
		_, err = devices.CreateTemperatureController("sample", &probe_two)
		if err != nil {
			log.Fatalf("Failed to create %v", err)
		}

		err = probe_two.UpdateTemperature("37C")
		if err != nil {
			log.Fatalf("Failed to update %v", err)
		}

		avgTemperature := temperatureController.AverageTemperature()
		if float64(36.0) != avgTemperature.Celsius() {
			t.Fatalf("Expected 36C, but got %v", avgTemperature)
		}
	})
}

func TestTemperatureControllerUpdateOutput(t *testing.T) {
	devices.ClearControllers()

	probe := devices.TempProbeDetail{
		PhysAddr:   "ARealAddress",
		ReadingRaw: physic.Temperature(0),
	}
	err := probe.UpdateTemperature("35C")
	if err != nil {
		log.Fatalf("Failed to update %v", err)
	}
	temperatureController, err := devices.CreateTemperatureController("sample", &probe)

	if err != nil {
		t.Fatalf("Failed to create the controller: %v", err)
	}

	t.Run("Adds to the last readings up to 5 times", func(t *testing.T) {
		for i := 1; i <= 5; i++ {
			temperatureController.UpdateOutput()
			if i != len(temperatureController.LastReadings) {
				t.Fatalf("Expected %v temperature reading but got %v", i, len(temperatureController.LastReadings))
			}
		}
	})

	t.Run("The 6th temperature removes the oldest temperature", func(t *testing.T) {
		toDelete := temperatureController.LastReadings[0]

		if len(temperatureController.LastReadings) != 5 {
			t.Fatalf("Expected %v temperature reading but got %v", 5, len(temperatureController.LastReadings))
		}

		temperatureController.UpdateOutput()
		if len(temperatureController.LastReadings) != 5 {
			t.Fatalf("Expected %v temperature reading but got %v", 5, len(temperatureController.LastReadings))
		}

		if &temperatureController.LastReadings[0] == &toDelete {
			t.Fatalf("Expected %v to not be the same as the deleted value!", temperatureController.LastReadings[0])
		}
	})

	t.Run("When auto mode and no output control, do nothing", func(t *testing.T) {
		temperatureController.Mode = "auto"
		temperatureController.UpdateOutput()
	})

	t.Run("When off mode and no output control, do nothing", func(t *testing.T) {
		temperatureController.Mode = "off"
		temperatureController.UpdateOutput()
	})

	t.Run("When manual mode and no output control, do nothing", func(t *testing.T) {
		temperatureController.Mode = "manual"
		temperatureController.UpdateOutput()
	})

	t.Run("When auto mode and only heat settings configured, use the cool settings cycle time", func(t *testing.T) {
		temperatureController.Mode = "auto"
		testPin := gpiotest.Pin{N: "GPIO21", Num: 10, Fn: "I2C1_SDA"}
		out21 := devices.OutPin{Identifier: "GPIO21", FriendlyName: "GPIO21", PinIO: &testPin}
		outputControl := devices.OutputControl{HeatOutput: &out21}
		temperatureController.OutputControl = &outputControl
		temperatureController.HeatSettings = devices.PidSettings{Configured: true, CycleTime: 12, Proportional: 1}
		temperatureController.SetPointRaw.Set("40C")
		temperatureController.UpdateOutput()

		if outputControl.CycleTime != temperatureController.HeatSettings.CycleTime {
			t.Fatalf("Cycle time incorrect: %v, expected: %v", outputControl.CycleTime, temperatureController.HeatSettings.CycleTime)
		}
	})

	t.Run("When auto mode and only cool settings configured, use the cool settings cycle time", func(t *testing.T) {
		temperatureController.Mode = "auto"
		testPin := gpiotest.Pin{N: "GPIO21", Num: 10, Fn: "I2C1_SDA"}
		out21 := devices.OutPin{Identifier: "GPIO21", FriendlyName: "GPIO21", PinIO: &testPin}
		outputControl := devices.OutputControl{HeatOutput: &out21}
		temperatureController.OutputControl = &outputControl
		temperatureController.HeatSettings.Configured = false
		temperatureController.CoolSettings = devices.PidSettings{Configured: true, CycleTime: 13, Proportional: 1}
		temperatureController.SetPointRaw.Set("40C")
		temperatureController.UpdateOutput()

		if outputControl.CycleTime != temperatureController.CoolSettings.CycleTime {
			t.Fatalf("Cycle time incorrect: %v, expected: %v", outputControl.CycleTime, temperatureController.HeatSettings.CycleTime)
		}
	})

	t.Run("When manual mode and output control, use the manual mode settings", func(t *testing.T) {
		temperatureController.Mode = "manual"
		temperatureController.ManualSettings = devices.ManualSettings{DutyCycle: 90, CycleTime: 14, Configured: true}
		temperatureController.UpdateOutput()

		if temperatureController.OutputControl.DutyCycle != 90 {
			t.Fatalf("Expected Duty cycle of 90, but got %v", temperatureController.DutyCycle)
		}

		if temperatureController.OutputControl.CycleTime != 14 {
			t.Fatalf("Expected cycle time of 14, but got %v", temperatureController.DutyCycle)
		}
	})

	t.Run("When mode off with output control, set the duty cycle to 0", func(t *testing.T) {
		temperatureController.Mode = "off"
		temperatureController.UpdateOutput()

		if temperatureController.OutputControl.DutyCycle != 0 {
			t.Fatalf("Expected Duty cycle of 0, but got %v", temperatureController.DutyCycle)
		}
	})
}

func TestTemperatureControllerCalculate(t *testing.T) {
	devices.ClearControllers()
	rand.Seed(time.Now().UnixNano())

	stubNow := func() time.Time { return time.Unix(1615715366, 0) }

	probe := devices.TempProbeDetail{
		PhysAddr:   "ARealAddress",
		ReadingRaw: physic.Temperature(0),
	}
	err := probe.UpdateTemperature("35C")
	if err != nil {
		log.Fatalf("Failed to update %v", err)
	}
	temperatureController, err := devices.CreateTemperatureController("sample", &probe)

	if err != nil {
		t.Fatalf("Failed to create the controller: %v", err)
	}

	t.Run("Calculate sets the PreviousCalculationTime on first run and returns the current duty cycle", func(t *testing.T) {
		temperatureController.UpdateOutput()
		var output = temperatureController.Calculate(temperatureController.AverageTemperature(), stubNow)
		if output != 0 {
			t.Fatalf("Expected output to be %v, but got %v", 0, output)
		}
		if temperatureController.PreviousCalculationTime != stubNow() {
			t.Fatalf("Expected previous calculation time to be %v, but got %v", stubNow(), temperatureController.PreviousCalculationTime)
		}
	})

	t.Run("Calculate does not update anything if the current time is under 100ms ahead", func(t *testing.T) {
		temperatureController.UpdateOutput()
		stubNext := func() time.Time { return time.Unix(1615715366, int64(rand.Intn(100)*1_000_000)) }
		var output = temperatureController.Calculate(temperatureController.AverageTemperature(), stubNext)
		if output != 0 {
			t.Fatalf("Expected output to be %v, but got %v", 0, output)
		}
		if temperatureController.PreviousCalculationTime != stubNow() {
			t.Fatalf("Expected previous calculation time to be %v, but got %v", stubNow(), temperatureController.PreviousCalculationTime)
		}
	})

	t.Run("Calculate updates previous time when the difference is over 100ms", func(t *testing.T) {
		offset := int64(rand.Intn(100)+100) * 1_000_000
		stubNext := func() time.Time { return time.Unix(1615715366, offset) }
		var output = temperatureController.Calculate(temperatureController.AverageTemperature(), stubNext)
		if output != 0 {
			t.Fatalf("Expected output to be %v, but got %v", 0, output)
		}
		if temperatureController.PreviousCalculationTime != stubNext() {
			t.Fatalf("Expected previous calculation time to be %v, but got %v", stubNext(), temperatureController.PreviousCalculationTime)
		}
	})

	t.Run("Calculate uses proportional value when set", func(t *testing.T) {
		temperatureController.HeatSettings.Proportional = 10
		temperatureController.PreviousCalculationTime = stubNow()
		temperatureController.SetPointRaw.Set("36C")
		offset := int64(rand.Intn(100)+100) * 1_000_000
		stubNext := func() time.Time { return time.Unix(1615715366, offset) }
		var output = temperatureController.Calculate(temperatureController.AverageTemperature(), stubNext)
		if output != 18 {
			t.Fatalf("Expected output to be %v, but got %v", 18, output)
		}
		if temperatureController.PreviousCalculationTime != stubNext() {
			t.Fatalf("Expected previous calculation time to be %v, but got %v", stubNext(), temperatureController.PreviousCalculationTime)
		}
	})

	t.Run("Calculate uses proportional value when set with a large delta caps to 100", func(t *testing.T) {
		temperatureController.HeatSettings.Proportional = 10
		temperatureController.PreviousCalculationTime = stubNow()
		temperatureController.SetPointRaw.Set("100C")
		offset := int64(rand.Intn(100)+100) * 1_000_000
		stubNext := func() time.Time { return time.Unix(1615715366, offset) }
		var output = temperatureController.Calculate(temperatureController.AverageTemperature(), stubNext)
		if output != 100 {
			t.Fatalf("Expected output to be %v, but got %v", 100, output)
		}
		if temperatureController.PreviousCalculationTime != stubNext() {
			t.Fatalf("Expected previous calculation time to be %v, but got %v", stubNext(), temperatureController.PreviousCalculationTime)
		}
	})

	t.Run("Calculate uses proportional and integral values when set", func(t *testing.T) {
		temperatureController.HeatSettings.Proportional = 10
		temperatureController.HeatSettings.Integral = 0
		temperatureController.PreviousCalculationTime = stubNow()
		temperatureController.SetPointRaw.Set("36C")
		offset := int64((101 + 100) * 1_000_000)
		stubNext := func() time.Time { return time.Unix(1615715366, offset) }
		var output = temperatureController.Calculate(temperatureController.AverageTemperature(), stubNext)
		if output != 18 {
			t.Fatalf("Expected output to be %v, but got %v", 18, output)
		}

		temperatureController.TotalDiff = 0
		temperatureController.HeatSettings.Integral = 0.1
		stubNext = func() time.Time { return time.Unix(1615715366, offset*2) }
		output = temperatureController.Calculate(temperatureController.AverageTemperature(), stubNext)
		if output != 54 {
			t.Fatalf("Expected output to be %v, but got %v", 54, output)
		}
		if temperatureController.PreviousCalculationTime != stubNext() {
			t.Fatalf("Expected previous calculation time to be %v, but got %v", stubNext(), temperatureController.PreviousCalculationTime)
		}
	})
}

func TestRemoveProbeFromController(t *testing.T) {
	devices.ClearControllers()
	probe := devices.TempProbeDetail{
		PhysAddr: "ARealAddress",
	}
	temperatureController, err := devices.CreateTemperatureController("sample", &probe)
	if err != nil {
		t.Fatalf("Failed to create the Temperature Controller: %v", err)
	}

	if temperatureController == nil {
		t.Fatal("No Pid Controller returned for sample")
	}

	t.Run("Can remove the final probe from a controller", func(t *testing.T) {
		err := temperatureController.RemoveProbe("ARealAddress")
		if err != nil {
			t.Fatal(err)
		}

		if len(temperatureController.TempProbeDetails) != 0 {
			t.Fatalf("Expected no probes but got %v! %v", len(temperatureController.TempProbeDetails), temperatureController.TempProbeDetails)
		}
	})

	t.Run("A non existent probe returns an error", func(t *testing.T) {
		err := temperatureController.RemoveProbe("AFakeAddress")
		if err == nil {
			t.Fatal("Expected an error but didn't get one")
		}

		if err.Error() != "could not find a probe with address AFakeAddress" {
			t.Fatalf("Incorrect error: %v", err.Error())
		}
	})
}
