CREATE TABLE IF NOT EXISTS pvz (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    registration_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    city VARCHAR(50) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_pvz_city ON pvz(city);
