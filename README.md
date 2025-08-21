# Financial Aggregator Service

A Go and gRPC-based service for aggregating and managing financial transactions from multiple sources, including CSV bank exports and APIs.
## Features

### Core Functionality
- **Multi-bank Transaction Import**: Support for CSV uploads from various banks (Revolut, American Express)
- **Monzo API Integration**: Direct integration with Monzo bank API for real-time transaction sync
- **Transaction Management**: CRUD operations for financial transactions with categorization
- **User & Bank Management**: Multi-user support with bank association management
- **Category System**: Flexible transaction categorization with income/outcome classification

### API Endpoints
- `GET /transactions` - Retrieve filtered transactions by month/year
- `PATCH /transactions/{id}` - Update transaction category and type
- `POST /upload-csv` - Upload bank CSV files for transaction parsing
- `GET /monzo/auth-url` - Get Monzo OAuth authentication URL
- `GET /monzo/callback` - Handle Monzo OAuth callback
- `GET /monzo/account` - Get Monzo account id
- `GET /monzo/transactions` - Load transactions from Monzo API
- `GET /banks` - List supported banks and their import methods
- `GET /users` - List system users
- `GET /categories` - List transaction categories
- `GET /transaction-types` - List transaction types

## Architecture

### Project Structure
```
├── api/                    # Protocol buffer definitions
├── cmd/                    # Application entry points
├── config/                 # Configuration files
├── internal/               # Internal application code
│   ├── app/               # Application initialization
│   ├── config/            # Configuration management
│   ├── server/            # gRPC/HTTP server implementation
│   ├── service/           # Business logic services
│   └── utils/             # Utility packages
├── migrations/            # Database migrations
├── pkg/                   # Generated protobuf code
└── web/                   # React frontend
```

### Backend Services
- **gRPC Server** with HTTP/REST gateway
- **PostgreSQL** database with connection pooling
- **Modular Service Architecture**:
    - Bank Service
    - Category Service
    - Monzo Integration Service
    - Transaction Service
    - Uploader Service
    - User Service

### Frontend
- **React SPA** with Vite build system
- **Chart Visualizations** using Recharts
- **Transaction Management Interface**
- **CSV Upload Interface**

## Technology Stack

### Backend
- **Go 1.24.2**
- **gRPC** with HTTP gateway
- **PostgreSQL** with pgx driver
- **Protocol Buffers** for API definitions
- **Viper** for configuration management
- **Zap** for structured logging

### Frontend
- **React 19**
- **Vite** for build tooling
- **Recharts** for data visualization
- **Lucide React** for icons

### Infrastructure
- **Docker** & Docker Compose
- **PostgreSQL 16**
- **Database migrations** with Goose

## Getting Started

### Prerequisites
- Go 1.24.2+
- Docker & Docker Compose
- Node.js (for frontend development)
- Protocol Buffers compiler (for API changes)

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/Everest13/fin-aggregator-service.git
   cd fin-aggregator-service
   ```

2. **Configure the service using environment variables or the config files in `/config`**

3. Generate Go code from Protocol Buffers
    ```bash
    make proto-gen
    ```
    Or use make proto-gen if defined in your Makefile.

4. **Start the services**
   ```bash
   docker-compose up -d
   ```
   This will start:
    - PostgreSQL database on port 5434
    - Backend service on port 8082

5. **Run frontend (development)**
   ```bash
   cd web
   npm install
   npm run dev
   ```

### Configuration

The service can be configured using **environment variables** (`.env`) or **YAML values** (`values.yaml`). Environment variables override YAML settings if both are present.

#### `.env` File
This file is used for quick local setup and Docker development. Example variables:

```env
# Monzo API
MONZO_CLIENT_ID=<your_monzo_client_id>
MONZO_CLIENT_SECRET=<your_monzo_client_secret>

# Database
DB_PASSWORD=<your_db_password>
DB_PORT=5432

# HTTP Server
HTTP_PORT=8080

# gRPC Server
GRPC_PORT=8081

# Frontend
FRONTEND_PORT=5173
```

#### `values.yaml` File
The service can be configured using environment variables (.env) or YAML values (values.yaml). Environment variables override YAML settings if both are present.
```yaml
# Monzo API
monzo:
  redirect_uri: "http://localhost:5173/monzo/callback"  # Local development
  # For production use: "https://myapp.com/monzo/callback"

# gRPC Server
grpc:
  network: "tcp"

# HTTP Server
http:
  host: "0.0.0.0"
  graceful_timeout: "15s"
  client_timeout: "30s"

# Database
database:
  name: "fin_aggregator_db"
  user: "fin_aggregator_user"
  host: "localhost"
  ssl_mode: "disable"
  max_cons: 20
  min_cons: 5
  max_con_lifetime: "1h"
```

### Monzo Integration

To integrate Monzo with the service, follow these steps:

1. **Register an OAuth client** on Monzo and obtain:
    - `MONZO_CLIENT_ID`
    - `MONZO_CLIENT_SECRET`
    - `REDIRECT_URI` — the URL where Monzo will redirect users after authentication
        - Local development example: `http://localhost:5173/monzo/callback`
        - Production example: `https://myapp.com/monzo/callback`  
          Follow the official Monzo documentation: [Acquire an Access Token](https://docs.monzo.com/#acquire-an-access-token)
2. **Get the authentication URL** via `/monzo/auth-url`.
3. **Complete the OAuth flow** in the browser.
4. **Load transactions** via `/monzo/transactions`.
A Monzo authorization window will appear; the user must approve access in the Monzo app and confirm within the window. Transactions are loaded only after approval.

## Usage

### Transactions Interface

To manage transactions, follow the general workflow:
1. Choose the target user and the bank associated with that user.
2. Load transactions according to the bank integration type:
    - **CSV Upload**: upload the bank CSV file and review any parsing errors.
    - **Monzo API**: initiate the transaction sync via the API. A Monzo authorization window will appear; the user must approve access in the Monzo app and confirm in the popup. Transactions are loaded only after approval.

### Transaction Management

- View and filter transactions by month and year.
- Update transaction categories and types.
- Analyze transactions in the **Analytics** tab, which visualizes spending by categories over the year.


## Database Schema

The service uses the following main entities:

- **Transactions**: Core financial transaction records, partitioned by `transaction_date` and linked to users, banks, and categories.
- **Users**: System users with associated banks.
- **Banks**: Supported financial institutions, with optional custom headers for CSV/API imports.
- **Categories**: Transaction categorization system, including category keywords for automated tagging.
- **User-Bank Associations**: Link users to the banks they have accounts in.
- **Bank Header Mappings**: Map bank-specific CSV/API headers to standard transaction fields.

Migrations are located in `/migrations` and handled automatically on startup.

## Development

### API Development
The service uses Protocol Buffers for API definitions. To regenerate code after changes:

### Generate Go code after .proto changes:
```bash
make proto-gen
```

### Running Tests
```bash
go test ./...
```

### Building for Production
```bash
# Backend
go build -o bin/fin-aggregator-service ./cmd

# Frontend
cd web
npm run build
```




