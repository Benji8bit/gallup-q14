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

Платформа подтягивает **оргструктуру и контекст нагрузки** из корпоративной PostgreSQL-базы Delivery Sapiens (read-only). В SQLite сохраняются только **агрегаты и справочники** — без email и ФИО.

### Что синхронизируется

| Источник (Delivery) | Назначение в Gallup Q14 |
|---------------------|-------------------------|
| `ods.employee` | Направления, должности, грейды, тип занятости, стаж |
| `vdm_query.v_data_mart_without_total` | Ожидаемый охват, band нагрузки (mandays) |

Справочники используются в **форме опроса** (выпадающие списки) и на **HR-дашборде** (охват, срезы по группам, контекст Delivery).

### Переменные окружения

Задайте в системных переменных Windows (или в `.env` при локальном запуске):

| Переменная | Описание | Пример |
|------------|----------|--------|
| `DELIVERY_SAPIENS_DB_USER` | Пользователь PostgreSQL | `<db-user>` |
| `DELIVERY_SAPIENS_DB_PASSWORD` | Пароль | *(секрет)* |
| `DELIVERY_SAPIENS_DB_HOST` | Хост (опционально) | `<delivery-host>` |
| `DELIVERY_SAPIENS_DB_PORT` | Порт (опционально) | `5432` |
| `DELIVERY_SAPIENS_DB_NAME` | База (опционально) | `postgres` |
| `DELIVERY_SYNC_INTERVAL_HOURS` | Интервал авто-синхронизации, часы; `0` = выкл. | `24` |

Требуется **Python 3** и пакет `psycopg2-binary` для скрипта синхронизации.

### Ручная синхронизация

```powershell
pip install psycopg2-binary
python scripts/sync_delivery_reference.py
```

Либо кнопка **Delivery** на HR-дашборде (`POST /api/admin/sync-delivery`).

### Автоматическая синхронизация

При запуске backend с заданным `DELIVERY_SAPIENS_DB_PASSWORD` и `DELIVERY_SYNC_INTERVAL_HOURS > 0` (по умолчанию **24**):

1. Синхронизация выполняется **сразу после старта** сервера
2. Затем **каждые N часов** по расписанию

Логи в консоли backend: `delivery sync ok: ...` или `delivery sync error: ...`.

Чтобы отключить расписание и синхронизировать только вручную:

```powershell
$env:DELIVERY_SYNC_INTERVAL_HOURS = "0"
```

### Охват опроса

На дашборде KPI **«Охват (Delivery штат)»** = участники опроса / число сотрудников в `ods.employee`. Знаменатель обновляется при каждой синхронизации.

### Справочник оргструктуры (1:1 с Delivery)

Скрипт `sync_delivery_reference.py` загружает в форму опроса и дашборд **полный** справочник из `ods.employee` без усечения:

| Поле в опросе | Источник Delivery | Примечание |
|---------------|-------------------|------------|
| Направление | `employee_direction` | все значения |
| Должность | `position` | **все** должности (раньше — топ-20) |
| Грейд | `grade` | **как в Delivery** (K1, DE3, …), без свёртки в Junior/Lead |
| Тип занятости | `employee_type` | все значения |
| Стаж | `date_from` | три корзины: &lt;1 год, 1–3 года, 3+ лет |

Отдельного среза «руководство» нет: директора и прочие роли попадают в те же измерения, что заведены в Delivery (например, направление Backoffice, должность из справочника, грейд из `grade`).

На дашборде срезы показывают **все группы с ответами** (без лимита «топ-8»). После изменения скрипта нужна повторная синхронизация Delivery.

### Ручной перенос справочника Delivery (pilot)

Если VPS **не видит** PostgreSQL Delivery (`<delivery-host>`), справочник загружается **офлайн** с машины в VPN. Кнопка **Delivery** на дашборде остаётся (решение по ней — позже); авто-sync на сервере отключён (`DELIVERY_SYNC_INTERVAL_HOURS=0`).

**На машине с доступом к Delivery:**

```powershell
pip install psycopg2-binary
$env:DB_PATH = "C:\Work\gallup_q14\backend\data\gallup-q14.db"
python scripts/sync_delivery_reference.py
python scripts/export_delivery_reference_sql.py
```

Получится `scripts/delivery_reference_seed.sql` (таблицы `delivery_org_options`, `delivery_context_stats`, `delivery_sync_meta`).

**На interxion:**

```powershell
$hostkey = "<vps-ssh-hostkey-fingerprint>"
$pass = $env:VPS_SSH_PASSWORD

pscp -batch -hostkey $hostkey -pw $pass `
  scripts/delivery_reference_seed.sql root@<vps-host>:/opt/gallup-q14/scripts/
pscp -batch -hostkey $hostkey -pw $pass `
  scripts/apply_delivery_reference.sh root@<vps-host>:/opt/gallup-q14/scripts/

plink -batch -hostkey $hostkey -pw $pass root@<vps-host> `
  "bash /opt/gallup-q14/scripts/apply_delivery_reference.sh /opt/gallup-q14/scripts/delivery_reference_seed.sql"
```

Скрипт делает backup БД, останавливает `gallup-q14`, применяет SQL, запускает сервис. Ответы опроса не затрагиваются.

**Снимок на 06.07.2026:** полный справочник из Delivery.

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
| [docs/02-setup.md](docs/02-setup.md) | Установка и запуск |
| [docs/06-admin-guide.md](docs/06-admin-guide.md) | Руководство HR и администратора |

## Лицензия

Внутренний продукт Sapiens Solutions.
