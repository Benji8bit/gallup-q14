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
3. Заполните все 17 вопросов (Q00 + Q01–Q14 + S01–S02)
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
