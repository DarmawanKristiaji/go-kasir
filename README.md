# Go Kasir - Point of Sale REST API

![Go Version](https://img.shields.io/badge/Go-1.23.0-blue)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Supabase-green)
![Deployment](https://img.shields.io/badge/Railway-Deployed-purple)

Modern Point of Sale (POS) REST API built with Go, featuring clean architecture, PostgreSQL database, and automatic migrations.

## 🚀 Live Production

**Production API:** [https://go-kasir-railway.dakr.my.id/](https://go-kasir-railway.dakr.my.id/)

Access the root endpoint to see complete API documentation and available endpoints.

## ✨ Features

- ✅ **Clean Architecture** - Layered structure (Handlers → Services → Repositories → Models)
- ✅ **Product Management** - Full CRUD operations with stock tracking
- ✅ **Category Management** - Organize products by categories
- ✅ **Relational Data** - Products linked to categories with LEFT JOIN queries
- ✅ **Auto Migrations** - Database schema automatically created/updated on startup
- ✅ **IPv4 Optimization** - Multi-fallback DNS resolution for Railway deployment
- ✅ **Transaction Pooler** - Optimized connection pooling with Supabase
- ✅ **Environment Config** - Secure configuration via environment variables

## 📋 Prerequisites

- Go 1.23.0 or higher
- PostgreSQL database (Supabase recommended)
- Git

## 🛠️ Installation

### 1. Clone Repository

```bash
git clone https://github.com/DarmawanKristiaji/go-kasir.git
cd go-kasir
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Configure Environment

Create `.env` file in the project root:

```env
PORT=8080
DB_CONN=host=your-db-host port=6543 user=your-user password=your-password dbname=postgres sslmode=require options=-c search_path=public
```

**Important:** Never commit `.env` file to Git. It's already in `.gitignore`.

### 4. Run Application

```bash
go run main.go
```

Or build and run:

```bash
go build -o app
./app
```

Server will start on `http://localhost:8080` (or the PORT specified in `.env`)

## 📡 API Endpoints

### Root
- `GET /` - API documentation and endpoint list

### Health Check
- `GET /health` - Service health status and database connectivity

### Categories

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/categories` | List all categories |
| POST | `/categories` | Create new category |
| GET | `/categories/{id}` | Get category by ID |
| PUT | `/categories/{id}` | Update category |
| DELETE | `/categories/{id}` | Delete category |

**Category JSON Structure:**
```json
{
  "id": 1,
  "name": "Minuman",
  "description": "Kategori Minuman"
}
```

### Products

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/produk` | List all products (with category names) |
| POST | `/api/produk` | Create new product |
| GET | `/api/produk/{id}` | Get product by ID (with category name) |
| PUT | `/api/produk/{id}` | Update product |
| DELETE | `/api/produk/{id}` | Delete product |

**Product JSON Structure:**
```json
{
  "id": 1,
  "name": "Sprite",
  "price": 5000,
  "stock": 100,
  "category_id": 1,
  "category_name": "Minuman"
}
```

## 💻 Example Usage

### Create Category

```bash
curl -X POST https://go-kasir-railway.dakr.my.id/categories \
  -H "Content-Type: application/json" \
  -d '{"name":"Minuman","description":"Kategori Minuman"}'
```

### Create Product with Category

```bash
curl -X POST https://go-kasir-railway.dakr.my.id/api/produk \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Sprite",
    "price": 5000,
    "stock": 100,
    "category_id": 1
  }'
```

### Get All Products (with Category Names)

```bash
curl https://go-kasir-railway.dakr.my.id/api/produk
```

Response includes `category_name` via LEFT JOIN:
```json
[
  {
    "id": 1,
    "name": "Sprite",
    "price": 5000,
    "stock": 100,
    "category_id": 1,
    "category_name": "Minuman"
  }
]
```

## 🏗️ Project Structure

```
.
├── main.go                 # Application entry point & DI setup
├── database/
│   └── database.go         # DB connection & migrations
├── models/
│   └── models.go           # Data structures
├── repositories/
│   ├── product_repository.go
│   └── category_repository.go
├── services/
│   ├── product_service.go
│   └── category_service.go
├── handlers/
│   ├── product_handler.go
│   └── category_handler.go
├── migrations/
│   └── init.sql            # Initial schema (auto-applied)
├── go.mod                  # Go dependencies
└── .env                    # Environment config (not in git)
```

## 🔧 Architecture

**Layered Architecture Pattern:**

```
HTTP Request
    ↓
Handlers (HTTP layer)
    ↓
Services (Business logic)
    ↓
Repositories (Data access)
    ↓
Database (PostgreSQL)
```

**Benefits:**
- Clear separation of concerns
- Easy to test and maintain
- Scalable codebase
- Reusable business logic

## 🚢 Deployment

### Railway Deployment

Application is configured for automatic deployment on Railway:

1. **Push to GitHub** - Automatic trigger on `main` branch
2. **Build** - Railway runs `go build`
3. **Start** - Application starts automatically
4. **Migrations** - Database schema auto-created on first connection

### Environment Variables on Railway

Set these in Railway dashboard:

- `PORT` - Automatically set by Railway
- `DB_CONN` - Your database connection string

**Connection String Format:**
```
host=your-pooler-host port=6543 user=postgres.xxx password=xxx dbname=postgres sslmode=require options=-c search_path=public
```

## 🗄️ Database Schema

### Categories Table
```sql
CREATE TABLE categories (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### Products Table
```sql
CREATE TABLE products (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    price INT NOT NULL,
    stock INT NOT NULL,
    category_id BIGINT REFERENCES categories(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

**Relationship:** Products have optional foreign key to Categories with `ON DELETE SET NULL`

## 🔐 Environment Configuration

### Required Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `DB_CONN` | PostgreSQL connection string | See format above |

### Database Connection

Application supports:
- Direct Supabase connection (port 5432)
- Transaction pooler (port 6543) - **Recommended for Railway**
- Session pooler

**IPv4 Resolution:** Automatic multi-fallback DNS resolution ensures Railway compatibility:
1. System resolver with IPv4 filter
2. Cloudflare UDP4 resolver (1.1.1.1:53)
3. DNS-over-HTTPS fallback

## 🧪 Testing

Run locally with test database:

```bash
export PORT=8080
export DB_CONN="your-test-db-connection"
go run main.go
```

Test endpoints:
```bash
# Health check
curl http://localhost:8080/health

# List categories
curl http://localhost:8080/categories

# List products
curl http://localhost:8080/api/produk
```

## 📦 Dependencies

- **github.com/lib/pq** - PostgreSQL driver
- **github.com/spf13/viper** - Configuration management

Install all dependencies:
```bash
go mod download
```

## 🐛 Troubleshooting

### Database Connection Issues

**Problem:** `network is unreachable` errors on Railway

**Solution:** Use transaction pooler (port 6543) with `options=-c search_path=public`

### Migration Errors

**Problem:** Table already exists

**Solution:** Migrations use `CREATE TABLE IF NOT EXISTS` and `ALTER TABLE IF NOT EXISTS` - safe to re-run

### Routes Return 404

**Problem:** Database not connected

**Solution:** Check `DB_CONN` environment variable is set correctly. Routes are only registered when database connection succeeds.

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to branch (`git push origin feature/AmazingFeature`)
5. Open Pull Request

## 📝 License

This project is open source and available under the [MIT License](LICENSE).

## 👨‍💻 Author

**Darmawan Kristiaji**

- GitHub: [@DarmawanKristiaji](https://github.com/DarmawanKristiaji)

## 🔗 Links

- **Production API:** [https://go-kasir-railway.dakr.my.id/](https://go-kasir-railway.dakr.my.id/)
- **Repository:** [https://github.com/DarmawanKristiaji/go-kasir](https://github.com/DarmawanKristiaji/go-kasir)
- **Railway Dashboard:** [Railway App](https://railway.app)

## 📖 Version History

- **v1.0.0** - Initial release
  - Product & Category CRUD
  - Layered architecture
  - Auto migrations
  - Railway deployment
  - IPv4 optimization

---

**Note:** This API is designed for educational purposes and production use. Always secure your environment variables and never commit sensitive data to version control.
