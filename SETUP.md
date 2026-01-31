# Go Kasir - Setup & Deployment Guide

## Production Deployment (Railway)

### 1. Add Environment Variables to Railway

Go to: https://railway.app/dashboard
1. Select **go-kasir** project
2. Click on **go-kasir** service
3. Go to **Variables** tab
4. Add these variables:

#### Variable 1: PORT
- **Key:** `PORT`
- **Value:** `8080`

#### Variable 2: DB_CONN (IMPORTANT!)
- **Key:** `DB_CONN`
- **Value:** 
```
host=db.vrllrcvihvbnxsqjqfoo.supabase.co port=5432 user=postgres password=RVkX9HvXb*.+P.a dbname=postgres sslmode=require
```

5. Click **Save** - Railway will auto-redeploy

### API Endpoints

Once DB_CONN is configured, all endpoints will be available:

#### Health Check
```
GET https://go-kasir-production-1268.up.railway.app/health
```

#### Categories
```
GET    /categories        - List all categories
POST   /categories        - Create category
GET    /categories/{id}   - Get category by ID
PUT    /categories/{id}   - Update category
DELETE /categories/{id}   - Delete category
```

#### Products (with Category Join)
```
GET    /api/produk        - List all products with category info
POST   /api/produk        - Create product
GET    /api/produk/{id}   - Get product with category
PUT    /api/produk/{id}   - Update product
DELETE /api/produk/{id}   - Delete product
```

## Local Development

### Setup
```bash
# Install Go 1.23+ from https://go.dev/dl

# Clone repository
git clone https://github.com/DarmawanKristiaji/go-kasir.git
cd "Go Kasir"

# Install dependencies
go mod download

# Create .env file with:
PORT=8080
DB_CONN=host=db.vrllrcvihvbnxsqjqfoo.supabase.co port=5432 user=postgres password=RVkX9HvXb*.+P.a dbname=postgres sslmode=require
```

### Run Locally
```bash
go run main.go
# Server runs on http://localhost:8080
```

### Build
```bash
go build -o app.exe .
.\app.exe
```

## Architecture

### Layered Architecture Pattern
```
HTTP Request
    ↓
Handlers (HTTP routing)
    ↓
Services (Business logic)
    ↓
Repositories (Database operations)
    ↓
Database (Supabase PostgreSQL)
```

### File Structure
```
go-kasir/
├── main.go                  # Entry point, DI setup
├── .env                     # Environment configuration
├── go.mod / go.sum         # Dependencies
├── database/
│   └── database.go         # PostgreSQL connection
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
└── migrations/
    └── init.sql            # Database schema
```

## Key Features

✅ **Session 2 - Task 1: Layered Architecture for Categories**
- Repository: Full CRUD database operations
- Service: Business logic layer
- Handler: HTTP routing with proper error handling
- Dependency Injection pattern

✅ **Session 2 - Challenge: Product-Category Relationship**
- Added `category_id` foreign key to products
- Products include category information in responses
- LEFT JOIN queries for category data
- Support for category assignment when creating/updating products

## Troubleshooting

### Database Connection Errors

If you see "Database not connected" errors:

1. **Check Railway Variables:**
   - Verify `DB_CONN` is set in Railway Variables tab
   - Verify `PORT` is set (default: 8080)

2. **Test Connection Locally:**
   ```bash
   $env:DB_CONN="host=db.vrllrcvihvbnxsqjqfoo.supabase.co port=5432 user=postgres password=RVkX9HvXb*.+P.a dbname=postgres sslmode=require"
   go run main.go
   ```

3. **Check Logs:**
   - Health endpoint shows database status: `GET /health`
   - Should show `"database":"connected"`

### Connection String Format

For Railway, use **key=value DSN format**:
```
host=<host> port=<port> user=<user> password=<password> dbname=<db> sslmode=require
```

NOT URL format:
```
postgresql://user:password@host:port/dbname
```

## Production URL

https://go-kasir-production-1268.up.railway.app
