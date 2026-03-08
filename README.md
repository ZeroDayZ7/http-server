# http-server (portfolio backend)

High-performance HTTP API server written in **Go** using **Fiber**.

## Architecture

The system uses an **event-driven architecture** to process user interactions.

* **Redis Streams** – primary event ingestion layer for interactions (likes, dislikes, visits)
* **Workers** – asynchronously consume events and process statistics
* **MySQL** – durable storage and backup layer
* **Lua scripts** – atomic Redis operations for consistency
* **Hash-based rate limiting** – prevents users from sending excessive interactions

## Data Flow

1. Client sends interaction request
2. API writes event to **Redis Stream**
3. Worker consumes event and processes it
4. Aggregated statistics are stored in **MySQL**

## Features

* Fiber HTTP server
* Redis Streams event processing
* MySQL persistence
* Lua scripts for atomic Redis operations
* Hash-based rate limiting for interaction control
* Middleware: CORS, Helmet, Rate Limiting, Logging
* Graceful shutdown
* Docker support

## Project Structure

```
cmd/                application entrypoint (main.go)

config/             application configuration and middleware setup
                    (CORS, compression, security headers, rate limiter, database, Redis)

internal/
  db/               sqlc generated database layer, queries and models
  di/               dependency injection configuration (Google Wire)
  errors/           centralized application error handling
  handler/          HTTP request handlers
  middleware/       request validation middleware
  redis/            Redis Streams integration, Lua scripts and producers/consumers
  repository/       data access layer (MySQL implementation)
  router/           API route definitions
  server/           server lifecycle and graceful shutdown
  service/          business logic layer
  shared/           shared utilities (logger, uuid, common middleware)
  validator/        request validation utilities
  worker/           background workers processing Redis stream events

migrations/         database migration files
logs/               application log files
scripts/            helper and build scripts

Makefile            build and development commands
docker-compose.yml  local development environment
sqlc.yaml           sqlc configuration for generating database code

```