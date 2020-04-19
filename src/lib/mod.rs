
use diesel::prelude::*;
pub mod schema;

fn read_brewery_name() -> String {
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
   return s;
}

pub fn update_brewery_name(connection: SqliteConnection) {
    use schema::settings::dsl::*;
    let new_brewery_name = read_brewery_name();
    let rows_inserted = diesel::insert_into(settings)
        .values(&brewery_name.eq(new_brewery_name)).execute(&connection);

    if Ok(1) == rows_inserted {
        println!("Updated!")
    } else {
        println!("Failed to insert!");
    }
}