use std::{io,num};

#[derive(Debug)]
pub enum W1Error {
    Io(io::Error),
    Parse(num::ParseIntError),
    BadSerialConnection,
}

impl From<io::Error> for W1Error {
    fn from(err: io::Error) -> W1Error {
        W1Error::Io(err)
    }
}

impl From<num::ParseIntError> for W1Error {
    fn from(err: num::ParseIntError) -> W1Error {
        W1Error::Parse(err)
    }
}
