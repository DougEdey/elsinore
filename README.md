# Elsinore2.0
StrangeBrew Elsinore re-write in Rust

## Setting Up

I've decided to start using [Rust Rocket](https://rocket.rs/) for this as it seems like a simple playground.

You'll need [Rustup](https://rustup.rs/) for the nightly builds that Rocket requires

## Dependencies

 * `sqlite`, `sqlite3`, `libsqlite3-0` (these are required for diesel-cli)

## Commands to run
 * `rustup default nightly` -> Installs Rust nightly for Rocket
 * `cargo install diesel_cli --no-default-features --features sqlite` --> Just installs diesel-cli for sqlite, you can switch to postgres/mysql if you want
 * `diesel setup` -> Initializes the database
 * `diesel migration run` -> Runs any pending migrations

## Object Relation Mapping (ORM)

I'm trying out [Diesel](https://diesel.rs/guides/getting-started/) as it seems fairly simple, supports migrations, and above all, makes sense to me.

With Diesel, I'm using [dotenv](https://docs.rs/dotenv/0.15.0/dotenv/) which puts the environment variables in the local `.env` file.

By default, the database will be at `elsinore.db`, but this can be customized with `DATABASE_URL` in this file
