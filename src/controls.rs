// This handles the general control methods and the like from individual controllers
// Examples include PID, GPIO, temperature probes
use std::{thread, time};
use one_wire::TempProbe;

pub mod one_wire;
pub mod w1_errors;


pub fn pid_loop() {
    let mut temp_probe = TempProbe::new("28-00000406f49f".to_string());
    
    loop {
        calculate_pid();
        next_iteration();
        temp_probe.update_temp();
        println!("Read temp: {}", temp_probe);
    }
}

fn calculate_pid() {
    println!("Called controls::calculate_pid");

}

fn next_iteration() {
    thread::sleep(time::Duration::from_millis(5000));
}