#![feature(proc_macro_hygiene, decl_macro)]

#[macro_use] extern crate rocket;
#[macro_use] extern crate diesel;
#[macro_use] extern crate diesel_migrations;

//#[cfg(all(feature = "sqlite", not(any(feature = "postgres", feature = "mysql"))))]
embed_migrations!();

pub mod models;
pub mod controls;
mod database_helpers;

use std::thread;

#[get("/")]
fn index() -> &'static str {
    "Hello, world!"
}

fn main() {
    use self::database_helpers::{update_brewery_name, get_brewery_name, establish_connection};
    
    let connection = establish_connection();
    let migration_result = embedded_migrations::run(&connection);
    println!("Migrations {:?}", migration_result);

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

