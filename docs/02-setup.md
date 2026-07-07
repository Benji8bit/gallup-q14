# Установка и запуск

## Предварительные требования

| Компонент | Версия |
|-----------|--------|
| Go | 1.22+ |
| Node.js | 20+ |
| npm | 10+ |

Проверка:

```powershell
go version
node --version
npm --version
```

## Клонирование и структура

```powershell
cd c:\Work\gallup_q14
```

Скопируйте переменные окружения:

```powershell
copy .env.example .env
```

Отредактируйте `.env` — как минимум задайте надёжный `ADMIN_PASSWORD`.

## Режим разработки (рекомендуется)

Запускайте backend и frontend в **двух терминалах**.

### Терминал 1 — Backend

```powershell
cd backend
$env:ADMIN_PASSWORD = "change-me"
$env:PORT = "8080"
$env:CORS_ORIGIN = "http://localhost:5173"
$env:DB_PATH = "./data/gallup-q14.db"
go run ./cmd/server
```

Ожидаемый вывод:

```
starting server on :8080
```

Проверка health:

```powershell
curl http://localhost:8080/api/health
```

### Терминал 2 — Frontend

```powershell
cd frontend
npm install
npm run dev
```

Откройте http://localhost:5173

Vite проксирует `/api/*` на `http://localhost:8080`.

## Режим «один сервер»

Подходит для демо и локального тестирования production-сборки:

```powershell
cd frontend
npm install
npm run build

cd ..\backend
$env:ADMIN_PASSWORD = "change-me"
go run ./cmd/server
```

Откройте http://localhost:8080

Backend автоматически раздаёт статику из `frontend/dist`.

## Первый проход

1. Откройте http://localhost:5173 (или :8080)
2. Нажмите **Пройти опрос**
3. Заполните все 13 вопросов (E01 + Q01–Q12) и обязательный блок «О вас»
4. Отправьте форму
5. Войдите в `/admin` с паролем `change-me`
6. Проверьте дашборд и экспорт CSV

## Сброс данных

Удалите файл базы:

```powershell
Remove-Item backend\data\gallup-q14.db -ErrorAction SilentlyContinue
```

При следующем запуске миграции создадут схему заново.

## Сборка backend

```powershell
cd backend
go build -o bin/gallup-q14.exe ./cmd/server
.\bin\gallup-q14.exe
```

## Сборка frontend

```powershell
cd frontend
npm run build
```

Артефакты: `frontend/dist/`

## Troubleshooting

### `ADMIN_PASSWORD is required`

Задайте переменную окружения перед запуском сервера.

### CORS-ошибки в браузере

Убедитесь, что `CORS_ORIGIN` совпадает с URL фронтенда (в dev — `http://localhost:5173`).

### «Вы уже отправляли опрос в этом квартале»

Очистите `localStorage.respondent_token` в DevTools или используйте другой браузер/режим инкognito.

### Порт занят

```powershell
$env:PORT = "8081"
```

И обновите proxy в `frontend/vite.config.ts`.

## Локальная копия Delivery (без VPN)

Справочники Delivery хранятся **офлайн** в SQLite-файле. PostgreSQL нужен только при **ежемесячном** обновлении зеркала (корпоративный VPN).

| Файл | Назначение | В git |
|------|------------|-------|
| `backend/data/delivery_mirror.db` | Снимок `v_employee`, `employee`, data mart (квартал) | **нет** |
| `backend/data/gallup-q14.db` | БД приложения (опрос, справочники, ответы) | **нет** |
| `scripts/delivery_reference_seed.sql` | Агрегаты для VPS (после sync из зеркала) | **нет** |

> **VPS:** хранит только seed + app DB — **без зеркала и email**. Зеркало остаётся на рабочей машине.

### Таблицы зеркала

| Таблица | Содержимое |
|---------|------------|
| `mirror_meta` | Дата pull, хост-источник, квартал, счётчики строк |
| `mirror_v_employee` | Снимок `ods.v_employee` |
| `mirror_employee` | Снимок `ods.employee` |
| `mirror_data_mart` | Mandays за текущий квартал |

> В зеркале есть **email** (служебно, для join). В git не коммитить.

### Скрипты

