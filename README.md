# Dekamond Task - OTP Authentication API

A simple Go REST API for OTP-based user authentication with JWT tokens, built with Gin framework and MongoDB.

## API Endpoints

| Method | Endpoint              | Description                  |
| ------ | --------------------- | ---------------------------- |
| POST   | `/send-otp`           | Send OTP to phone number     |
| POST   | `/verify-otp`         | Verify OTP and get JWT token |
| GET    | `/users`              | Get all users (paginated)    |
| GET    | `/users/{id}`         | Get user by ID               |
| GET    | `/users/search`       | Search users by phone number |
| GET    | `/swagger/index.html` | Swagger documentation        |

## Prerequisites

### For Local Development

- Go 1.24 or later
- MongoDB 4.4 or later
- Git

### For Docker Development

- Docker
- Docker Compose

## Running Locally

1. **Clone the repository**

   ```bash
   git clone <repository-url>
   cd dekamond-task
   ```

2. **Install dependencies**

   ```bash
   go mod download
   ```

3. **Start MongoDB**

   ```bash
   # Using Docker
   docker run -d --name mongodb -p 27017:27017 mongo

   # Or install MongoDB locally and start the service
   ```

4. **Set environment variables**

   ```bash
   # Create a .env file from .env.example and fill with your configuration
   cp .env.example .env

   ```

5. **Generate Swagger docs** (if using swaggo)

   ```bash
   # Install swag
   go install github.com/swaggo/swag/cmd/swag@latest

   # Generate docs
   swag init
   ```

6. **Run the application**

   ```bash
   go run main.go
   ```

7. **Access the application**
   - API: http://localhost:8080
   - Swagger UI: http://localhost:8080/swagger/index.html

## Running with Docker

1. **Clone the repository**

   ```bash
   git clone <repository-url>
   cd dekamond-task
   ```

2. **Start with Docker Compose**

   ```bash
   docker-compose up -d
   ```

3. **Access the services**
   - API: http://localhost:8080
   - Swagger UI: http://localhost:8080/swagger/index.html
   - MongoDB: localhost:27017

4. **View logs**

   ```bash
   docker-compose logs -f app
   ```

5. **Stop the services**
   ```bash
   docker-compose down
   ```

## Environment Variables

| Variable         | Description               |
| ---------------- | ------------------------- |
| `PORT`           | Server port               |
| `MONGO_URI`      | MongoDB connection string |
| `DB_NAME`        | Database name             |
| `JWT_SECRET_KEY` | JWT signing secret        |

## Usage Examples

### 1. Send OTP

```bash
curl -X POST http://localhost:8080/send-otp \
  -H "Content-Type: application/json" \
  -d '{"phone": "09126378234"}'
```

### 2. Verify OTP

```bash
# Check console for the OTP code, then verify
curl -X POST http://localhost:8080/verify-otp \
  -H "Content-Type: application/json" \
  -d '{"phone": "09126378234", "otp": "123456"}'
```

### 3. Get Users (with pagination)

```bash
curl "http://localhost:8080/users?page=1&page_size=10"
```

### 4. Search Users

```bash
curl "http://localhost:8080/users/search?phone=0912&page=1&page_size=5"
```

## Rate Limiting

The `/send-otp` endpoint is rate-limited to:

- **3 requests per 10 minutes** per phone number

## Database choice justification.

Since the schema for the "users" isn't finilized, and I didn't want to do every thing in the memory, starting with a NoSQL DB seemed a good option. By using a repository pattern and abstracting away how data is stroing in the "database" we can easily swap the MongoDB for a SQL database.

The choice of the app's database is more complicated than this. With this amount of information about the thing that I build, MongoDB is just fine. I need more information to make sure I'm choosing the right database for the project.
