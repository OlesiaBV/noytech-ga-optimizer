-- Создание таблицы inter_city_rates (межгород)
CREATE TABLE inter_city_rates (
    volume_m3 NUMERIC(5,1) NOT NULL,
    weight_tons NUMERIC(5,1) NOT NULL,
    rate_per_km NUMERIC(8,2) NOT NULL,
    PRIMARY KEY (volume_m3, weight_tons)
);