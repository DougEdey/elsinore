table! {
    one_wire_devices (id) {
        id -> Integer,
        address -> Text,
        #[sql_name = "type"]
        type_ -> Text,
    }
}

table! {
    settings (brewery_name) {
        brewery_name -> Text,
    }
}

allow_tables_to_appear_in_same_query!(
    one_wire_devices,
    settings,
);
