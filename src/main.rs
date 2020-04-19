#![feature(proc_macro_hygiene, decl_macro)]

#[macro_use] extern crate rocket;
#[macro_use] extern crate diesel;
extern crate dotenv;


pub mod schema;
pub mod models;
pub mod controls;

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
    use self::schema::settings::dsl::*;
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

fn update_brewery_name(connection: SqliteConnection) {
    use self::schema::settings::dsl::*;
    use std::io::{stdin,stdout,Write};
    let mut s = String::new();

    println!("Welcome to Elsinore! What would you like to call your brewery?");
    println!("(default: Elsinore)");
    print!("> ");

    let _=stdout().flush();
    stdin().read_line(&mut s).expect("Defaulting to Elsinore.");

    if let Some('\n')=s.chars().next_back() {
        s.pop();
    }
    if let Some('\r')=s.chars().next_back() {
        s.pop();
    }

    let rows_inserted = diesel::insert_into(settings)
        .values(&brewery_name.eq(s)).execute(&connection);

    if Ok(1) == rows_inserted {
        println!("Updated!")
    } else {
        println!("Failed to insert!");
    }
}