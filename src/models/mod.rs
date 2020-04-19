pub struct TemperatureProbe {
    pub id: i32,
    pub probe_type: String,
}

pub struct OneWireDevice {
    pub id: i32,
    pub address: String,
}

#[derive(Queryable)]
pub struct Setting {
    pub brewery_name: Option<String>,
}
