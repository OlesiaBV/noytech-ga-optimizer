# NOYTECH GA Optimizer
Двухуровневый генетический алгоритм для оптимизации сети филиалов и маршрутов лайнхолов компании NOYTECH

## Быстрый запуск
# Сборка и запуск всех сервисов (PostgreSQL + миграции + приложение)

```
docker-compose up --build
```

После запуска:
- Сервер будет доступен на http://localhost:8080
- База данных PostgreSQL запустится на порту 5432

## API
Сервис предоставляет два HTTP-эндпоинта:

### 1. Загрузка данных
POST /upload
Content-Type: multipart/form-data
- Отправьте два файла:
  - stat.xlsx — грузы, терминалы, тарифы
  - filled_distances_MKR.xlsx — матрица расстояний

### 2. Запуск оптимизации
POST /optimize
Content-Type: application/json
Пример тела запроса:
```
{
  "delivery_days": ["wed", "fri"],
  "ga_settings_level_1": {
    "num_generations": 50,
    "num_individuals": 100,
    "selection_type": "1",
    "crossover_type": "1",
    "mutation_type": "1",
    "stopping_criterion": 5
  }
}
```

Ответ:
- 200 OK — оптимизация успешна
- 400 Bad Request — ошибка валидации (некорректные дни, параметры ГА и т.д.)
- 500 Internal Server Error — ошибка при выполнении ГА

## Структура проекта
```
.
├── cmd/app/                # main.go
├── internal/
│   ├── handler/            # HTTP-обработчики
│   ├── services/
│   │   ├── optimizer/      # Логика оптимизации (GA Level 1)
│   │   └── importer/       # Импорт данных из Excel
│   └── storages/           # Работа с PostgreSQL
├── migrations/             # SQL-миграции (через утилиту migrate)
├── api/proto/              # .proto файлы и сгенерированный код
└── docker-compose.yml      # Запуск PostgreSQL + приложения
```

## Важно
- В Excel-файлах (`stat.xlsx`, `filled_distances_MKR.xlsx`) обязательно соблюдайте названия листов и колонок
