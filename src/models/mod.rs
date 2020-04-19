#[macro_use]
extern crate diesel;

use diesel::prelude::*;

table! {
    
}
pub struct TemperatureProbe {
    pub id: i32,
    pub type: String,
}

pub struct OneWireDevice {
    pub id: i32,
    pub address: String,
}
