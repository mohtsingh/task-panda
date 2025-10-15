
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

CREATE TABLE IF NOT EXISTS tasks (
                                     id SERIAL PRIMARY KEY,
                                     category TEXT NOT NULL,
                                     title TEXT NOT NULL,
                                     description TEXT,
                                     budget NUMERIC,
                                     location TEXT,
                                     date DATE
);

-- Updated tasks table with additional columns
ALTER TABLE tasks 
ADD COLUMN created_by INTEGER REFERENCES profiles(id),
ADD COLUMN status TEXT DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'ACCEPTED', 'IN_PROGRESS', 'COMPLETED', 'CANCELLED')),
ADD COLUMN accepted_provider_id INTEGER REFERENCES profiles(id),
ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;


ALTER TABLE profiles 
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
CREATE TRIGGER update_profiles_updated_at BEFORE UPDATE ON profiles FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS device_tokens (
    id SERIAL PRIMARY KEY,
    profile_id INTEGER NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    token TEXT NOT NULL,
    platform VARCHAR(20),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_device_tokens_profile ON device_tokens(profile_id, platform, is_active);

