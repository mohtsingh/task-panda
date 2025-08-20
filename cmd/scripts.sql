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

-- Updated tasks table with additional columns
ALTER TABLE tasks 
ADD COLUMN created_by INTEGER REFERENCES profiles(id),
ADD COLUMN status TEXT DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'ACCEPTED', 'IN_PROGRESS', 'COMPLETED', 'CANCELLED')),
ADD COLUMN accepted_provider_id INTEGER REFERENCES profiles(id),
ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;


-- Add trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_tasks_updated_at BEFORE UPDATE ON tasks FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
