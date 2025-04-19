CREATE TABLE IF NOT EXISTS receptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    date_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    pvz_id UUID REFERENCES pvz(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'in_progress' CHECK (status IN ('in_progress', 'closed'))
);

CREATE INDEX IF NOT EXISTS idx_receptions_pvz_id ON receptions(pvz_id);
CREATE INDEX IF NOT EXISTS idx_receptions_status ON receptions(status);
CREATE INDEX IF NOT EXISTS idx_receptions_date_time ON receptions(date_time);
