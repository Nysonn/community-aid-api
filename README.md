# CommunityAid API

## 1. Project Overview

CommunityAid is a backend REST API for a community aid platform that connects people
in need with volunteers and donors. Registered users can post emergency requests for
medical assistance, food, rescue, or shelter. Community members can respond with offers
of help or record donations against a request. A team of administrators can review and
approve or reject requests, manage users, and monitor activity through a dashboard.

The API is built for use by a frontend single-page application. All protected routes
require a valid JWT issued by Clerk.

---

## 2. Tech Stack

| Component       | Technology                          |
|-----------------|-------------------------------------|
| Language        | Go 1.25                             |
| HTTP framework  | Gin                                 |
| Database        | PostgreSQL via Neon (serverless)    |
| Migrations      | Goose                               |
| Authentication  | Clerk (JWT verification)            |
| Media uploads   | Cloudinary                          |
| Transactional email | Resend                          |
| Containerisation | Docker, Docker Compose             |
| Hot reload      | Air                                 |

---

## 3. Architecture Overview

The project follows a layered architecture:

- **Handlers** (`internal/handlers/`) receive HTTP requests, validate input, call the
  appropriate service method, and write the HTTP response. No SQL lives here.
- **Services** (`internal/services/`) contain all database interactions and business
  rules. Each service owns one resource domain (users, requests, offers, donations).
- **Middleware** (`internal/middleware/`) handles cross-cutting concerns: CORS, Clerk JWT
  verification with database user lookup, admin role enforcement, and per-IP rate limiting.
- **Models** (`internal/models/`) define Go structs for database rows and request/response
  bodies with JSON and validation tags.
- **Helpers** (`internal/helpers/`) provide shared utilities for input validation, response
  formatting, and pagination parameter parsing.
- **Config** (`internal/config/`) loads all environment variables at startup, failing fast
  if a required variable is absent.
- **Packages** (`pkg/`) wrap third-party SDK initialisations for Cloudinary, Resend, and
  Clerk so the rest of the codebase does not depend directly on SDK internals.
- **Migrations** (`internal/migrations/`) contain sequential Goose SQL migration files that
  define and seed the database schema.

The entry point is `cmd/server/main.go`, which wires all dependencies together and starts
the HTTP server with graceful shutdown.

---

## 4. Prerequisites

- Go 1.25 or higher (for local development without Docker)
- Docker and Docker Compose
- A Neon PostgreSQL database (or any PostgreSQL 14+ instance)
- A Clerk account with an application configured
- A Cloudinary account
- A Resend account with a verified sender domain

---

## 5. Environment Setup

Copy the example file and fill in all values:

```
cp .env.example .env
```

| Variable                | Description                                                      |
|-------------------------|------------------------------------------------------------------|
| `APP_ENV`               | Runtime environment label, e.g. `development` or `production`   |
| `PORT`                  | Port the HTTP server listens on, e.g. `3000`                     |
| `DB_URL`                | PostgreSQL connection string including credentials and SSL params |
| `CLOUDINARY_CLOUD_NAME` | Cloudinary cloud name from the dashboard                         |
| `CLOUDINARY_API_KEY`    | Cloudinary API key                                               |
| `CLOUDINARY_API_SECRET` | Cloudinary API secret                                            |
| `RESEND_API_KEY`        | Resend API key for sending transactional email                   |
| `CLERK_SECRET_KEY`      | Clerk secret key used to verify JWTs server-side                 |
| `CLERK_PUBLISHABLE_KEY` | Clerk publishable key (used for reference; verified server-side) |
| `ALLOWED_ORIGINS`       | Comma-separated list of allowed CORS origins. Leave empty to allow all (development default). |

---

## 6. Running the Project

Start the API with live reload inside Docker:

```
make dev
```

This runs `docker-compose up --build` targeting the `dev` stage of the Dockerfile.
Air watches for file changes and recompiles automatically. The API is available at
`http://localhost:3000`.

Stop all containers:

```
make down
```

---

## 7. Running Migrations

Apply all pending migrations:

```
make migrate-up
```

Roll back the most recent migration:

```
make migrate-down
```

Check the current migration status of the database:

```
make migrate-status
```

Migration files live in `internal/migrations/` and are managed by Goose. Files must
follow Goose naming conventions, for example `00001_create_users_table.sql`.

Migration `005_seed_admin.sql` inserts a placeholder admin user using
`ON CONFLICT (email) DO NOTHING`, so it is safe to run multiple times.

---

## 8. API Overview

All response bodies follow a consistent envelope:

Success: `{ "success": true, "data": ... }`
Paginated: `{ "success": true, "data": [...], "meta": { "total", "page", "page_size", "total_pages" } }`
Error: `{ "success": false, "error": "..." }`

### Auth

| Method | Path                       | Auth     | Description                                        |
|--------|----------------------------|----------|----------------------------------------------------|
| POST   | /api/v1/auth/register      | None     | Register or retrieve a user by Clerk ID after sign-up |

### Users

