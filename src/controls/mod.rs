// This handles the general control methods and the like from individual controllers
// Examples include PID, GPIO, temperature probes
pub mod one_wire;

pub fn calculate_pid() {
    println!("Called controls::calculate_pid");
}