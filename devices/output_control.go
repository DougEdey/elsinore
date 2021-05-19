package devices

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
)

// OutputControl is a basic struct to handle heating outputs with a duty cyclke
type OutputControl struct {
	gorm.Model
	HeatOutput *OutPin
	DutyCycle  int64 `gorm:"-"`
	CycleTime  int64 `gorm:"-"`
}

// OutPin represents a stored output pin with a friendly name
type OutPin struct {
	gorm.Model

	Identifier   string
	FriendlyName string
	PinIO        gpio.PinIO `gorm:"-"`
	onTime       *time.Time
	offTime      *time.Time
}

func (op *OutPin) off() {
	if op.offTime != nil {
		return
	}
	if op.PinIO == nil {
		op.PinIO = gpioreg.ByName(op.Identifier)
	}

	if err := op.PinIO.Out(gpio.Low); err != nil {
		log.Fatal(err)
	}

	curTime := time.Now()
	op.offTime = &curTime
	op.onTime = nil
}

func (op *OutPin) on() {
	if op.onTime != nil {
		return
	}
	if op.PinIO == nil {
		op.PinIO = gpioreg.ByName(op.Identifier)
	}

	if err := op.PinIO.Out(gpio.High); err != nil {
		log.Fatal(err)
	}

	curTime := time.Now()
	op.offTime = nil
	op.onTime = &curTime
}

// CalculateOutput - Turn on and off the output pin for this output control depending on the duty cycle
func (o *OutputControl) CalculateOutput() {
	if o.HeatOutput.offTime == nil && o.HeatOutput.onTime == nil {
		o.HeatOutput.off()
	}
	if o.DutyCycle == 0 {
		o.HeatOutput.off()
	} else if o.DutyCycle > 0 {
		cycleSeconds := (o.CycleTime * o.DutyCycle) / 100
		fmt.Printf("off: %v, on: %v", o.HeatOutput.offTime, o.HeatOutput.onTime)
		if o.HeatOutput.onTime != nil {
			// it's on, do we need to turn it off?
			changeAt := time.Since(*o.HeatOutput.onTime)
			if changeAt.Seconds() >= float64(cycleSeconds) {
				fmt.Printf("Heat output turning off after %v seconds", changeAt.Seconds())
				o.HeatOutput.off()
			} else {
				remaining := changeAt.Seconds() - float64(cycleSeconds)
				fmt.Printf("Heat output to turn off in %v seconds", remaining)
			}
		} else if o.HeatOutput.offTime != nil {
			// it's off, do we need to turn it on?
			changeAt := time.Since(*o.HeatOutput.offTime)
			offSeconds := o.CycleTime - cycleSeconds
			if changeAt.Seconds() >= float64(offSeconds) {
				fmt.Printf("Heat output turning on after %v seconds", changeAt.Seconds())
				o.HeatOutput.on()
			} else {
				remaining := changeAt.Seconds() - float64(offSeconds)
				fmt.Printf("Heat output to turn on in %v seconds", remaining)
			}
		}
	} else if o.DutyCycle < 0 {
		// No support for cooler outputs yet
		o.HeatOutput.off()
	}
}

// RunControl -> Run the output controller for a heating output
func (o *OutputControl) RunControl() {
	fmt.Println("Starting output control")
	duration, err := time.ParseDuration("500ms")
	if err != nil {
		log.Fatal(err)
	}
	ticker := time.NewTicker(duration)
	quit := make(chan struct{})

	for {
		select {
		case <-ticker.C:
			o.CalculateOutput()
		case <-quit:
			ticker.Stop()
			fmt.Println("Stop")
			return
		}
	}
}
