# Таблицы Шульте — тестирование устойчивости внимания

Веб-приложение для автоматизации методики исследования внимания «таблицы Шульте»
(практикум «Разработка ПО для автоматизации методик исследования внимания»).

* **Испытуемый** проходит 5 таблиц 5×5 (числа 1…25 в случайном порядке), система измеряет время и ошибки.
* Рассчитываются показатели: эффективность работы **ЭР**, степень врабатываемости **ВР**,
  психическая устойчивость **ПУ**, строится кривая утомляемости.
* **Психолог** ведёт карточки испытуемых, смотрит результаты, динамику, сравнивает тесты и группы,
  настраивает параметры отображения (размер шрифта, цвет фона, секундомер, звук ошибки).

## Структура репозитория

| Папка | Содержимое |
|-------|-----------|
| `backend/`  | Go 1.24, чистая архитектура: `internal/domain` → `internal/usecase` → `internal/adapter` (HTTP, PostgreSQL, память) → `cmd/server` |
| `frontend/` | React 18 + TypeScript + MUI 6 (Vite), графики `@mui/x-charts` |
| `infra/`    | Dockerfile’ы, `docker-compose.yml`, `nginx.conf`, пример `.env` |
| `reports/`  | Отчёты по практическим и лабораторным работам (.docx) |

### Слои backend (правило зависимостей направлено внутрь)

```
cmd/server (composition root)
   └─ adapter/http, adapter/repository/{postgres,memory}, infrastructure/{config,token,platform}
         └─ usecase  (сервисы + порты-интерфейсы)
               └─ domain (сущности, расчёт ЭР/ВР/ПУ, динамика, сравнение)
```

## Быстрый старт (Docker)

```bash
make up                 # или: docker compose -f infra/docker-compose.yml --env-file infra/.env up --build
# http://localhost:8081  — пароль психолога по умолчанию: admin (меняется в infra/.env)
```

## Запуск для разработки

```bash
make backend-run        # backend на :8080, данные в памяти (STORAGE=memory)
make frontend-install   # один раз
make frontend-run       # http://localhost:5173
```

Для работы с PostgreSQL: `STORAGE=postgres DATABASE_URL=postgres://user:pass@host:5432/db?sslmode=disable`.

## Тесты

```bash
make backend-test       # go vet + go test -race -cover
TEST_DATABASE_URL=postgres://... make backend-test   # + интеграционные тесты PostgreSQL
make frontend-test      # tsc + vitest
```

## Конфигурация backend (переменные окружения)

| Переменная | По умолчанию | Назначение |
|-----------|--------------|-----------|
| `HTTP_ADDR` | `:8080` | адрес HTTP-сервера |
| `STORAGE` | `postgres` | `postgres` или `memory` |
| `DATABASE_URL` | — | строка подключения (обязательна для `postgres`) |
| `ADMIN_PASSWORD` | `admin` | пароль режима психолога |
| `TOKEN_SECRET` | `dev-secret-change-me` | секрет подписи токенов |
| `TOKEN_TTL_HOURS` | `12` | срок действия токена |
| `CORS_ORIGIN` | `http://localhost:5173` | разрешённый источник для dev-режима |

> `go.sum` не хранится в репозитории: `make backend-deps` (и сборка в Docker) выполняет `go mod tidy`.

## REST API

| Метод и путь | Доступ | Описание |
|--------------|--------|----------|
| `POST /api/auth/admin` · `POST /api/auth/subject` | все | вход психолога (пароль) / испытуемого (выбор карточки) |
| `GET /api/public/subjects` | все | список для экрана входа |
| `GET/POST /api/subjects`, `GET/PUT/DELETE /api/subjects/{id}` | психолог (GET по id — и сам испытуемый) | карточки испытуемых |
| `GET /api/groups`, `GET /api/groups/{name}/stats` | психолог | группы и сравнение в группе |
| `GET/PUT /api/settings` | все / психолог | параметры тестирования |
| `POST /api/tests`, `POST /api/tests/{id}/finish` | психолог / сам испытуемый | старт теста и фиксация результатов |
| `GET /api/subjects/{id}/results`, `GET /api/results/{id}` | психолог / сам испытуемый | результаты |
| `GET /api/subjects/{id}/dynamics` | психолог / сам испытуемый | динамика |
| `GET /api/compare?a=&b=` | психолог | сравнение двух тестирований |
| `DELETE /api/results/{id}` | психолог | удаление результата |