| Скрипт | VPN | Действие |
|--------|-----|----------|
| `pull_delivery_mirror.py` | **да** (рабочая машина) | PostgreSQL → локальный `delivery_mirror.db` |
| `sync_delivery_reference.py` | нет | зеркало → `gallup-q14.db` |
| `export_delivery_reference_sql.py` | нет | app DB → `delivery_reference_seed.sql` |
| `upload-reference-to-vps.ps1` | нет | seed → VPS + apply |
| `apply_delivery_reference.sh` | нет | **на VPS:** seed → app DB |
| `delivery-monthly-sync.ps1` | да | pull + sync + export + upload |
| `register-delivery-monthly-task.ps1` | — | задача Windows (1-е число, 08:00) |

```powershell
pip install psycopg2-binary

# Раз в квартал (VPN) — обновление грейдов DE
powershell -File scripts/delivery-monthly-sync.ps1

# Без VPN — пересборка справочника из уже скачанного зеркала
python scripts/sync_delivery_reference.py
```

Переменные: см. `.env.example` (`DELIVERY_MIRROR_PATH`, `DELIVERY_SAPIENS_DB_*`, `DELIVERY_SYNC_INTERVAL_HOURS`).

### DBeaver

Подключения добавлены в workspace DBeaver (`General/.dbeaver/data-sources.json`):

| Имя в DBeaver | Путь |
|---------------|------|
| **Delivery Mirror (Gallup Q14 local)** | `C:/Work/gallup_q14/backend/data/delivery_mirror.db` |
| **Gallup Q14 app (local)** | `C:/Work/gallup_q14/backend/data/gallup-q14.db` |

После правки `data-sources.json` перезапустите DBeaver или **Database → Refresh**.

Примеры запросов: `scripts/gallup_delivery_mirror_queries.sql` (копия также в `General/Scripts/gallup_q14_delivery_mirror.sql` в DBeaver).

### Первый pull

Если `delivery_mirror.db` ещё нет:

```powershell
$env:DELIVERY_SAPIENS_DB_HOST = "<delivery-host>"   # или User env
python scripts/pull_delivery_mirror.py
python scripts/sync_delivery_reference.py
```

Без зеркала (dev) `sync_delivery_reference.py` и кнопка **Delivery** завершатся ошибкой. На VPS вместо зеркала используется `delivery_reference_seed.sql`.

## Тесты backend

Пакет `backend/internal/analytics` — unit-тесты расчётов дашборда. Требуется Go 1.20+ (полный `go test ./...` — Go 1.22+ из-за зависимости SQLite).

```powershell
cd backend
go test ./internal/analytics/... -v
```

### Группа 1: scoring (`scoring_test.go`)

| Тест | Что проверяет |
|------|----------------|
| `TestBuildScoreContext_*` | Роли вопросов (Q00/eNPS/engagement), шкала 1–5, legacy 1–6 |
| `TestClassifyEnps` | Сегменты eNPS: промоутер / нейтрал / критик |
| `TestBuildEnpsScore` | Формула eNPS = % промоутеров − % критиков |
| `TestEnpsScoreModel` | Маппинг в `models.EnpsScore` |
| `TestEnpsTrendModel` | Точка тренда по кварталу |

### Группа 2: engine (`engine_test.go`)

| Тест | Что проверяет |
|------|----------------|
| `TestPct` / `TestAvg` | Вспомогательные расчёты процентов и среднего |
| `TestMapSixToFive` | Пересчёт legacy-шкалы 1–6 в 1–5 |
| `TestBuildDashboard_empty` | Пустой ввод: заглушки измерений и рекомендаций |
| `TestBuildDashboard_engagementAndEnps` | Сводный KPI, eNPS, тренды, охват Delivery |

### Группа 3: segments (`segments_test.go`)

| Тест | Что проверяет |
|------|----------------|
| `TestResponseRate` | Охват опроса = участники / штат Delivery |
| `TestBuildSegmentBreakdown` | Срезы вовлечённости по direction/grade/… |
| `TestBuildEnpsSegmentBreakdown` | Срезы eNPS по группам Delivery |

### Группа 4: recommendations (`recommendations_test.go`)

| Тест | Что проверяет |
|------|----------------|
| `TestEngagementBand` | Пороги «сильная / внимание / критическая» зона |
| `TestBuildRecommendations_empty` | Заглушка при отсутствии измерений |
| `TestBuildRecommendations_withData` | Общие и dimension-рекомендации |
| `TestRecommendationsSummary` | Краткий список для дашборда |
