   
use diesel::prelude::*;
use dotenv::dotenv;
use std::env;
pub mod schema;

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
    if s.len() == 0 {
        s = "Elsinore".to_string();
    }
    return s;
}

pub fn get_brewery_name(connection: &SqliteConnection) -> Option<String> {
    use schema::settings::dsl::*;
    
    let str_bname = settings.select(brewery_name).load::<String>(connection);

    let mut bname = None;

    match str_bname {
        Ok(x) => if x.len() > 0 {
            bname = Some(x.last().cloned().unwrap());
        },
        Err(x) => println!("Error: {:?}", x),
    }

    return bname;
}

pub fn update_brewery_name(connection: &SqliteConnection, new_name: Option<String>) {
    use schema::settings::dsl::*;
    let new_brewery_name = match new_name {
        Some(x) => x,
        None => read_brewery_name(),
    };

    let rows_inserted = diesel::insert_into(settings)
        .values(&brewery_name.eq(new_brewery_name)).execute(connection);

    if Ok(1) == rows_inserted {
        println!("Updated!")
    } else {
        println!("Failed to insert!");
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn init() -> SqliteConnection {
        dotenv::dotenv().ok();
        let database_url = std::env::var("TEST_DATABASE_URL").expect("TEST_DATABASE_URL must be set");
        let db = SqliteConnection::establish(&database_url)
            .expect(&format!("Error connecting to {}", database_url));
        db.begin_test_transaction().unwrap();
        db
    }

    #[test]
    fn update_brewery_name_reads_input() {
        use self::schema::settings::dsl::*;
        use chrono::Utc;

        let connection = init();
        let new_name = format!("NewName {}", Utc::now());
        update_brewery_name(&connection, Some(new_name.to_string()));

        let result = settings.select(brewery_name).load::<String>(&connection);

        assert!(result.is_ok(), format!("{:?}", result));
        assert_eq!(new_name, result.ok().unwrap()[0]);
        
    }
}
