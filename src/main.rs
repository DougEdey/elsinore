#![feature(proc_macro_hygiene, decl_macro)]

#[macro_use] extern crate rocket;

mod controls;
use std::thread;

#[get("/")]
fn index() -> &'static str {
    "Hello, world!"
}

fn main() {
    thread::spawn(|| {
        controls::pid_loop();
    });
    rocket::ignite().mount("/", routes![index]).launch();
    
}