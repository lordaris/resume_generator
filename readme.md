# Resume Generator

An app created to generate resumes.

# Resume Generator Application

A full-stack application with Go backend and React frontend, using Docker for containerization.

## Prerequisites

- Docker 20.10+
- Docker Compose 2.0+
- Go 1.20+ (for local backend development)
- Node.js 18+ and Yarn/npm (for local frontend development)

## Quick Start

1. **Clone the repository**

   ```bash
   git clone git@github.com:lordaris/resume_generator.git
   cd resume_generator
   ```

2. **Setup environment**

   ```bash
   cp .env.example .env
   # Edit .env with your values (it works with the default values too)
   ```

   ```bash
   cd backend && cp .env.example .env
   # Yes, I have a .env file inside backend with almost the same values as the one in the root folder, it works for local backend development.
   ```

3. **Start everything and run migrations**

   ```bash
   make dev
   ```

   This command:

   - Starts all containers (Postgres, Redis, backend, frontend)
   - Runs database migrations
   - Verifies container status, database connectivity, and application health
   - Makes the application ready to use

4. **Access applications**
   - Backend: <http://localhost:8080> (or your configured BACKEND_PORT)
   - Frontend: <http://localhost:3000> (or your configured FRONTEND_PORT)

## Development Workflows

For development workflows please check the makefile as it is self documented.
You can also run `make help` (in unix systems). If you're using windows without a unix shell, you still can read the comments in the makefile.
