use onewire::{OneWire, DeviceSearch, DS18B20};

trait TempProbe {
    fn temperature(&self) {
        println!("Called DS18B20::temperature()");
        // let mut wire = OneWire::new(&mut one, false);
    
        // if wire.reset(&mut delay).is_err() {
        //     // missing pullup or error on line
        //     loop {}
        // }
        
        // let mut search = DeviceSearch::new_for_family(DS18B20.device.family_code());
    }
}