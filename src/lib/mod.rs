pub mod schema;

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

pub fn get_brewery_name() -> Option<String> {
    use self::schema::settings::dsl::*;
    
    let connection = establish_connection();
    let str_bname  = settings.select(brewery_name).load::<String>(&connection);

    let mut bname = None;

    match str_bname {
        Ok(x) => if x.len() > 0 {
            bname = Some(x.last().cloned().unwrap());
        },
        Err(x) => println!("Error: {:?}", x),
    }

    return bname;
}
pub fn update_brewery_name(new_name: Option<String>) {
    use schema::settings::dsl::*;
    let connection = establish_connection();
    let new_brewery_name = match new_name {
        Some(x) => x,
        None => read_brewery_name(),
    };

    let rows_inserted = diesel::insert_into(settings)
        .values(&brewery_name.eq(new_brewery_name)).execute(&connection);

    if Ok(1) == rows_inserted {
        println!("Updated!")
    } else {
        println!("Failed to insert!");
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn update_brewery_name_reads_input() {
        use diesel::result::Error;
        use self::schema::settings::dsl::*;

        let connection = establish_connection();
        connection.test_transaction::<_, Error, _>(|| {
            update_brewery_name(Some("NewName".to_string()));

            let result = settings.select(brewery_name).load::<String>(&connection)?;

            assert_eq!("NewName", result[0]);
            Ok(1);
        });
    }
}