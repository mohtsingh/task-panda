CREATE TABLE profiles (
	id SERIAL PRIMARY KEY,
	full_name TEXT NOT NULL,
	email TEXT UNIQUE NOT NULL,
	address TEXT,
	phone_number TEXT,
	bio TEXT,
	role TEXT NOT NULL
);

ALTER TABLE profiles ADD CONSTRAINT unique_email UNIQUE (email);

ALTER TABLE profiles ADD COLUMN photo BYTEA;

