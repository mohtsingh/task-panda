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

-- Create offers table to store service provider offers
CREATE TABLE IF NOT EXISTS offers (
    id SERIAL PRIMARY KEY,
    task_id INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    provider_id INTEGER NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    offered_price NUMERIC NOT NULL,
    message TEXT,
    status TEXT DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'ACCEPTED', 'REJECTED')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(task_id, provider_id) -- One offer per provider per task
);

-- Create chats table to store individual chat conversations
CREATE TABLE IF NOT EXISTS chats (
    id SERIAL PRIMARY KEY,
    task_id INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    customer_id INTEGER NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    provider_id INTEGER NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    offer_id INTEGER NOT NULL REFERENCES offers(id) ON DELETE CASCADE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(task_id, customer_id, provider_id) -- One chat per customer-provider pair per task
);

-- Create messages table to store individual chat messages
CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    chat_id INTEGER NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    sender_id INTEGER NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    message_text TEXT NOT NULL,
    message_type TEXT DEFAULT 'TEXT' CHECK (message_type IN ('TEXT', 'OFFER_UPDATE', 'SYSTEM')),
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_offers_task_id ON offers(task_id);
CREATE INDEX IF NOT EXISTS idx_offers_provider_id ON offers(provider_id);
CREATE INDEX IF NOT EXISTS idx_chats_task_id ON chats(task_id);
CREATE INDEX IF NOT EXISTS idx_chats_active ON chats(is_active);
CREATE INDEX IF NOT EXISTS idx_messages_chat_id ON messages(chat_id);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);

-- Add trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_tasks_updated_at BEFORE UPDATE ON tasks FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_offers_updated_at BEFORE UPDATE ON offers FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_chats_updated_at BEFORE UPDATE ON chats FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
