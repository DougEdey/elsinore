// This handles the general control methods and the like from individual controllers
// Examples include PID, GPIO, temperature probes
use std::{thread, time};

pub mod one_wire;

pub fn pid_loop() {
    loop {
        one_wire::temperature();
        calculate_pid();
        
        next_iteration();
    }
}

fn calculate_pid() {
    println!("Called controls::calculate_pid");
}

fn next_iteration() {
    thread::sleep(time::Duration::from_millis(5000));
}