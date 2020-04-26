extern crate embedded_hal as hal;

use std::{fs,io,num,fmt};
use std::path::PathBuf;
use crate::controls::w1_errors::*;


static W1_PATH_PREFIX: &str = "/sys/bus/w1/devices";
static W1_PATH_SUFFIX: &str = "w1_slave";

pub struct MilliCelsius(f64);

impl MilliCelsius {
    pub fn to_fahrenheit(&self) -> f64 {
        (self.0) / 5.0 * 9.0 + 32.0
    }

    pub fn as_celsius(&self) -> f64 {
        self.0
    }
}

impl fmt::Display for MilliCelsius {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "{}", self.as_celsius())
    }
}

pub struct TempProbe {
    address: String,
    temp_c: Option<MilliCelsius>,
}

impl TempProbe {
    pub fn new(address: String) -> TempProbe {
        TempProbe {address, temp_c: None }
    }

    pub fn read_raw(&self) -> io::Result<String> {
        let mut path = PathBuf::from(W1_PATH_PREFIX);
        path.push(&self.address);
        path.push(W1_PATH_SUFFIX);
        fs::read_to_string(path)
    }

    pub fn read_temp(&self) -> Result<MilliCelsius, W1Error> {
        let temp_data = self.read_raw()?;
        if !temp_data.contains("YES") {
            return Err(W1Error::BadSerialConnection);
        }
        Ok(MilliCelsius(parse_temp(temp_data)?))
    }

    pub fn update_temp(&mut self) {
        self.temp_c = Some(self.read_temp().unwrap());
    }
}

impl fmt::Display for TempProbe {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        match &self.temp_c {
            Some(x) => write!(f, "address: {}, temperature: {} C", self.address, x.as_celsius()),
            None => write!(f, "address: {}", self.address),
        }        
    }
}


fn parse_temp(temp_data: String) -> Result<f64, num::ParseIntError> {
    let (_, temp_str) = temp_data.split_at(temp_data.find("t=").unwrap() + 2);
    let temp_u = temp_str.trim().parse::<u32>();
    return Ok((temp_u? as f64) / 1000.0);
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_temp() {
        let temp_data ="6e 01 55 05 7f 7e a5 66 f2 : crc=f2 YES
6e 01 55 05 7f 7e a5 66 f2 t=22875".to_string();
        assert_eq!(Ok(22875), parse_temp(temp_data));
    }
}