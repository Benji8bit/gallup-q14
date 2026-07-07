# Gallup Q14 — опрос вовлечённости Sapiens Solutions

MVP платформы для **ежеквартального анонимного мониторинга вовлечённости** сотрудников консалтинговой компании [Sapiens Solutions](https://sapiens.solutions/).

## Возможности

- Веб-форма опроса: **Gallup Q12 (DE-консалтинг) + eNPS**, шкала согласия 1–5
- Полная **анонимность** (токен в localStorage только для защиты от повторной отправки)
- Накопление ответов в **SQLite** по квартальным раундам
- **HR-дашборд**: тренды вовлечённости и eNPS, radar по блокам, срезы по Delivery, рекомендации
- **CSV-экспорт** для передачи в HR
- Оформление в стиле Sapiens: тёмный фон, зелёные акценты

## Быстрый старт

### Требования

- Go 1.22+
- Node.js 20+
- npm

### 1. Backend

```powershell
cd backend
$env:ADMIN_PASSWORD = "change-me"
$env:PORT = "8080"
$env:CORS_ORIGIN = "http://localhost:5173"
go run ./cmd/server
```

### 2. Frontend (режим разработки)

```powershell
cd frontend
npm install
npm run dev
```

Откройте http://localhost:5173

### 3. Production-like (один сервер)

```powershell
cd frontend
npm run build
cd ../backend
$env:ADMIN_PASSWORD = "change-me"
go run ./cmd/server
```

Откройте http://localhost:8080 — backend отдаёт собранный frontend.

## Доступы (MVP)

| Роль | URL | Пароль |
|------|-----|--------|
| Сотрудник | `/survey` | — |
| HR / руководство | `/admin` | значение `ADMIN_PASSWORD` |

> В production обязательно смените пароль и используйте HTTPS.

## Методология опроса (Q12 DE + eNPS)

С 06.07.2026 опрос соответствует **ТЗ руководства** (А. Гельмут): адаптированный Gallup Q12 для data engineering консалтинга и быстрый индекс **eNPS**.

### Состав опроса

| Блок | Вопросы | Шкала |
|------|---------|-------|
| Удовлетворённость | Q00 | 1–5 |
| Роль и ресурсы | Q01–Q03 | 1–5 (согласие) |
| Признание и поддержка | Q04–Q06 | 1–5 |
| Голос, смысл и команда | Q07–Q10 | 1–5 |
| Обратная связь и развитие | Q11–Q12 | 1–5 |
| eNPS | E01 | 0–10 |

### Как считаются KPI на дашборде

| Метрика | Формула | Пороги |
|---------|---------|--------|
| **Индекс вовлечённости** | % ответов 4–5 по Q01–Q12 | < 35% — зона риска |
| **Удовлетворённость (Q00)** | Среднее 1–5 | Отдельно от вовлечённости |
| **eNPS (E01)** | % промоутеров (9–10) − % критиков (0–6) | > +30 хорошо, < 0 тревожно |

На дашборде также доступны:

- **Тренд eNPS** по кварталам
- **Срезы eNPS** по группам Delivery (направление, грейд, тип занятости и др.)
- Сегментация текущего квартала: промоутеры / нейтралы / критики

### Исторические данные

Ответы, собранные до перехода на шкалу 1–5 (балл 6 в старой форме), автоматически пересчитываются при аналитике. Вопросы Q13, Q14, S01, S02 сняты с опроса, но сохранены в БД.

Подробнее — в разделе **Методология опроса** выше.

## Интеграция с Delivery Sapiens

Вся работа со справочниками идёт **через локальное зеркало** (`backend/data/delivery_mirror.db`), не напрямую из PostgreSQL. В PostgreSQL ходит только **ежемесячный pull** зеркала (VPN).

| Слой | Где | Содержимое | PII |
|------|-----|------------|-----|
| **Зеркало** | рабочая машина (dev) | `v_employee`, `employee`, data mart | email (служебно) |
| **App DB** | dev + VPS | агрегаты, справочники формы | **нет** |
| **Seed SQL** | VPS | экспорт агрегатов для apply | **нет** |

**DBeaver (без VPN):** `Delivery Mirror` → `delivery_mirror.db`, `Gallup Q14 app` → `gallup-q14.db`. Подробнее: [docs/02-setup.md](docs/02-setup.md#локальная-копия-delivery-без-vpn).

### Что хранится в зеркале

| Таблица-источник (Delivery) | Локальная таблица | Назначение |
|-------------------------------|-------------------|------------|
| `ods.v_employee` + `ods.employee` | `mirror_v_employee`, `mirror_employee` | Штат компании, оргполя формы |
| `vdm_query.v_data_mart_without_total` (текущий квартал) | `mirror_data_mart` | Delivery-активность, band нагрузки |

### Переменные окружения

| Переменная | Описание | Когда нужна |
|------------|----------|-------------|
| `DELIVERY_MIRROR_PATH` | Путь к зеркалу | Dev: sync из зеркала |
| `DELIVERY_REFERENCE_SEED_PATH` | Путь к seed SQL | VPS: кнопка Delivery |
| `DELIVERY_SAPIENS_DB_*` | PostgreSQL Delivery | **Только** `pull_delivery_mirror.py` (VPN) |
| `DELIVERY_SYNC_INTERVAL_HOURS` | Пересборка из зеркала, ч; `0` = выкл. | Dev backend |

Требуется **Python 3**; для pull — `pip install psycopg2-binary`.

### Ежемесячный цикл (VPN, рабочая машина)

```powershell
powershell -File scripts/delivery-monthly-sync.ps1
# pull → зеркало → sync app DB → export seed → upload seed на VPS
```

Или только upload уже собранного seed:

```powershell
python scripts/sync_delivery_reference.py
python scripts/export_delivery_reference_sql.py
powershell -File scripts/upload-reference-to-vps.ps1
```

### Повседневная работа (без VPN)

**Dev:** `python scripts/sync_delivery_reference.py` или кнопка **Delivery** — пересборка из **зеркала**.

**VPS (pilot):** зеркала **нет**. Кнопка **Delivery** переприменяет `delivery_reference_seed.sql` (агрегаты без email). Обновление данных — только через monthly upload с рабочей машины.

### Автоматическая синхронизация backend (dev)

При `DELIVERY_SYNC_INTERVAL_HOURS > 0` backend периодически запускает sync **из зеркала**. На VPS авто-sync отключён (`0`); seed обновляется upload-скриптом.

### Охват опроса

На дашборде KPI **«Охват (штат компании)»** = участники опроса / **актуальный штат компании** на дату синхронизации Delivery (`ods.v_employee`: последняя запись по человеку, открытый период `date_to ≥ сегодня`, не уволен). Справочники в форме и знаменатели срезов считаются по тому же составу.

Отдельно в блоке «Контекст Delivery» показывается **Delivery-активность (квартал)** — люди с `mandays > 0` в data mart (для сравнения, не для знаменателя опроса).

### Справочник оргструктуры (1:1 с Delivery)

Скрипт `sync_delivery_reference.py` загружает справочник из `ods.v_employee` + `ods.employee`:

| Поле в опросе | Источник Delivery | Примечание |
|---------------|-------------------|------------|
| Направление | `employee_direction` | актуальный штат компании |
| Должность | `position` | **каскад:** только должности выбранного направления + «Другая должность» |
| Грейд | `grade` | **каскад** по направлению, коды как в Delivery (K1, DE3, …) |
| Тип занятости | `employee_type` | актуальный штат |
| Стаж | `date_from` | три корзины: &lt;1 год, 1–3 года, 3+ лет |

В форме опроса **нет чисел в скобках** — только подписи. Сначала выбирается направление, затем сужаются списки должности и грейда (матрица `direction × position/grade` в `delivery_org_option_scopes`).

Отдельного среза «руководство» нет: директора и прочие роли попадают в те же измерения, что заведены в Delivery (например, направление Backoffice, должность из справочника, грейд из `grade`).

На дашборде срезы показывают **все группы с ответами** (без лимита «топ-8»). После изменения скрипта нужна повторная синхронизация Delivery.

### Ручной перенос справочника Delivery (pilot)

VPS **не хранит зеркало** (нет email). На сервер попадает только **`delivery_reference_seed.sql`** — агрегаты после sync из зеркала на машине в VPN.

```powershell
powershell -File scripts/delivery-monthly-sync.ps1
```

Параметры SSH — user env `INTERXION_SWI_*` (см. внутренний runbook деплоя).

> Pilot: self-signed TLS на `:8443` — при первом входе браузер покажет предупреждение (см. [docs/06-admin-guide.md](docs/06-admin-guide.md)).

## Структура репозитория

```
gallup_q14/
├── backend/          # Go API + SQLite + аналитика
├── frontend/         # React + Vite + Tailwind
├── scripts/          # ETL Delivery Sapiens → SQLite
├── docs/             # Документация
├── .env.example      # Пример переменных окружения
└── README.md
```

## Документация

| Документ | Описание |
|----------|----------|
| [docs/02-setup.md](docs/02-setup.md) | Установка, локальное зеркало Delivery, DBeaver |
| [docs/06-admin-guide.md](docs/06-admin-guide.md) | Руководство HR и администратора |

## Лицензия

Внутренний продукт Sapiens Solutions.
