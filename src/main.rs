#![feature(proc_macro_hygiene, decl_macro)]

#[macro_use] extern crate rocket;
#[macro_use] extern crate diesel;
extern crate dotenv;
extern crate chrono;

pub mod models;
pub mod controls;
mod lib;

use std::thread;

#[get("/")]
fn index() -> &'static str {
    "Hello, world!"
}

fn main() {
    use self::lib::{update_brewery_name, get_brewery_name, establish_connection};

    let connection = establish_connection();
    let brewery_name = get_brewery_name(&connection);
    
    match brewery_name {
        Some(x) => println!("Brewery name is {}", x),
        None => update_brewery_name(&connection, None),
    }

    thread::spawn(|| {
        controls::pid_loop();
    });
    rocket::ignite().mount("/", routes![index]).launch();
}

