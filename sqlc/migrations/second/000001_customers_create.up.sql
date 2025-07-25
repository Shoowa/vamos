CREATE TABLE IF NOT EXISTS customers (
    id UUID PRIMARY KEY DEFAULT generate_uuidv7()::uuid,
    forename text NOT NULL,
    surname text NOT NULL,
    dob date,
    male bool
);
