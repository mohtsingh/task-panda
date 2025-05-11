CREATE TABLE IF NOT EXISTS tasks (
                                     id SERIAL PRIMARY KEY,
                                     category TEXT NOT NULL,
                                     title TEXT NOT NULL,
                                     description TEXT,
                                     budget NUMERIC,
                                     location TEXT,
                                     date DATE
);
