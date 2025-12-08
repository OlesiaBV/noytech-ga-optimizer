-- Создание таблицы shipments
CREATE TABLE shipments (
    id TEXT PRIMARY KEY,
    weight_kg NUMERIC(10, 3) NOT NULL CHECK (weight_kg > 0),
    volume_m3 NUMERIC(10, 3) NOT NULL CHECK (volume_m3 > 0),
    destination_city TEXT NOT NULL,
    date DATE NOT NULL
);

CREATE INDEX idx_shipments_destination_city ON shipments(destination_city);
