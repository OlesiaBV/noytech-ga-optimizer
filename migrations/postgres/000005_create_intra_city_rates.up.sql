-- Создание таблицы intra_city_rates (внутригород)
CREATE TABLE intra_city_rates (
    volume_m3 NUMERIC(5,1) NOT NULL,
    weight_tons NUMERIC(5,1) NOT NULL,
    rate_fixed NUMERIC(8,2) NOT NULL,
    PRIMARY KEY (volume_m3, weight_tons)
);