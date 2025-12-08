-- Создание справочника терминалов
CREATE TABLE terminals (
    city TEXT PRIMARY KEY,
    direction TEXT NOT NULL,
    distance_from_moscow_km INTEGER NOT NULL CHECK (distance_from_moscow_km >= 0)
);

CREATE INDEX idx_terminals_direction ON terminals(direction);