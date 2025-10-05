# BE-K16

A social media newsfeed backend service built with Go, following Clean Architecture principles and microservices patterns.

## 1. Tech Stack

This project leverages modern Go technologies and cloud-native tools, organized by their purpose:

### Core Framework & Language
- **[Go](https://go.dev/)** 1.21+ - Primary programming language

### Handler Layer (API & Communication)
- **[Gin](https://gin-gonic.com/en/docs/quickstart/)** - HTTP web framework for RESTful APIs
- **[gRPC](https://grpc.io/docs/what-is-grpc/)** - High-performance RPC framework for inter-service communication ([Go Quick Start](https://grpc.io/docs/languages/go/quickstart/))
- **[Protocol Buffers](https://protobuf.dev/)** - Language-neutral data serialization for gRPC services

### Service Layer (Business Logic)
- **[bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)** - Password hashing for secure authentication

### DAO Layer (Data Access)
- **[MySQL](https://www.mysql.com/)** 8.0 - Primary relational database
- **[GORM](https://gorm.io/docs/)** - ORM library for MySQL database operations
- **[golang-migrate](https://github.com/golang-migrate/migrate)** - Database migration tool for version control
- **[Redis](https://redis.io/)** 7.2 - In-memory cache for performance optimization
- **[go-redis](https://github.com/redis/go-redis)** - Redis client for caching layer ([Installation Guide](https://github.com/redis/go-redis?tab=readme-ov-file#installation))
- **[Kafka](https://kafka.apache.org/)** - Distributed event streaming platform for asynchronous processing
- **[Kafka Go Client](https://github.com/segmentio/kafka-go)** - Kafka client for event streaming

### Observability & Monitoring
- **[Prometheus](https://prometheus.io/)** - Metrics collection and time-series database
- **[Grafana](https://grafana.com/)** - Metrics visualization and dashboards
- **[Zap](https://github.com/uber-go/zap)** - Structured, high-performance logging library

### Configuration & Environment
- **[godotenv](https://github.com/joho/godotenv)** - Load environment variables from `.env` files
- **[env](https://github.com/caarlos0/env)** - Parse environment variables into structs

### Development & Testing
- **[Testify](https://github.com/stretchr/testify)** - Testing toolkit with assertions and mocks
- **[Mockery](https://github.com/vektra/mockery)** - Mock code autogeneration for testing
- **[sqlmock](https://github.com/DATA-DOG/go-sqlmock)** - Mock SQL driver for database testing

### DevOps & Deployment
- **[Docker](https://www.docker.com/)** - Containerization platform ([Docker Curriculum](https://docker-curriculum.com/))
- **[Docker Compose](https://docs.docker.com/compose/)** - Multi-container orchestration for local development

## 2. Project Structure

This project follows **Clean Architecture** principles, separating concerns into distinct layers with clear dependencies. The architecture ensures maintainability, testability, and scalability.

### 2.1. Directory Structure

```
BE-K14/
├── cmd/                          # Application entry points
│   ├── http/                     # HTTP API server
│   ├── grpc/                     # Unified gRPC service (handles both user and post operations)
│   └── newsfeed_worker/          # Background worker for newsfeed processing
│
├── config/                       # Configuration management
│   ├── http_config.go            # HTTP server configuration
│   ├── grpc_config.go            # gRPC service configuration
│   ├── newsfeed_worker_config.go # Newsfeed worker configuration
│   └── log_config.go             # Logging configuration
│
├── internal/                     # Private application code
│   ├── handler/                  # Handler Layer (Presentation)
│   │   ├── http/                 # HTTP handlers (REST API endpoints)
│   │   ├── grpc/                 # Unified gRPC handlers (user and post operations)
│   │   ├── newsfeed_processor/   # Kafka consumer handlers
│   │   └── proto/                # Protocol Buffer definitions
│   │
│   ├── service/                  # Service Layer (Business Logic)
│   │   ├── user_service/         # User domain business logic
│   │   ├── post_service/         # Post domain business logic
│   │   └── model/                # Domain models (shared across services)
│   │
│   ├── dao/                      # DAO Layer (Data Access)
│   │   ├── user_dao/             # User database operations
│   │   ├── user_cache/           # User Redis cache operations
│   │   ├── post_dao/             # Post database operations
│   │   ├── post_cache/           # Post Redis cache operations
│   │   └── kafka_producer/       # Kafka message producer
│   │
│   └── common/                   # Shared utilities
│       └── error.go              # Common error definitions
│
├── pkg/                          # Public reusable packages
│   ├── logger/                   # Logging utilities (Zap wrapper)
│   ├── monitor/                  # Monitoring utilities (Prometheus)
│   └── time_util/                # Time manipulation utilities
│
├── script/                       # Utility scripts
│   └── db_migration/             # Database migration files
│
├── monitoring/                   # Observability configuration
│   ├── prometheus/               # Prometheus configuration
│   └── grafana/                  # Grafana dashboards and datasources
│
├── data/                         # Local development data (Docker volumes)
│   ├── db/                       # MySQL data files
│   └── redis/                    # Redis data files
│
└── docker-compose.yml            # Docker services orchestration
```

### 2.2. Clean Architecture Overview

Clean Architecture organizes code into layers with the **Dependency Rule**: source code dependencies must point **inward only**, toward higher-level policies. Inner layers know nothing about outer layers.

**Layers in this project:**
1. **Service Layer** (Core/Innermost) - Contains core business rules and use cases
2. **Handler Layer** (Outer) - Handles external communication (HTTP, gRPC)
3. **DAO Layer** (Outer) - Manages data persistence and external data sources

#### Dependency Flow

The key principle is that **both Handler and DAO layers depend on the Service layer** (the core), not the other way around. The Service layer defines interfaces that the outer layers must implement.

```
┌─────────────────────────────────────────────────────────┐
│  Handler Layer (Outer)                                  │
│  - Depends on: Service Layer (implements & calls it)    │
│  - Knows about: User request/Response formats, routing  │
│  - Doesn't know: Database details                       │
└────────────────────┬────────────────────────────────────┘
                     │
                     │ depends on & calls
                     ▼
┌─────────────────────────────────────────────────────────┐
│  Service Layer (CORE - Innermost)                       │
│  - Depends on: NOTHING (pure business logic)            │
│  - Defines interfaces: DAI, CacheDAI                    │
│  - Knows about: Business rules, domain models           │
│  - Doesn't know: HTTP/gRPC, databases, kafka, ...       │
└────────────────────┬────────────────────────────────────┘
                     ▲
                     │ implements interfaces
                     │ defined by Service
┌─────────────────────────────────────────────────────────┐
│  DAO Layer (Outer)                                      │
│  - Depends on: Service Layer (implements its interfaces)│
│  - Implements: DAI, CacheDAI interfaces                 │
│  - Knows about: SQL queries, data mapping               │
│  - Doesn't know: Business logic, HTTP/grpc handlers     │
└─────────────────────────────────────────────────────────┘
```

**Key Points**:
- **Service Layer is the core** - it defines the business logic and interfaces
- **Handler Layer depends on Service** - calls service methods to execute business logic
- **DAO Layer depends on Service** - implements data access interfaces defined by Service
- Service layer has **no dependencies** on outer layers (Dependency Inversion Principle).
- This protects business logic from changes in UI, database, or external systems

#### Data Transfer Between Layers

Normally, each communication has a DTO (data transfer object), for example: handler-service, service-dao, etc. In this project, to simplify those DTO types, we will reuse business objects to transfer data as well.. Data flows through the layers using **domain models** defined in `internal/service/model/`:

**1. Request Flow (Client → Database)**

```
HTTP/gRPC Request
       ↓
┌───────────────────────────────────────────────────┐
│  Handler Layer                                    │
│                                                   │
│  1. Parse request (JSON/Protobuf)                 |
│  2. Convert to domain model (model.User)          |
│  3. Call service method                           |
│     service.Signup(ctx, user)                     |                 
└────────┬──────────────────────────────────────────┘
         │ domain model (model.User)
         ▼
┌───────────────────────────────────────────────────┐
│  Service Layer                                    │
│                                                   │
│  1. Validate business rules                       |
│  2. Apply business logic (hash password)          |
│  3. Call DAO method                               |
│     dai.Create(ctx, user)                         |
└────────┬──────────────────────────────────────────┘
         │ domain model (model.User)
         ▼
┌───────────────────────────────────────────────────┐
│   DAO Layer                                       │
│                                                   │
│  1. Convert domain model to DB model              |
│  2. Execute SQL query                             |
│  3. Return domain model                           |
└───────────────────────────────────────────────────┘
         ↓
    Database
```

**2. Response Flow (Database → Client)**

```
    Database
         ↓
┌───────────────────────────────────────────────────┐
│   DAO Layer                                       │
│                                                   │
│  1. Fetch data from DB                            |
│  2. Convert DB model to domain model              |
│  3. Return model.User                             |
└────────┬──────────────────────────────────────────┘
         │ domain model (model.User)
         ▼
┌───────────────────────────────────────────────────┐
│  Service Layer                                    │
│                                                   │
│  1. Apply business logic (if needed)
│  2. Return model.User
└────────┬──────────────────────────────────────────┘
         │ domain model (model.User)
         ▼
┌───────────────────────────────────────────────────┐
│  Handler Layer                                    │
│                                                   │
│  1. Convert domain model to response format
│  2. Return JSON/Protobuf response
└───────────────────────────────────────────────────┘
         ↓
HTTP/gRPC Response
```

#### Why Use Interfaces?

The Service Layer (core) defines interfaces for what it needs, and outer layers implement them. This is **Dependency Inversion Principle** in action:

**Service Layer defines DAO interfaces (what it needs from data layer):**
```go
// in internal/service/user_service/user_service.go
type UserDAI interface {
    Create(ctx context.Context, user *model.User) (*model.User, error)
    GetByUsername(ctx context.Context, username string) (*model.User, error)
    // ... other methods
}

type UserService struct {
    dai UserDAI  // Service depends on interface it defines
}
```

**DAO Layer implements the interface:**
```go
// in internal/dao/user_dao/user_db.go
type UserDAI struct {
    db *gorm.DB
}

// UserDAI implements the UserDAI interface defined by Service
func (d *UserDAI) Create(ctx context.Context, user *model.User) (*model.User, error) {
    // implementation details
}
```

**Handler Layer also defines Service interfaces (for testing):**
```go
// in internal/handler/grpc/handler.go
type UserService interface {
    Signup(ctx context.Context, user *model.User) (*model.User, error)
    Login(ctx context.Context, user *model.User) (*model.User, error)
    Follow(ctx context.Context, userId, peerId int64) (*model.Follow, error)
    // ... other methods
}

type PostService interface {
    CreatePost(ctx context.Context, post *model.Post) (*model.Post, error)
    GetPostByUserID(ctx context.Context, userId int, paging model.Paging) ([]*model.Post, error)
    GetNewsfeed(ctx context.Context, userId int, paging model.Paging) ([]*model.Post, error)
}

type userGrpcHandler struct {
    userService UserService  // Handler depends on interfaces (for easy testing)
    postService PostService
}

func (h *userGrpcHandler) Signup(ctx context.Context, req *grpc_pb.SignupRequest) (*grpc_pb.SignupResponse, error) {
    // Convert request to domain model
    user := &model.User{
        Username: req.GetUserName(),
        Password: req.GetPassword(),
        // ...
    }
    
    // Call service through interface
    result, err := h.userService.Signup(ctx, user)
    
    // Convert domain model to response
    return &grpc_pb.SignupResponse{...}, nil
}
```

**Note**: While Handler could depend on concrete Service implementation (since Service is the core), using an interface here makes Handler testing easier by allowing mock services.

**Benefits**:
- **Service Layer is independent**: Business logic doesn't depend on databases, UI, or frameworks
- **Testing at all layers**: Both Handler and Service can be tested independently with mocks
- **Flexibility**: Can change handler/database implementations (HTTP → gRPC, MySQL → PostgreSQL, ...) without affecting Service. Business logic is protected from infrastructure changes

#### Domain Models

All layers communicate using shared domain models from `internal/service/model/`:

- `model.User` - User entity
- `model.Post` - Post entity
- `model.Follow` - Follow relationship
- `model.Paging` - Pagination parameters

### 2.3. Clean Architecture Implementation

#### 1. Handler Layer (`internal/handler/`)
**Responsibility:** Handle external requests and responses

- **HTTP Handlers** (`http/`): REST API endpoints using Gin framework
  - Parse HTTP requests
  - Call service layer methods
  - Format HTTP responses
  - Handle authentication/authorization middleware
  
- **gRPC Handlers** (`grpc/`): Unified RPC service implementation
  - Implements Protocol Buffer service definitions for both user and post operations
  - Converts between protobuf messages and domain models
  - Applies interceptors for logging, monitoring, and authentication

- **Event Handlers** (`newsfeed_processor/`): Kafka message consumers
  - Process asynchronous events
  - Trigger business logic based on events

**Dependencies:** Handler → Service (calls service layer, knows nothing about DAO)

#### 2. Service Layer (`internal/service/`)
**Responsibility:** Implement business logic and use cases

- Contains domain-specific business rules
- Orchestrates data flow between handlers and DAOs
- Validates business constraints
- Implements transaction logic
- Independent of delivery mechanisms (HTTP, gRPC, etc.)

**Example:** `user_service/user_service.go`
- `Signup()`: Validates username uniqueness, hashes passwords, creates users
- `Login()`: Authenticates users, verifies credentials
- `Follow()`: Manages user relationships with business rules

**Dependencies:** Service → DAO (calls DAO interfaces, doesn't know about databases/cache implementation)

#### 3. DAO Layer (`internal/dao/`)
**Responsibility:** Abstract data access and persistence

- **Database DAOs** (`user_dao/`, `post_dao/`): GORM-based database operations
  - CRUD operations
  - Query building
  - Transaction management
  
- **Cache DAOs** (`user_cache/`, `post_cache/`): Redis-based caching
  - Cache-aside pattern implementation
  - TTL management
  
- **Message Queue** (`kafka_producer/`): Event publishing
  - Asynchronous message production

**Dependencies:** DAO → External Systems (databases, cache, message queues)

### Additional Components

Beyond the three main layers, the project includes supporting components for configuration, logging, and monitoring:

#### Configuration (`config/`)

Manages environment-specific configuration for all services.

**Purpose**: Centralized configuration management using environment variables

**Key files**:
- `env_type.go` - Defines environment types (local, production)
- `http_config.go` - HTTP server configuration (host, port, JWT key, gRPC client address)
- `grpc_config.go` - gRPC service configuration (database, Redis, Kafka settings)
- `newsfeed_worker_config.go` - Worker configuration (Kafka, Redis settings)
- `log_config.go` - Logging configuration

**How it works**:
- Uses `godotenv` to load `.env` files in local development
- Uses `env` package to parse environment variables into structs
- Each service loads its own configuration at startup
- Validates required fields before starting services

**Example**: Each service calls its config loader (e.g., `config.LoadHttpConfig()`, `config.LoadGrpcConfig()`) which reads from environment variables or `.env` file.

#### Logging (`pkg/logger/`)

Provides structured logging across all services using Zap.

**Purpose**: Structured logging with multiple log levels

**Key files**:
- `log.go` - Logger interface and initialization
- `zap.go` - Zap logger implementation

**Features**:
- **Structured logging**: Log fields as key-value pairs instead of formatted strings
- **Log levels**: Debug, Info, Warn, Error with environment-based filtering


**Usage patterns**:
```go
logger.Info("grpc created", logger.F("user_id", userId), logger.F("username", username))
logger.Error("database error", logger.E(err))
logger.Debug("cache hit", logger.F("key", cacheKey))
```

#### Monitoring (`pkg/monitor/`)

Exposes Prometheus metrics for observability.

**Purpose**: Track application performance and health metrics

**Key files**:
- `api.go` - Defines and exports Prometheus metrics

**Sample metrics exposed**:
- **Request Counter** (`newsfeed_api_status_count`): Total requests by endpoint, method, and status code
- **Request Latency** (`newsfeed_api_status_latency`): Response time percentiles (p50, p90, p99)

**How it works**:
- Metrics are defined using Prometheus client library
- HTTP middleware automatically records metrics for every request
- Metrics are exposed at `/metrics` endpoint in Prometheus format
- Prometheus scrapes this endpoint every 5 seconds

**Integration**:
- Used by HTTP middleware to track all API requests
- Can be extended to add custom business metrics (e.g., signup rate, post creation rate)
- Works seamlessly with Prometheus and Grafana for visualization

#### Utilities (`pkg/time_util/`)

Common utility functions used across services.

**Purpose**: Reusable helper functions

**Example**: Time manipulation utilities for consistent timestamp handling

### 2.4. Benefits of This Architecture

1. **Testability**: Each layer can be tested independently with mocks
2. **Maintainability**: Clear separation of concerns makes code easier to understand
3. **Flexibility**: Easy to swap implementations (e.g., change from MySQL to PostgreSQL)
4. **Scalability**: Services can be deployed independently as microservices
5. **Reusability**: Business logic is decoupled from delivery mechanisms
6. **Observability**: Built-in logging and monitoring for production debugging

## 3. How to Run Project

This section provides step-by-step instructions to set up and run the project locally.

### Prerequisites

Ensure you have the following installed:
- **Go** 1.21+ ([Download](https://go.dev/dl/))
- **Docker** & **Docker Compose** ([Installation Guide](https://docs.docker.com/get-docker/))
- **golang-migrate** CLI tool for database migrations
  ```bash
    go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
    ```
- **Protocol Buffers compiler** (if modifying `.proto` files)
  ```bash
  # macOS
  brew install protobuf
  
  # Linux
  apt install -y protobuf-compiler
  ```

### Step 1: Clone the Repository

```bash
git clone <repository-url>
cd BE-K14
```

### Step 2: Install Go Dependencies

```bash
go mod download
```

### Step 3: Start Infrastructure Services

Start all required infrastructure services (MySQL, Redis, Kafka, Zookeeper, Prometheus, Grafana) using Docker Compose:

```bash
make up
# or
docker-compose up -d
```

This will start:
- **MySQL** on port `3306`
- **Redis** on port `6379`
- **Kafka** on port `9092`
- **Zookeeper** on port `2181`
- **Prometheus** on port `9090`
- **Grafana** on port `3000` (admin/admin)

**Verify services are running:**
```bash
docker ps
```

### Step 4: Initialize Kafka Topics

Create the required Kafka topic for post events:

```bash
make kafka_init_topic
```

This creates a `posts` topic with 3 partitions. You only need to run this once.

**Verify topic creation:**
```bash
make kafka_topic_status
```

### Step 5: Run Database Migrations

Apply database schema migrations to MySQL:

```bash
make migrate_up_db
```

This creates the necessary tables (`users`, `user_users`, etc.).

**Check migration status:**
```bash
make migrate_status
```

**Access MySQL directly (optional):**
```bash
make exec_db
```

### Step 6: Configure Environment Variables

Create a `.env` file in the project root with the following configuration:

```bash
# Environment
ENV=local

# HTTP Server
HTTP_HOST=localhost
HTTP_PORT=8080

# gRPC Service
GRPC_HOST=localhost
GRPC_PORT=50051

# Database (MySQL)
DATABASE_USER=grpc
DATABASE_PASSWORD=pass1234
DATABASE_HOST=localhost
DATABASE_PORT=3306
DATABASE_NAME=newsfeed

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_ENABLED=true

# Kafka
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC=posts
KAFKA_CONSUMER_GROUP=newsfeed_worker

# JWT Secret
JWT_KEY=your-secret-jwt-key
```

### Step 7: Start Application Services

The project consists of multiple microservices that depend on each other. Start them in the following order:

#### 7.1. Start gRPC Service (Port 50051)

This unified service handles all gRPC requests for both user and post operations.

```bash
go run cmd/grpc/main.go
```

**Dependencies:** MySQL, Redis, Kafka

#### 7.2. Start Newsfeed Worker

This background worker consumes post events from Kafka and processes newsfeeds.

```bash
go run cmd/newsfeed_worker/main.go
```

**Dependencies:** Redis, Kafka

#### 7.3. Start HTTP API Server (Port 8080)

This is the main API gateway that clients interact with.

```bash
go run cmd/http/main.go
```

**Dependencies:** gRPC Service

### Step 8: Verify Services are Running

Check that all services are healthy:

```bash
# Check HTTP API health
curl http://localhost:8080/metrics

# Check Prometheus targets
open http://localhost:9090/targets

# Check Grafana dashboards
open http://localhost:3000
```

### Service Dependency Graph

```
┌─────────────────┐
│   HTTP Server   │  (Port 8080)
│   (API Gateway) │
└────────┬────────┘
         │
         │ calls gRPC
         │
         ▼
┌────────────────────────────┐
│    gRPC Service            │  (Port 50051)
│ (User + Post Operations)   │
└────────┬───────────────────┘ 
         │        
         ├──────────────────├─────────|
         ▼                  ▼         ▼
    ┌────────┐         ┌────────┐ ┌────────┐
    │  MySQL │         │ Redis  │ │ Kafka  │
    └────────┘         └────────┘ └───┬────┘
                                      │
                                      ▼
                              ┌────────────────┐
                              │ Newsfeed Worker│
                              └────────────────┘
```

### Common Shell Commands

```bash
# Start all Docker services
make up

# Stop all Docker services
make down

# Run database migration (up)
make migrate_up_db

# Rollback database migration (down)
make migrate_down_db

# Check migration version
make migrate_status

# Initialize Kafka topic
make kafka_init_topic

# Check Kafka topic status
make kafka_topic_status

# Check Kafka consumer group status
make kafka_consumer_status

# Access MySQL CLI
make exec_db

# Access Redis CLI
make exec_redis

# Generate Protocol Buffer files (after modifying .proto)
make gen_proto
```

### Troubleshooting

**Issue: Database connection refused**
- Ensure MySQL container is running: `docker ps | grep mysql`
- Wait a few seconds for MySQL to fully initialize
- Check credentials in `.env` match `docker-compose.yml`

**Issue: gRPC connection failed**
- Ensure the gRPC service is started before the HTTP server
- Check that port 50051 is not in use

**Issue: Kafka topic not found**
- Run `make kafka_init_topic` to create the topic
- Verify with `make kafka_topic_status`

**Issue: Port already in use**
- Check if another process is using the port: `lsof -i :8080`
- Kill the process or change the port in `.env`

### Development Workflow

1. Make code changes
2. If you modified `.proto` files, regenerate code: `make gen_proto`
3. If you modified database schema, create a new migration file
4. Restart the affected service
5. Test your changes

### Testing

Run unit tests:
```bash
go test ./...
```

Run tests with coverage:
```bash
go test -cover ./...
```

## 4. Unit Testing

This project follows a comprehensive testing strategy that leverages Clean Architecture to ensure each layer can be tested independently. The separation of concerns makes it easy to write fast, reliable unit tests without external dependencies.

### How Clean Architecture Enables Testability

Clean Architecture makes testing easier through three key principles:

1. **Dependency Inversion**: Each layer depends on interfaces, not concrete implementations
   - Service layer depends on `UserDAI` interface, not the actual database implementation
   - Handler layer depends on `UserService` interface, not the concrete service
   - This allows us to easily swap real implementations with mocks during testing
   
2. **Isolation**: Each layer can be tested independently by mocking its dependencies
   - Test handlers without starting gRPC servers or HTTP servers
   - Test services without connecting to databases or Redis
   - Test DAOs with controlled database mocks
   - Each test focuses on one layer's responsibility

3. **Single Responsibility**: Each component has one reason to change
   - Handlers only handle request/response transformation
   - Services only contain business logic
   - DAOs only handle data access
   - Clear boundaries make it obvious what to test in each layer

### Testing the Three Layers

#### 1. Testing the DAO Layer

**Purpose**: Verify database operations and SQL queries work correctly

**Strategy**: Use `sqlmock` to mock database connections without a real database

**What to test**:
- SQL query generation (INSERT, SELECT, UPDATE, DELETE)
- Data mapping between database models and domain models
- Transaction handling
- Error handling for database failures

**Tools**: `sqlmock`, `testify/assert`, GORM

**Example test cases**:
- `TestUserDAI_Create` - Verify user creation SQL and returned ID

#### 2. Testing the Service Layer

**Purpose**: Verify business logic is correct and complete

**Strategy**: Mock the DAO layer using `mockery`-generated mocks

**What to test**:
- Business rules and validations (e.g., username uniqueness, password strength)
- Data transformations (e.g., password hashing)
- Orchestration of multiple DAO calls
- Error handling and custom error codes
- Cache logic (cache hit/miss scenarios)

**Tools**: `testify/mock`, `mockery`, `testify/assert`

**Example test cases**:
- `TestUserService_Signup` - Test username validation, password hashing, user creation
- `TestUserService_Login` - Test authentication logic
- `TestUserService_Follow` - Test relationship creation with validation

**Key pattern**: Create mock DAOs, set expectations with `.On()`, inject mocks into service, verify behavior

#### 3. Testing the Handler Layer

**3.1. Testing gRPC Handlers**

**Purpose**: Verify gRPC request/response transformation

**Strategy**: Mock the service layer

**What to test**:
- Protocol Buffer to domain model conversion
- Domain model to Protocol Buffer conversion
- Error code mapping from service errors to gRPC status codes
- Request validation

**Tools**: `testify/mock`, `mockery`, Protocol Buffers

**Example test cases**:
- `TestGrpcHandler_Signup` - Test protobuf transformation for user signup
- `TestGrpcHandler_Follow` - Test request parsing and response building
- `TestGrpcHandler_CreatePost` - Test post creation protobuf handling

**3.2. Testing HTTP Handlers**

**Purpose**: Verify HTTP request/response handling

**Strategy**: Use `httptest` to simulate HTTP requests, mock gRPC clients

**What to test**:
- JSON request parsing
- HTTP routing
- Response formatting
- Status code mapping
- Middleware (authentication, logging)

**Tools**: `httptest`, `testify/mock`, `gin`

**Example test cases**:
- `TestServer_Signup` - Test JSON parsing and HTTP response
- `TestServer_Login` - Test authentication flow
- `TestServer_Follow` - Test JWT middleware integration

### Using Interfaces and Mocks

#### Why Interfaces Enable Testing

In this project, each layer defines interfaces for its dependencies:

```
Service Layer defines:
- UserDAI interface (implemented by user_dao)
- UserCacheDAI interface (implemented by user_cache)
- PostDAI interface (implemented by post_dao)
- PostCacheDAI interface (implemented by post_cache)

Handler Layer defines:
- UserService interface (implemented by user_service)
- PostService interface (implemented by post_service)
```

This allows tests to inject mock implementations instead of real ones.

#### Generating Mocks with Mockery

1. **Install Mockery**:
   ```bash
  go install github.com/vektra/mockery/v3@latest
  ```

2. **Configure** `.mockery.yaml` to specify which interfaces to mock

3. **Generate mocks**:
   ```bash
   mockery
   ```

4. **Use mocks in tests**:
   - Create mock: `mockDAI := new(MockUserDAI)`
   - Set expectations: `mockDAI.On("GetByUsername", ctx, "username").Return(user, nil)`
   - Inject into component: `service := &UserService{dai: mockDAI}`
   - Verify calls: `mockDAI.AssertExpectations(t)`

#### Mock Configuration

The `.mockery.yaml` file specifies:
- Output directory: same as source file
- Filename pattern: `<source>_mock.go`
- Struct name pattern: `Mock<InterfaceName>`
- Which interfaces to generate mocks for

### Testing Best Practices

1. **Test Naming Convention**
   - File: `<filename>_test.go`
   - Function: `Test<StructName>_<MethodName>`
   - Sub-tests: Use `t.Run("scenario description", ...)`

2. **Test Structure (AAA Pattern)**
   - **Arrange**: Set up test data and mocks
   - **Act**: Execute the function under test
   - **Assert**: Verify the results

3. **Test Coverage Goals**
   - DAO Layer: SQL correctness and data mapping
   - Service Layer: All business logic branches and edge cases
   - Handler Layer: Request/response transformation and error handling

4. **What to Mock vs. What Not to Mock**
   - **Mock**: External dependencies (databases, caches, APIs, gRPC clients)
   - **Don't mock**: Domain models, utility functions, simple data structures

5. **Test Independence**
   - Each test should run independently
   - Use `t.Run()` for organizing related test cases
   - Clean up resources in `defer` statements

### Benefits of This Testing Approach

1. **Fast Tests**: No external dependencies means tests run in milliseconds
2. **Reliable**: Tests don't fail due to network issues or database state
3. **Isolated**: Each layer tested independently with clear boundaries
4. **Easy to Debug**: When a test fails, you know exactly which layer has the problem
5. **Comprehensive**: Can test error cases that are hard to reproduce with real systems
6. **Refactoring Safety**: Tests ensure behavior remains correct when refactoring

### Test Dependency Flow

```
┌─────────────────────────────────────────────────────────┐
│ Handler Tests                                           │
│ - Mock: Service Layer                                   │
│ - Test: Request/Response transformation                 │
│ - Tools: httptest, mockery                              │
└────────────────────┬────────────────────────────────────┘
                     │ depends on
                     ▼
┌─────────────────────────────────────────────────────────┐
│ Service Tests                                           │
│ - Mock: DAO Layer (database, cache)                     │
│ - Test: Business logic, validation                      │
│ - Tools: testify/mock, mockery                          │
└────────────────────┬────────────────────────────────────┘
                     │ depends on
                     ▼
┌─────────────────────────────────────────────────────────┐
│ DAO Tests                                               │
│ - Mock: Database connections                            │
│ - Test: SQL queries, data mapping                       │
│ - Tools: sqlmock, GORM                                  │
└─────────────────────────────────────────────────────────┘
```

Each layer only mocks the layer below it, creating a clear testing hierarchy that mirrors the Clean Architecture dependency flow.

## 5. (Additional) Monitoring with Prometheus and Grafana

This project includes built-in monitoring using Prometheus for metrics collection and Grafana for visualization. The monitoring stack is automatically set up with Docker Compose.

### Architecture Overview

```
┌─────────────┐
│ Application │ (Port 8080)
│  Services   │ Exposes /metrics endpoint
└──────┬──────┘
       │ scrapes metrics every 5s
       ▼
┌─────────────┐
│ Prometheus  │ (Port 9090)
│             │ Stores time-series data
└──────┬──────┘
       │ queries data
       ▼
┌─────────────┐
│  Grafana    │ (Port 3000)
│             │ Visualizes metrics
└─────────────┘
```

### What Metrics Are Collected

The application automatically exports the following metrics:

#### 1. API Request Counter
- **Metric**: `newsfeed_api_status_count`
- **Type**: Counter
- **Labels**: 
  - `api`: API endpoint path (e.g., `/user/signup`)
  - `method`: HTTP method (GET, POST, etc.)
  - `status`: HTTP status code (200, 404, 500, etc.)
- **Purpose**: Track total number of requests per endpoint and status code

#### 2. API Request Latency
- **Metric**: `newsfeed_api_status_latency`
- **Type**: Summary (p50, p90, p99 percentiles)
- **Labels**: Same as counter
- **Purpose**: Track response time distribution for performance monitoring

### How to Access Monitoring

#### Step 1: Ensure Services Are Running

Make sure Docker Compose services are up:
```bash
make up
# or
docker-compose up -d
```

This starts:
- Prometheus on `http://localhost:9090`
- Grafana on `http://localhost:3000`

#### Step 2: Start Application Services

Start at least one application service (e.g., HTTP server) to generate metrics:
```bash
go run cmd/http/main.go
```

The HTTP server exposes metrics at `http://localhost:8080/metrics`

#### Step 3: Access Prometheus

Open Prometheus UI: **http://localhost:9090**

**What you can do**:
- **View targets**: Go to Status → Targets to see if Prometheus is successfully scraping your application
- **Query metrics**: Use the query interface to explore metrics
  - Example query: `newsfeed_api_status_count`
  - Example query: `rate(newsfeed_api_status_count[1m])` (requests per second)
  - Example query: `newsfeed_api_status_latency{quantile="0.99"}` (p99 latency)
- **View graphs**: Visualize metric trends over time

**Common Prometheus Queries**:
```promql
# Total requests per endpoint
sum by (api) (newsfeed_api_status_count)

# Request rate (requests per second)
rate(newsfeed_api_status_count[1m])

# Error rate (4xx and 5xx responses)
sum by (api) (rate(newsfeed_api_status_count{status=~"4..|5.."}[1m]))

# P99 latency per endpoint
newsfeed_api_status_latency{quantile="0.99"}

# Average latency
rate(newsfeed_api_status_latency_sum[1m]) / rate(newsfeed_api_status_latency_count[1m])
```

#### Step 4: Access Grafana

Open Grafana UI: **http://localhost:3000**

**Default credentials**:
- Username: `admin`
- Password: `admin`

**What you can do**:
1. **Prometheus is pre-configured**: The Prometheus datasource is automatically set up
2. **Create dashboards**: Build custom dashboards to visualize metrics
3. **Set up alerts**: Configure alerts for critical metrics (e.g., high error rate, slow response time)

**Creating Your First Dashboard**:
1. Click "+" → "Dashboard" → "Add new panel"
2. Select "Prometheus" as the data source
3. Enter a query (e.g., `rate(newsfeed_api_status_count[1m])`)
4. Choose visualization type (Graph, Gauge, Table, etc.)
5. Save the dashboard

**Recommended Panels**:
- **Request Rate**: `rate(newsfeed_api_status_count[1m])` - Line graph
- **Error Rate**: `sum(rate(newsfeed_api_status_count{status=~"4..|5.."}[1m]))` - Gauge
- **P99 Latency**: `newsfeed_api_status_latency{quantile="0.99"}` - Line graph
- **Status Code Distribution**: `sum by (status) (newsfeed_api_status_count)` - Pie chart
- **Top Endpoints by Traffic**: `topk(5, sum by (api) (rate(newsfeed_api_status_count[1m])))` - Bar chart

### Configuration Files

#### Prometheus Configuration
Location: `monitoring/prometheus/prometheus.yml`

```yaml
global:
  scrape_interval: 5s  # Scrape metrics every 5 seconds

scrape_configs:
  - job_name: "newsfeed_apps"
    static_configs:
      - targets:
        - host.docker.internal:8080  # Application metrics endpoint
```

**Note**: `host.docker.internal` allows Prometheus (running in Docker) to access services on the host machine.

#### Grafana Datasource Configuration
Location: `monitoring/grafana/provisioning/datasources/datasource.yml`

Prometheus is automatically configured as the default datasource when Grafana starts.

### How Metrics Are Implemented

The application uses Prometheus client library with Gin middleware:

1. **Metrics Definition** (`pkg/monitor/api.go`):
   - Defines Counter and Summary metrics
   - Registers metrics with Prometheus

2. **Middleware** (`internal/handler/http/middleware.go`):
   - Wraps every HTTP request
   - Records request count and latency
   - Exports metrics with labels (endpoint, method, status code)

3. **Metrics Endpoint** (`internal/handler/http/server.go`):
   - Exposes `/metrics` endpoint using `promhttp.Handler()`
   - Prometheus scrapes this endpoint

### Monitoring Best Practices

1. **Use Labels Wisely**: Labels create separate time series - avoid high cardinality (e.g., don't use user IDs as labels)

2. **Monitor Key Metrics**:
   - **Request Rate**: Understand traffic patterns
   - **Error Rate**: Detect issues quickly
   - **Latency**: Identify performance problems
   - **Saturation**: Track resource usage (CPU, memory, connections)

3. **Set Up Alerts**: Configure Grafana alerts for:
   - High error rate (> 5%)
   - Slow response time (p99 > 1s)
   - Service unavailability

4. **Create Dashboards**: Build dashboards for different audiences:
   - **Operations**: System health, error rates, latency
   - **Business**: Request volume, user activity
   - **Development**: Endpoint performance, cache hit rates

### Troubleshooting

**Prometheus can't scrape application metrics**:
- Verify application is running and `/metrics` endpoint is accessible: `curl http://localhost:8080/metrics`
- Check Prometheus targets page: http://localhost:9090/targets
- Ensure `host.docker.internal` resolves correctly (may need to use `host.docker.internal` on Mac/Windows, or `172.17.0.1` on Linux)

**Grafana shows "No data"**:
- Verify Prometheus datasource is configured: Configuration → Data Sources
- Check if Prometheus has data: Query in Prometheus UI first
- Verify time range in Grafana matches when metrics were generated

**No metrics appearing**:
- Make sure you've made requests to the application to generate metrics
- Check that the monitoring middleware is enabled in the HTTP server
