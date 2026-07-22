CREATE TABLE IF NOT EXISTS entry (
    store_id VARCHAR(255) NOT NULL,
    store_name VARCHAR(255) NOT NULL,
    vendor VARCHAR(255) CHECK (vendor IN ('chagee', 'luckin')) NOT NULL,
    raw_data JSONB NOT NULL,
    waiting_cups INT NOT NULL,
    waiting_time INT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    coordinates POINT NOT NULL,
    PRIMARY KEY (store_id, vendor, created_at)
)
