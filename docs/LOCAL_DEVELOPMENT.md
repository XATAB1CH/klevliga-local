# Локальная разработка на Windows

Для запуска нужны Node.js LTS с npm, Go 1.26+ и PostgreSQL 17. Для Go race detector нужен MinGW-w64/GCC. Frontend-тесты используют Vitest, Testing Library и Playwright Chromium; версии закреплены в package-lock.json.

В текущем окружении Node.js и Go установлены через winget. Дополнительный Go 1.26 и PostgreSQL находятся в `%LOCALAPPDATA%\KlevLiga\tools`. Скрипты автоматически добавляют инструменты в PATH своего процесса. Docker и WSL для этого варианта запуска не требуются; существующий `infra/docker-compose.yml` остаётся альтернативой для PostgreSQL.

## Запуск

Из корня проекта:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File scripts/Start-LocalDatabase.ps1
```

Скрипт создаёт базы `klevliga` и `klevliga_test`, если их ещё нет. Данные хранятся в `%LOCALAPPDATA%\KlevLiga\postgres`, вне OneDrive. Сервер слушает только `127.0.0.1:5434`. Случайный локальный пароль сохраняется в `connection.json` в той же папке. При первом запуске создаётся игнорируемый Git файл `services/api/config.yaml`; существующий конфиг не перезаписывается. Не добавляйте эти файлы в Git и не публикуйте их содержимое.

В двух отдельных терминалах:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File scripts/Run-Api.ps1
```

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File scripts/Run-Web.ps1
```

Интерфейс: http://127.0.0.1:5173. API: http://127.0.0.1:8080/api/v1. Завершение каждого процесса — Ctrl+C. Остановить PostgreSQL:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File scripts/Stop-LocalDatabase.ps1
```

`ExecutionPolicy Bypass` применяется только к запущенному процессу и не меняет системную политику выполнения PowerShell.

## Проверки

```powershell
cd apps/web
npm.cmd ci
npm.cmd run test
npm.cmd run build
npm.cmd exec playwright install chromium
npm.cmd run test:e2e
```

Playwright сам запускает Vite и проверяет Chromium на ширинах 393 и 1440 px. UI-регрессии используют типизированные тестовые ответы API; интеграционные backend-тесты обращаются к настоящей PostgreSQL.

Из корня проекта при запущенной базе:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File scripts/Test-Api.ps1
powershell.exe -NoProfile -ExecutionPolicy Bypass -File scripts/Test-Api.ps1 -Race
```

Для другой тестовой PostgreSQL задайте `KLEVLIGA_TEST_DATABASE_URL` в окружении. Каждый интеграционный тест создаёт изолированную схему и удаляет только её. База приложения для тестов не используется. Для CI задавайте URL явно: без URL Go пропускает интеграционные тесты, а локальный скрипт требует существующие настройки базы.

Если новый терминал не видит установленные через winget Node.js или Go, перезапустите терминал либо используйте приведённые скрипты. GNU Make не обязателен: PowerShell-скрипты покрывают запуск и тесты без shell-синтаксиса Unix.
