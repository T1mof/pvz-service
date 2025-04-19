CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    date_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    type VARCHAR(50) NOT NULL CHECK (type IN ('electronics', 'clothes', 'footwear')),
    reception_id UUID REFERENCES receptions(id) ON DELETE CASCADE,
    sequence_num INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_products_reception_id ON products(reception_id);
CREATE INDEX IF NOT EXISTS idx_products_type ON products(type);
CREATE INDEX IF NOT EXISTS idx_products_sequence_num ON products(sequence_num);
