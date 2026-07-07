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

Платформа использует **локальную копию** справочников Delivery Sapiens (`backend/data/delivery_mirror.db`). В PostgreSQL ходит только **ежемесячное обновление** зеркала (нужен VPN); повседневная работа и кнопка **Delivery** на дашборде читают **только локальный файл**.

**DBeaver (без VPN):** подключения `Delivery Mirror (Gallup Q14 local)` и `Gallup Q14 app (local)` — SQLite. Подробнее: [docs/02-setup.md](docs/02-setup.md#локальная-копия-delivery-без-vpn). Пример SQL: [scripts/gallup_delivery_mirror_queries.sql](scripts/gallup_delivery_mirror_queries.sql).

### Что хранится в зеркале

| Таблица-источник (Delivery) | Локальная таблица | Назначение |
|-------------------------------|-------------------|------------|
| `ods.v_employee` + `ods.employee` | `mirror_v_employee`, `mirror_employee` | Штат компании, оргполя формы |
| `vdm_query.v_data_mart_without_total` (текущий квартал) | `mirror_data_mart` | Delivery-активность, band нагрузки |

В SQLite приложения сохраняются только **агрегаты и справочники** — без email и ФИО.

### Переменные окружения

| Переменная | Описание | Когда нужна |
|------------|----------|-------------|
| `DELIVERY_MIRROR_PATH` | Путь к зеркалу (опционально) | По умолчанию `backend/data/delivery_mirror.db` |
| `DELIVERY_SAPIENS_DB_*` | PostgreSQL Delivery | **Только** `pull_delivery_mirror.py` (VPN) |
| `DELIVERY_SYNC_INTERVAL_HOURS` | Пересборка справочника из зеркала, ч; `0` = выкл. | Backend (без VPN) |

Требуется **Python 3**; для pull — `pip install psycopg2-binary`.

### Ежемесячное обновление зеркала (VPN)

**Pilot (VPS):** зеркало живёт **на сервере** (`/opt/gallup-q14/data/delivery_mirror.db`). Backend и кнопка Delivery читают его локально — доступ снаружи не нужен.

Обновление зеркала из PostgreSQL делается **с рабочей машины в VPN**, затем файл заливается на VPS (`INTERXION_SWI_*` в user env, см. внутренний runbook):

```powershell
powershell -File scripts/delivery-monthly-sync.ps1
# pull → local mirror → sync → export → upload mirror.db → sync on VPS
```

Или только upload уже скачанного зеркала:

```powershell
powershell -File scripts/upload-mirror-to-vps.ps1
```

**Локальная разработка:** то же зеркало в `backend/data/delivery_mirror.db` (не коммитится).

### Повседневная синхронизация (без VPN)

```powershell
python scripts/sync_delivery_reference.py
```

Либо кнопка **Delivery** на HR-дашборде — пересобирает справочник из **локального зеркала**.

### Автоматическая синхронизация backend

При `DELIVERY_SYNC_INTERVAL_HOURS > 0` (по умолчанию **24**) backend периодически запускает `sync_delivery_reference.py` **из зеркала**. PostgreSQL не используется. Если зеркала нет — sync пропускается (см. лог).

```powershell
$env:DELIVERY_SYNC_INTERVAL_HOURS = "0"   # только вручную / кнопка
```

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

Если VPS **не видит** PostgreSQL Delivery, справочник обновляется **офлайн** с машины в VPN. На сервере backend читает локальное зеркало; авто-sync из зеркала на VPS обычно отключён (`DELIVERY_SYNC_INTERVAL_HOURS=0`).

**На машине с VPN** (раз в месяц или после изменений в Delivery):

```powershell
powershell -File scripts/delivery-monthly-sync.ps1
```

Скрипт: pull → локальное зеркало → sync app DB → export seed → **upload `delivery_mirror.db` на VPS** → sync на сервере. Параметры SSH — user env `INTERXION_SWI_*` (см. внутренний runbook деплоя).

Аварийный fallback без зеркала: `export_delivery_reference_sql.py` + `apply_delivery_reference.sh` на VPS (см. `scripts/apply_delivery_reference.sh`).

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
