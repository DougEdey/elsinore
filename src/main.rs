#![feature(proc_macro_hygiene, decl_macro)]

#[macro_use] extern crate rocket;
#[macro_use] extern crate diesel;
extern crate dotenv;

pub mod models;
pub mod controls;
mod lib;

use std::thread;
use diesel::prelude::*;




#[get("/")]
fn index() -> &'static str {
    "Hello, world!"
}

fn main() {
    use self::lib::{update_brewery_name, get_brewery_name};

    let brewery_name = get_brewery_name();
    
    match brewery_name {
        Some(x) => println!("Brewery name is {}", x),
        None => update_brewery_name(None),
    }

    thread::spawn(|| {
        controls::pid_loop();
    });
    rocket::ignite().mount("/", routes![index]).launch();
}

