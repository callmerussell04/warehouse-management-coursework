# Система управления складом и логистикой

<div align="center">

Веб-приложение для учета товаров, управления складскими операциями, контрагентами и заказами.

Курсовая работа по дисциплине **«Конструирование программного обеспечения»**<br>
Тема: **«Разработка сайта для управления складом и логистикой»**

![Go](https://img.shields.io/badge/Go-1.24.4-00ADD8?style=flat-square&logo=go&logoColor=white)
![React](https://img.shields.io/badge/React-19.2-61DAFB?style=flat-square&logo=react&logoColor=black)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-17-4169E1?style=flat-square&logo=postgresql&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-8.4-DC382D?style=flat-square&logo=redis&logoColor=white)
![Nginx](https://img.shields.io/badge/Nginx-Reverse_Proxy-009639?style=flat-square&logo=nginx&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=flat-square&logo=docker&logoColor=white)

</div>

## Содержание

- [О проекте](#о-проекте)
- [Возможности](#возможности)
- [Технологический стек](#технологический-стек)
- [Архитектура](#архитектура)
- [Структура репозитория](#структура-репозитория)
- [Быстрый запуск](#быстрый-запуск)
- [Конфигурация](#конфигурация)
- [Локальная разработка](#локальная-разработка)
- [API](#api)
- [Тестирование](#тестирование)
- [Статус проекта](#статус-проекта)

## О проекте

Система автоматизирует основные процессы склада: ведение номенклатуры и остатков, регистрацию прихода и расхода товаров, работу с поставщиками и клиентами, обработку заказов и формирование отчетности.

Приложение состоит из SPA-клиента, REST API, реляционной базы данных и Redis. Nginx предоставляет единую точку входа и проксирует запросы между frontend и backend. Доступ к функциям системы защищен короткоживущими JWT access-токенами и разграничен по ролям; отзывные refresh-сессии хранятся в Redis.

## Возможности

- **Аутентификация и роли.** Поддерживаются роли администратора (`admin`) и сотрудника склада (`worker`). Управление пользователями доступно только администратору.
- **Товары и остатки.** Создание и редактирование карточек товаров, учет текущего количества, оформление прихода и расхода, просмотр истории движений.
- **Контрагенты.** Ведение справочника клиентов и поставщиков с контактной информацией.
- **Заказы.** Оформление входящих и исходящих заказов, подбор товаров и контрагента, указание пункта назначения для отгрузок.
- **Жизненный цикл заказов.** Переходы между статусами `pending`, `processing`, `completed` и `canceled`; при завершении заказа складские остатки обновляются автоматически.
- **Дашборд.** Сводные показатели по товарам, заказам и контрагентам, график динамики входящих и исходящих заказов.
- **Отчетность.** Формирование оборотно-сальдовой ведомости в PDF за выбранный период.
- **Восстановление доступа.** Восстановление имени пользователя и сброс пароля с одноразовым кодом, хранящимся в Redis; отправка уведомлений через SMTP.
- **Интерактивная документация API.** OpenAPI-спецификация и Swagger UI доступны вместе с backend-сервисом.

## Технологический стек

| Уровень | Технологии |
| --- | --- |
| Frontend | React 19, Vite 7, React Router, React Bootstrap, Axios, Recharts, Nginx |
| Backend | Go 1.24.4, Gin, JWT, `database/sql` |
| Хранение данных | PostgreSQL 17, SQL-миграции |
| Кэш и временные данные | Redis 8.4 |
| Документация API | OpenAPI 3.0, Swagger UI |
| Отчеты | go-pdf/fpdf |
| Тестирование | Go testing, Testify, Pytest, Selenium |
| Инфраструктура | Nginx reverse proxy, Docker, Docker Compose |

## Архитектура

```mermaid
flowchart LR
    U[Пользователь] -->|http://localhost| N[Nginx gateway]
    N --> F[React SPA / Nginx]
    N -->|REST, JSON, JWT| B[Go / Gin API]
    B -->|Данные склада| P[(PostgreSQL)]
    B -->|Refresh-сессии, OTP, лимиты запросов| R[(Redis)]
    B -->|Письма| S[SMTP-сервер]
    B -->|Отчеты| D[PDF]
```

Backend организован по слоям `handler → service → repository`: HTTP-обработчики отвечают за транспорт и валидацию запросов, сервисы содержат бизнес-логику, а репозитории работают с PostgreSQL и Redis.

## Структура репозитория

```text
.
├── backend/
│   ├── api/openapi.yml       # OpenAPI-спецификация
│   ├── cmd/                  # точка входа backend-приложения
│   ├── internal/             # handlers, services, repositories, models
│   ├── migrations/           # миграции PostgreSQL
│   ├── Dockerfile
│   └── Makefile
├── frontend/
│   └── warehouse-management-system/
│       ├── src/              # React-приложение
│       ├── Dockerfile
│       ├── nginx.conf        # раздача production-сборки SPA
│       └── package.json
├── nginx/                    # конфигурация reverse proxy
├── ui-tests/                 # Selenium UI-тесты
├── docker-compose.yml        # запуск полного приложения
└── README.md
```

## Быстрый запуск

### Требования

- [Git](https://git-scm.com/)
- [Docker](https://docs.docker.com/get-docker/) с поддержкой Docker Compose

### 1. Клонирование репозитория

```bash
git clone https://github.com/callmerussell04/warehouse-management-coursework.git
cd warehouse-management-coursework
```

### 2. Настройка переменных окружения

Compose-файл ожидает локальный файл backend-конфигурации, который не хранится в Git. Скопируйте безопасный шаблон:

```bash
cp backend/.env.example backend/.env
```

Затем задайте собственные значения обязательных переменных в `backend/.env`:

```dotenv
JWT_SECRET=replace-with-at-least-32-random-bytes
ADMIN_USERNAME=admin
ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=replace-with-a-strong-password
ALLOWED_ORIGINS=http://localhost,http://localhost:5000
COOKIE_SECURE=false

# Необязательно: без SMTP письма только имитируются в логах backend
SMTP_HOST=
SMTP_PORT=587
SMTP_USERNAME=
SMTP_PASSWORD=
SMTP_FROM=
```

В Docker frontend использует относительные адреса API. Создавать для него `.env` не требуется. Подключение к PostgreSQL и Redis, внутренний порт backend и адрес доверенного reverse proxy уже заданы в `docker-compose.yml`.

### 3. Сборка и запуск

```bash
docker compose up --build -d
docker compose ps
```

После успешного запуска доступны:

| Компонент | Адрес |
| --- | --- |
| Web-интерфейс | [http://localhost](http://localhost) |
| REST API | [http://localhost/api](http://localhost/api) |
| Проверка состояния | [http://localhost/health](http://localhost/health) |
| Swagger UI | [http://localhost/docs/index.html](http://localhost/docs/index.html) |
| OpenAPI YAML | [http://localhost/openapi.yml](http://localhost/openapi.yml) |

Frontend и backend не публикуют собственные HTTP-порты на хост: все браузерные запросы проходят через reverse proxy на порту `80`.

> [!WARNING]
> При первом запуске backend создает администратора из `ADMIN_*`. Пароль должен содержать 12–72 байта. Значения из примера и `COOKIE_SECURE=false` предназначены только для локальной разработки; в production используйте случайные секреты, HTTPS и `COOKIE_SECURE=true`.

### 4. Остановка

```bash
docker compose stop
```

Чтобы удалить созданные контейнеры и сеть проекта, используйте:

```bash
docker compose down
```

Данные PostgreSQL и Redis сохраняются в именованных Docker-томах. Для их удаления вместе с контейнерами используйте `docker compose down -v`.

## Конфигурация

### Backend

| Переменная | По умолчанию | Назначение |
| --- | --- | --- |
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/warehouse_db?sslmode=disable` | Строка подключения к PostgreSQL |
| `REDIS_URL` | `localhost:6379` | Адрес Redis |
| `HTTP_PORT` | `8080` | Порт HTTP-сервера |
| `JWT_SECRET` | Обязательная | Секрет подписи access-токенов, не менее 32 байт |
| `ADMIN_USERNAME` | Обязательная | Логин создаваемого при первом запуске администратора |
| `ADMIN_EMAIL` | Обязательная | Email администратора |
| `ADMIN_PASSWORD` | Обязательная | Пароль администратора, 12–72 байта |
| `ALLOWED_ORIGINS` | Обязательная | Точный список разрешенных CORS-origin через запятую; wildcard не поддерживается |
| `COOKIE_SECURE` | Обязательная | Передавать refresh-cookie только по HTTPS (`true` для production) |
| `TRUSTED_PROXIES` | — | IP-адреса или CIDR доверенных reverse proxy через запятую; в Compose задается автоматически |
| `SMTP_HOST` | — | Хост SMTP-сервера |
| `SMTP_PORT` | — | Порт SMTP-сервера, обычно `587` |
| `SMTP_USERNAME` | — | Имя пользователя SMTP |
| `SMTP_PASSWORD` | — | Пароль SMTP |
| `SMTP_FROM` | — | Адрес отправителя |

Если `SMTP_HOST` или `SMTP_USERNAME` не заданы, сервис не отправляет реальные письма и записывает факт отправки в лог. Для полноценного восстановления доступа настройте SMTP.

### Frontend

| Переменная | Пример | Назначение |
| --- | --- | --- |
| `VITE_API_BASE_URL` | Пустое значение | Необязательный базовый URL REST API; нужен только при локальном запуске frontend вне Docker |

## Локальная разработка

### Backend

Для запуска без контейнера backend требуется Go 1.24.4, а PostgreSQL и Redis должны быть доступны локально. Инфраструктуру и миграции можно поднять через Compose:

```bash
docker compose up -d db redis migrate
```

Добавьте в `backend/.env` параметры локального подключения:

```dotenv
DATABASE_URL=postgres://postgres:postgres@localhost:5432/warehouse_db?sslmode=disable
REDIS_URL=localhost:6379
HTTP_PORT=8080
JWT_SECRET=replace-with-at-least-32-random-bytes
ADMIN_USERNAME=admin
ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=replace-with-a-strong-password
ALLOWED_ORIGINS=http://localhost,http://localhost:5000
COOKIE_SECURE=false
TRUSTED_PROXIES=
```

Go-приложение не загружает `.env` автоматически, поэтому экспортируйте переменные перед запуском:

```bash
cd backend
set -a
source .env
set +a
make run
```

### Frontend

Для разработки frontend требуется Node.js 20+ и npm:

```bash
cd frontend/warehouse-management-system
npm ci
printf 'VITE_API_BASE_URL=http://localhost:8080\n' > .env
npm run dev
```

Dev-сервер будет доступен по адресу [http://localhost:5000](http://localhost:5000).

## API

API использует 15-минутный JWT access-токен в заголовке `Authorization: Bearer <token>` и случайный refresh-токен в HTTP-only cookie. При обновлении токена refresh-сессия ротируется: старый cookie становится недействительным. Выход и смена пароля отзывают соответствующие серверные сессии. Публичные маршруты восстановления доступа и входа ограничены по частоте и при превышении лимита возвращают `429 Too Many Requests`.

Полное описание моделей, параметров, ответов и кодов ошибок находится в [OpenAPI-спецификации](backend/api/openapi.yml). При запущенном приложении ее можно изучать и выполнять запросы через [Swagger UI](http://localhost/docs/index.html).

Основные группы ресурсов:

- `/auth` — вход, выход, обновление токена и восстановление доступа;
- `/api/users` — управление пользователями, доступно администратору;
- `/api/products` — товары, остатки и история операций;
- `/api/counterparties` — клиенты и поставщики;
- `/api/orders` — входящие и исходящие заказы;
- `/api/reports` — формирование отчетов.

## Тестирование

### Backend

Команда поднимает отдельные тестовые PostgreSQL и Redis, генерирует моки и запускает все Go-тесты:

```bash
cd backend
make test
```

Требуются Go и Docker с поддержкой команды `docker compose`, используемой в `backend/Makefile`. При первом запуске также будут загружены `mockery` и `gotestsum`.

### Frontend

```bash
cd frontend/warehouse-management-system
npm ci
npm run lint
npm run build
```

### UI-тесты

UI-тесты написаны на Pytest и Selenium. Для их запуска требуются Python 3, Google Chrome и работающее приложение на `http://localhost`:

```bash
python3 -m venv .venv
source .venv/bin/activate
python -m pip install pytest selenium webdriver-manager
pytest ui-tests
```

> [!IMPORTANT]
> Тесты изменяют данные приложения и рассчитаны на тестовое окружение. Помимо администратора, некоторые сценарии ожидают заранее созданного и активированного пользователя `worker` с учетными данными, указанными в тестах. Не запускайте UI-тесты на базе с ценными данными.

## Статус проекта

Проект разработан в учебных целях как курсовая работа. Он демонстрирует построение полнофункционального веб-приложения с разделением на frontend и backend, REST API, ролевой моделью доступа, миграциями, интеграционными компонентами и автоматизированными тестами.