| Method | Path                              | Auth          | Description                            |
|--------|-----------------------------------|---------------|----------------------------------------|
| GET    | /api/v1/users/me                  | ClerkAuth     | Get the authenticated user's profile   |
| PUT    | /api/v1/users/me                  | ClerkAuth     | Update the authenticated user's profile |
| POST   | /api/v1/users/me/avatar           | ClerkAuth     | Upload a new avatar image              |
| GET    | /api/v1/admin/users               | Admin         | List all users (paginated)             |
| PUT    | /api/v1/admin/users/:id/activate  | Admin         | Activate a user account                |
| PUT    | /api/v1/admin/users/:id/deactivate | Admin        | Deactivate a user account              |

### Emergency Requests

| Method | Path                              | Auth          | Description                                     |
|--------|-----------------------------------|---------------|-------------------------------------------------|
| GET    | /api/v1/requests                  | None          | List approved requests with optional filters (paginated) |
| GET    | /api/v1/requests/:id              | None          | Get a single request by ID                      |
| POST   | /api/v1/requests                  | ClerkAuth     | Create a new request (multipart, supports media uploads) |
| GET    | /api/v1/requests/me               | ClerkAuth     | List the authenticated user's own requests      |
| PUT    | /api/v1/requests/:id              | ClerkAuth     | Update a request (owner or admin)               |
| DELETE | /api/v1/requests/:id              | Admin         | Delete a request                                |
| POST   | /api/v1/requests/:id/approve      | Admin         | Approve a request and notify the owner by email |
| POST   | /api/v1/requests/:id/reject       | Admin         | Reject a request and notify the owner by email  |
| GET    | /api/v1/admin/requests            | Admin         | List all requests regardless of status (paginated) |

### Offers

| Method | Path                              | Auth          | Description                                     |
|--------|-----------------------------------|---------------|-------------------------------------------------|
| POST   | /api/v1/offers                    | None          | Submit an offer on an approved request          |
| GET    | /api/v1/offers/request/:request_id | None         | List all offers for a request                   |
| PUT    | /api/v1/offers/:id/status         | ClerkAuth     | Update an offer's status (request owner or admin) |
| GET    | /api/v1/admin/offers              | Admin         | List all offers across all requests (paginated) |

### Donations

| Method | Path                              | Auth          | Description                                     |
|--------|-----------------------------------|---------------|-------------------------------------------------|
| POST   | /api/v1/admin/donations           | Admin         | Record a donation against a request             |
| GET    | /api/v1/admin/donations           | Admin         | List all donations with request titles (paginated) |
| GET    | /api/v1/admin/donations/:request_id | Admin       | List all donations for a specific request       |

### Admin Dashboard

| Method | Path                              | Auth          | Description                                     |
|--------|-----------------------------------|---------------|-------------------------------------------------|
| GET    | /api/v1/admin/stats               | Admin         | Aggregate counts and totals across all resources |

### Health

| Method | Path                  | Auth  | Description                  |
|--------|-----------------------|-------|------------------------------|
| GET    | /api/v1/health        | None  | Returns `{ "status": "ok" }` |

---

## 9. Folder Structure

```
community-aid-api/
├── cmd/
│   └── server/         Entry point: wires dependencies and starts the server
├── internal/
│   ├── config/         Loads and validates all environment variables at startup
│   ├── db/             Opens and verifies the PostgreSQL connection
│   ├── handlers/       HTTP handlers: parse input, call services, write responses
│   ├── helpers/        Shared utilities: validation, response formatting, pagination
│   ├── middleware/     Gin middleware: CORS, Clerk auth, admin guard, rate limiting
│   ├── migrations/     Goose SQL migration files (run in order)
│   ├── models/         Go structs for DB rows, input types, and response types
│   ├── routes/         Registers all routes and applies middleware groups
│   └── services/       Database access and business logic, one file per resource
└── pkg/
    ├── clerk/          Initialises the Clerk SDK with the secret key
    ├── cloudinary/     Initialises the Cloudinary client
    └── resend/         Initialises the Resend email client
```

---

## 10. Notes for Maintainers

**Tests.** There are currently no automated tests. All service methods are written
against `context.Context`-aware database methods (`QueryRowContext`, `QueryContext`,
`ExecContext`) which makes them straightforward to test with a real database or a
connection to a test schema in future.

**Hot reload.** In development, Air watches all `.go` and `.toml` files and recompiles
the binary into `./tmp/main` on change. The Air configuration is in `.air.toml`.

**Docker stages.** The Dockerfile has four stages:
1. `migrate` — installs Goose for running migrations outside the application process.
2. `dev` — installs Air and mounts the source tree as a volume for hot reload.
3. `builder` — compiles a stripped production binary with `CGO_ENABLED=0`.
4. `production` — copies the binary into a minimal Alpine image and runs it as a
   non-root user. Gin is set to release mode via `GIN_MODE=release`.

**Rate limiting.** The in-memory rate limiter allows 60 requests per minute per IP.
Idle IP entries are cleaned up after 5 minutes. For multi-instance deployments this
should be replaced with a Redis-backed solution.

**CORS.** In development, leave `ALLOWED_ORIGINS` empty to allow all origins. In
production, set it to a comma-separated list of your frontend origins, for example
`https://app.communityaid.com`.
