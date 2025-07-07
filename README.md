# TripLink Backend

## Getting Started

### Prerequisites
- Docker and Docker Compose installed on your system
- Go 1.20 or later installed

### Running the Application

1. **Start the database**
   ```bash
   docker-compose up -d
   or
   docker compose up -d
   ```

2. **Check database status**
   ```bash
   docker-compose ps
   or docker compose ps
   ```

3. **Run the Go application**
   ```bash
   go run main.go
   ```
   
   Or build and run:
   ```bash
   go build -o triplink
   ./triplink
   ```

4. **Stop the database**
   ```bash
   docker-compose down
   or docker compose down
   ```

### Database Configuration

- **Host**: localhost
- **Port**: 5432
- **Database**: triplink
- **Username**: postgres
- **Password**: password

### Connection String
```
postgresql://postgres:password@localhost:5432/triplink
```

### Accessing the Database

You can connect to the PostgreSQL database using any PostgreSQL client or command line:

```bash
docker exec -it triplink_postgres psql -U postgres -d triplink
```
