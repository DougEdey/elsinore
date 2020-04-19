#![feature(proc_macro_hygiene, decl_macro)]

#[macro_use] extern crate rocket;

mod controls;
use std::{thread, time};

#[get("/")]
fn index() -> &'static str {
    "Hello, world!"
}

fn main() {
    thread::spawn(|| {
        loop {
            controls::calculate_pid();
            controls::one_wire::temperature();

            thread::sleep(time::Duration::from_millis(5000));
        }
    });
    rocket::ignite().mount("/", routes![index]).launch();
    
}