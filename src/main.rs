#![feature(proc_macro_hygiene, decl_macro)]

#[macro_use] extern crate rocket;
#[macro_use] extern crate diesel;
extern crate dotenv;

pub mod models;
pub mod controls;
mod lib;

use std::thread;
use diesel::prelude::*;
use dotenv::dotenv;
use std::env;

pub fn establish_connection() -> SqliteConnection {
    dotenv().ok();

    let database_url = env::var("DATABASE_URL")
        .expect("DATABASE_URL must be set.");
    SqliteConnection::establish(&database_url)
        .expect(&format!("Error connecting to {}", database_url))
}

#[get("/")]
fn index() -> &'static str {
    "Hello, world!"
}

fn main() {
    use self::lib::schema::settings::dsl::*;
    use self::lib::update_brewery_name;

    let connection = establish_connection();
    let str_bname  = settings.select(brewery_name).limit(1).load::<Option<String>>(&connection)
        .expect("Error loading brewery name.");

    let mut bname = None;

    if str_bname.len() > 0 {
        bname = Some(&str_bname);
    } 
    
    match bname {
        Some(x) => println!("Brewery name is {}", x[0].as_ref().unwrap()),
        None => update_brewery_name(connection),
    }

    thread::spawn(|| {
        controls::pid_loop();
    });
    rocket::ignite().mount("/", routes![index]).launch();
}

