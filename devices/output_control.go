package devices

import (
	"fmt"
	"log"
	"math"
	"time"

	"gorm.io/gorm"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
)

// OutputControl is a basic struct to handle heating outputs with a duty cyclke
type OutputControl struct {
	gorm.Model
	HeatOutput *OutPin
	CoolOutput *OutPin
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
		log.Println("No off time")
		return
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
		log.Println("No on time")
		return
	}
	if err := op.PinIO.Out(gpio.High); err != nil {
		log.Fatal(err)
	}
	curTime := time.Now()
	op.offTime = nil
	op.onTime = &curTime
}

func (op *OutPin) reset() {
	op.PinIO = gpioreg.ByName(op.Identifier)
	if op.PinIO == nil {
		log.Fatalf("No Pin for %v!\n", op.Identifier)
	}
	
	op.off()
}

func (o *OutputControl) reset() {
	o.HeatOutput.reset()
	o.CoolOutput.reset()
}

// CalculateOutput - Turn on and off the output pin for this output control depending on the duty cycle
func (o *OutputControl) CalculateOutput() {
	cycleSeconds := math.Abs(float64(o.CycleTime * o.DutyCycle) / 100)

	if o.DutyCycle == 0 {
		o.HeatOutput.off()
	} else if o.DutyCycle > 0 {
		o.CoolOutput.off();

		fmt.Printf("Heat: off: %v, on: %v\n", o.HeatOutput.offTime, o.HeatOutput.onTime)
		if o.HeatOutput.onTime != nil {
			// it's on, do we need to turn it off?
			changeAt := time.Since(*o.HeatOutput.onTime)
			if changeAt.Seconds() >= float64(cycleSeconds) {
				fmt.Printf("Heat output turning off after %v seconds\n", changeAt.Seconds())
				o.HeatOutput.off()
			} else {
				remaining := changeAt.Seconds() - float64(cycleSeconds)
				fmt.Printf("Heat output to turn off in %v seconds\n", remaining)
			}
		} else if o.HeatOutput.offTime != nil {
			// it's off, do we need to turn it on?
			changeAt := time.Since(*o.HeatOutput.offTime)
			offSeconds := float64(o.CycleTime) - cycleSeconds
			if changeAt.Seconds() >= offSeconds {
				fmt.Printf("Heat output turning on after %v seconds\n", changeAt.Seconds())
				o.HeatOutput.on()
			} else {
				remaining := changeAt.Seconds() - offSeconds
				fmt.Printf("Heat output to turn on in %v seconds\n", remaining)
			}
		}
	} else if o.DutyCycle < 0 {
		o.HeatOutput.off();
		
		fmt.Printf("Cool: off: %v, on: %v\n", o.CoolOutput.offTime, o.CoolOutput.onTime)
		if o.CoolOutput.onTime != nil {
			// it's on, do we need to turn it off?
			changeAt := time.Since(*o.CoolOutput.onTime)
			if changeAt.Seconds() >= float64(cycleSeconds) {
				fmt.Printf("Cool output turning off after %v seconds\n", changeAt.Seconds())
				o.CoolOutput.off()
			} else {
				remaining := changeAt.Seconds() - float64(cycleSeconds)
				fmt.Printf("Cool output to turn off in %v seconds\n", remaining)
			}
		} else if o.CoolOutput.offTime != nil {
			// it's off, do we need to turn it on?
			changeAt := time.Since(*o.CoolOutput.offTime)
			offSeconds := float64(o.CycleTime) - cycleSeconds
			if changeAt.Seconds() >= offSeconds {
				fmt.Printf("Cool output turning on after %v seconds\n", changeAt.Seconds())
				o.CoolOutput.on()
			} else {
				remaining := changeAt.Seconds() - offSeconds
				fmt.Printf("Cool output to turn on in %v seconds\n", remaining)
			}
		}
	}
}

// RunControl -> Run the output controller for a heating output
func (o *OutputControl) RunControl() {
	fmt.Println("Starting output control")
	o.reset()
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
