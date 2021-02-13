package main

import (
	"fmt"
	"time"
	"log"
	"periph.io/x/periph/devices/ds18b20"
	"periph.io/x/periph/experimental/host/netlink"
	"periph.io/x/periph/host"
	// "periph.io/x/conn/v3/onewirereg"
	// "periph.io/x/conn/v3/onewire"
	// "periph.io/x/conn/v3/physic"
	// "periph.io/x/devices/v3/ds18b20"
)

func readTemperatures(messages chan string) {
	defer close(messages)
	fmt.Println("Reading temps.")
	_, err := host.Init()
	if err != nil {
		log.Fatalf("failed to initialize periph: %v", err)
	}

	oneBus, err := netlink.New(001)
	if err != nil {
		log.Printf("Could not open Netlink host: %v", err)
	} else {
		defer oneBus.Close()
	}

	// get 1wire address
	addresses, _ := oneBus.Search(false)
	fmt.Printf("Reading temps from %v devices.\n", addresses)
	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})

	for {
		select {
		case <- ticker.C:
			// readAddresses(oneBus, addresses, m)
			oneBus, _ := netlink.New(001)
			for _, device := range addresses {
				// init ds18b20
				sensor, _ := ds18b20.New(oneBus, device, 10)
		
				ds18b20.ConvertAll(oneBus, 10)
				temp, _ := sensor.LastTemp()
		
				messages <- fmt.Sprintf("Reading device %v: %v", device, temp)
			}
		case <- quit:
			ticker.Stop()
			fmt.Println("Stop")
			return
		}
	}
}


func main() {
	fmt.Println("Loaded and looking for temperatures")
	messages := make(chan string)
	go readTemperatures(messages)
	for {
		for message := range messages {
			fmt.Println(message)
		}
	}
}