# Resume Generator Application

A full-stack application with Go backend and React frontend, using Docker for containerization.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Quick Start (Docker)](#quick-start-docker)
- [Detailed Setup](#detailed-setup)
- [Configuration](#configuration)
- [Development Workflows](#development-workflows)
- [Production Deployment](#production-deployment)
- [Troubleshooting](#troubleshooting)
- [FAQ](#faq)

## Prerequisites

- Docker 20.10+
- Docker Compose 2.0+
- Go 1.20+ (for backend development)
- Node.js 18+ and Yarn (for frontend development)
- GNU Make (optional but recommended)

## Quick Start (Docker)

1. **Clone the repository**

   ```bash
   git clone https://github.com/yourusername/resume-generator.git
   cd resume-generator
   ```

2. **Setup environment**

   ```bash
   cp .env.example .env
   # Edit .env file with your values
   nano .env
   ```

3. **Start all services**

   ```bash
   make docker-up
   ```

4. **Run migrations**

   ```bash
   make migrate
   ```

5. **Access applications**
   - Backend: <http://localhost:8080>
   - Frontend: <http://localhost:3000>

## Detailed Setup

### Backend Setup

```bash
# Run with hot-reload for development
make run

# Run database migrations
make migrate

# Run tests
make test

# Build Docker image
make docker-build
```

### Frontend Setup

```bash
cd frontend

# Install dependencies
yarn install

# Start development server
yarn dev

# Build for production
yarn build
```

## Configuration

### Environment Variables (.env)

Note: I know the PORT is repeated, but I'm using both variables and I won't change it soon

```bash
# Backend
PORT=8080
DB_URL=postgres://postgres:postgres@postgres:5432/resume_generator?sslmode=disable
JWT_SECRET=your_secure_secret
REDIS_URL=redis://redis:6379/0

# Frontend
VITE_API_URL=http://localhost:8080/api/v1
FRONTEND_PORT=3000
BACKEND_PORT=8080
```

## Development Workflows

### Common Makefile Targets

```bash
# Start full stack
make docker-up

# Stop containers
make docker-down

# Run tests + linter
make verify

# Full cleanup
make clean

# View logs
make logs
```

### Hybrid Development

1. **Backend in Docker, Frontend locally**

   ```bash
   make docker-up  # Starts postgres, redis, backend
   cd frontend && yarn dev
   ```

2. **Frontend in Docker, Backend locally**

   ```bash
   make docker-up  # Starts postgres, redis, frontend
   make run       # Run backend locally
   ```

## Production Deployment

1. **Build production images**

   ```bash
   docker compose build --no-cache
   ```

2. **Start services**

   ```bash
   docker compose up -d --scale backend=3  # Example with 3 backend instances
   ```

3. **Configure reverse proxy**

   - Recommended: Nginx or Traefik
   - Set environment variables:

     ```ini
     VITE_API_URL=https://yourdomain.com/api/v1
     JWT_SECRET=your_production_secret
     ```

## Troubleshooting

### Common Issues

1. **Docker permission denied**

   ```bash
   sudo chown $USER /var/run/docker.sock
   ```

2. **Port conflicts**

   ```bash
   lsof -i :8080  # Find process using port
   ```

3. **Migration failures**

   ```bash
   make migrate-reset
   ```

4. **Missing dependencies**

   ```bash
   make frontend-install
   docker compose build --no-cache
   ```

## FAQ

**Q: How do I change the backend port?**
Nobody asked me this but it looks cool to have a FAQ section.

```bash
# .env
BACKEND_PORT=9090
make docker-down docker-up
```

**Q: How to reset the database?**

```bash
make migrate-reset
```

**Q: Where are the logs stored?**

```bash
make logs  # View all logs
make logs-app  # Backend logs
```

**Q: How to update dependencies?**

```bash
# Backend
go get -u ./...

# Frontend
cd frontend && yarn upgrade
```

**Q: How to access the database directly?**

```bash
docker compose exec postgres psql -U postgres
```
