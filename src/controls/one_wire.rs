//extern crate embedded_hal as hal;

use onewire::{OneWire, DeviceSearch, ds18b20};
use onewire::ds18b20::DS18B20;
use rppal::gpio::Gpio;
use rppal::hal::{Delay};

use std::thread;
use std::time::{Duration, Instant};
use std::error::Error;

trait TempProbe {
    fn temperature(&self) -> Result<(), Box<dyn Error>> {
        println!("Called DS18B20::temperature()");
        
        let mut gpioc = Gpio::new()?.get(4)?.into_output();
        let mut wire = OneWire::new(&mut gpioc, false);
    
        let mut delay = Delay::new();
        if wire.reset(&mut delay).is_err() {
            // missing pullup or error on line
            loop {}
        }
        
         // search for devices
        let mut search = DeviceSearch::new();
        while let Some(device) = wire.search_next(&mut search, &mut delay).unwrap() {
            match device.address[0] {
                ds18b20::FAMILY_CODE => {
                    let mut ds18b20 = DS18B20::new(device).unwrap();
                    
                    // request sensor to measure temperature
                    let resolution = ds18b20.measure_temperature(&mut wire, &mut delay).unwrap();
                    
                    // wait for compeltion, depends on resolution 
                    delay.delay_ms(resolution.time_ms());
                    thread::sleep(Duration::from_millis(u64::from(resolution.time_ms())));
                    
                    // read temperature
                    let temperature = ds18b20.read_temperature(&mut wire, &mut delay).unwrap();
                },
                _ => {
                    // unknown device type            
                }
            }
        }
        
        
        Ok(())
    }
}