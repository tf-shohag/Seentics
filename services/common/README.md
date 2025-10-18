# Common Service

A Go-based microservice that combines the functionality of both the Users and Workflows services into a single, unified service. This service provides user management, authentication, website management, workflow automation, and analytics capabilities.

## Features

### User Management
- User registration and authentication
- Profile management
- Password reset functionality
- Email verification
- **OAuth integration (Google & GitHub)**
- Social login support

### Website Management
- Website creation and management
- Tracking ID generation
- Privacy settings configuration
- Domain management

### Workflow Automation
- Workflow creation and execution
- **Real workflow execution engine with:**
  - Webhook actions with HMAC signatures
  - Email actions with variable substitution
  - Event tracking actions
  - Conditional logic and delays
  - Variable substitution ({{user.email}}, {{visitorId}}, etc.)
- Trigger and action configuration
- Workflow analytics and performance tracking
- Execution history with detailed results

### Cloud Features & Open Source Support
- **CLOUD_FEATURES_ENABLED flag** for SaaS vs Open Source deployment
- Subscription limits and plan enforcement (cloud only)
- Billing integration with Stripe (cloud only)
- Feature access control based on subscription plans
- Graceful degradation for open source deployment

### Additional Features
- Privacy compliance (GDPR)
- Support ticket system with message threading
- Event tracking and visitor analytics
- Admin panel functionality
- Internal API for microservice communication
- OAuth health check endpoints

## Architecture

The service follows a clean Go project structure:

```
services/common/
├── main.go                 # Entry point
├── config/                 # Configuration management
├── services/               # HTTP request handlers (renamed from controllers)
├── middleware/             # HTTP middleware
├── models/                 # Data structures
├── database/               # Database connections
├── routes/                 # Route definitions
├── utils/                  # Utility functions
└── Dockerfile             # Container definition
```

## Getting Started

### Prerequisites
- Go 1.21 or higher
- MongoDB
- Redis
- Docker (optional)

### Installation

1. Clone the repository and navigate to the common service:
```bash
cd services/common
```

2. Copy the environment file:
```bash
cp .env.example .env
```

3. Update the `.env` file with your configuration values.

4. Install dependencies:
```bash
go mod tidy
```

5. Run the service:
```bash
go run main.go
```

### Using Docker

1. Build the Docker image:
```bash
docker build -t seentics-common .
```

2. Run the container:
```bash
docker run -p 3004:3004 --env-file .env seentics-common
```

## API Endpoints

### Authentication
- `POST /api/v1/user/auth/register` - User registration
- `POST /api/v1/user/auth/login` - User login
- `POST /api/v1/user/auth/forgot-password` - Password reset request
- `POST /api/v1/user/auth/reset-password` - Password reset
- `GET /api/v1/user/auth/verify-email/:token` - Email verification
- `POST /api/v1/user/auth/google` - Google OAuth login
- `POST /api/v1/user/auth/github` - GitHub OAuth login
- `GET /api/v1/user/auth/oauth/health` - OAuth configuration health check

### User Management
- `GET /api/v1/user/users/profile` - Get user profile
- `PUT /api/v1/user/users/profile` - Update user profile
- `POST /api/v1/user/users/change-password` - Change password
- `DELETE /api/v1/user/users/account` - Delete account

### Website Management
- `GET /api/v1/user/websites` - Get user websites
- `POST /api/v1/user/websites` - Create website
- `GET /api/v1/user/websites/:id` - Get website details
- `PUT /api/v1/user/websites/:id` - Update website
- `DELETE /api/v1/user/websites/:id` - Delete website
- `POST /api/v1/user/websites/:id/regenerate-tracking` - Regenerate tracking ID

### Workflow Management
- `GET /api/v1/workflows` - Get workflows
- `POST /api/v1/workflows` - Create workflow
- `GET /api/v1/workflows/:id` - Get workflow details
- `PUT /api/v1/workflows/:id` - Update workflow
- `DELETE /api/v1/workflows/:id` - Delete workflow
- `POST /api/v1/workflows/:id/execute` - Execute workflow
- `GET /api/v1/workflows/:id/executions` - Get workflow executions

### Health Check
- `GET /health` - Service health status

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `3004` |
| `NODE_ENV` | Environment | `development` |
| `MONGO_URI` | MongoDB connection string | `mongodb://localhost:27017` |
| `REDIS_URL` | Redis connection string | `redis://localhost:6379` |
| `API_KEY` | API key for service authentication | - |
| `JWT_SECRET` | JWT signing secret | - |
| `GOOGLE_CLIENT_ID` | Google OAuth client ID | - |
| `GOOGLE_CLIENT_SECRET` | Google OAuth client secret | - |
| `GITHUB_CLIENT_ID` | GitHub OAuth client ID | - |
| `GITHUB_CLIENT_SECRET` | GitHub OAuth client secret | - |
| `FRONTEND_URL` | Frontend application URL | `http://localhost:3000` |
| `CLOUD_FEATURES_ENABLED` | Enable cloud features (billing, limits) | `false` |
| `WEBHOOK_SECRET` | Secret for HMAC webhook signatures | - |
| `EMAIL_API_KEY` | Email service API key | - |
| `EMAIL_FROM` | Default email sender | `noreply@seentics.com` |

## Development

### Project Structure

- **services/**: Contains HTTP request handlers (business logic)
- **middleware/**: HTTP middleware for authentication, API keys, etc.
- **models/**: Data structures and database schemas
- **database/**: Database connection and initialization
- **routes/**: Route definitions and grouping
- **config/**: Configuration management
- **utils/**: Utility functions and helpers

### Adding New Features

1. Define models in `models/`
2. Create service handlers in `services/`
3. Add routes in `routes/routes.go`
4. Add middleware if needed in `middleware/`

### Testing

```bash
go test ./...
```

## Deployment

The service can be deployed using Docker or directly as a Go binary. Make sure to:

1. Set proper environment variables
2. Ensure MongoDB and Redis are accessible
3. Configure proper API keys and secrets
4. Set up proper logging and monitoring

## Migration from Node.js Services

This Go service replaces the separate Node.js users and workflows services. Key benefits:

- **Performance**: Go's superior performance and lower memory usage
- **Simplicity**: Single service instead of multiple microservices
- **Type Safety**: Strong typing reduces runtime errors
- **Concurrency**: Better handling of concurrent requests
- **Deployment**: Simpler deployment with single binary

## Contributing

1. Follow Go best practices and conventions
2. Add tests for new functionality
3. Update documentation
4. Ensure proper error handling
5. Use structured logging

## License

MIT License
