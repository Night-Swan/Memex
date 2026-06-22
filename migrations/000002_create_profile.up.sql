CREATE TABLE profile (
    id UUID PRIMARY KEY,
    weight_kg NUMERIC,
    height_cm NUMERIC,
    age INTEGER,
    sex TEXT,
    activity_level TEXT,
    goals TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);