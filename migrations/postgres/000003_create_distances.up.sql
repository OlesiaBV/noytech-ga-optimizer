-- Матрица кратчайших расстояний (МКР)
CREATE TABLE distances (
    from_city TEXT NOT NULL,
    to_city TEXT NOT NULL,
    km INTEGER NOT NULL CHECK (km >= 0),
    UNIQUE (from_city, to_city)
);
-- Индексы для быстрого поиска расстояний от/до города
CREATE INDEX idx_distances_from_city ON distances(from_city);
CREATE INDEX idx_distances_to_city ON distances(to_city);