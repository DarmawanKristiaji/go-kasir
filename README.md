# Kasir API

Sistem kasir sederhana yang dibangun dengan Go. API ini menyediakan fitur CRUD lengkap untuk mengelola produk.

## Fitur

- ✅ Kelola data produk (CRUD lengkap)
- ✅ Terima request lewat HTTP
- ✅ Response dalam format JSON
- ✅ Siap di-deploy ke cloud (gratis!)

## Kebutuhan

- Go 1.21+
- Text editor (VSCode recommended)
- Terminal/Command Prompt
- Git (untuk deploy)

## Instalasi & Running

### Clone atau setup project

```bash
cd Go\ Kasir
go mod init kasir-api
```

### Run server

```bash
go run main.go
```

Server akan berjalan di `http://localhost:8080`

## API Endpoints

| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| GET | `/health` | Health check |
| GET | `/api/produk` | Ambil semua produk |
| POST | `/api/produk` | Buat produk baru |
| GET | `/api/produk/{id}` | Ambil produk by ID |
| PUT | `/api/produk/{id}` | Update produk |
| DELETE | `/api/produk/{id}` | Hapus produk |

## Testing dengan cURL

### Health Check
```bash
curl http://localhost:8080/health
```

### Get All Produk
```bash
curl http://localhost:8080/api/produk
```

### Create Produk
```bash
curl -X POST http://localhost:8080/api/produk \
  -H "Content-Type: application/json" \
  -d '{
    "nama": "Kopi Kapal Api",
    "harga": 2500,
    "stok": 200
  }'
```

### Get Produk by ID
```bash
curl http://localhost:8080/api/produk/1
```

### Update Produk
```bash
curl -X PUT http://localhost:8080/api/produk/1 \
  -H "Content-Type: application/json" \
  -d '{
    "nama": "Indomie Goreng Jumbo",
    "harga": 4000,
    "stok": 150
  }'
```

### Delete Produk
```bash
curl -X DELETE http://localhost:8080/api/produk/1
```

## Build Binary

### Build Standard
```bash
go build -o kasir-api
```

### Build Production (Smaller)
```bash
go build -ldflags="-s -w" -o kasir-api
```

### Cross-Compilation

Build untuk Windows (dari OS lain):
```bash
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o kasir-api.exe
```

Build untuk Linux (dari OS lain):
```bash
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o kasir-api
```

Build untuk Mac (dari OS lain):
```bash
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o kasir-api
```

## Jalankan Binary

### Mac/Linux
```bash
./kasir-api
```

### Windows
```bash
kasir-api.exe
```

## Deployment

### Railway

1. Push ke GitHub
2. Buka [railway.app](https://railway.app/)
3. Login dengan GitHub
4. Click "New Project"
5. Pilih "Deploy from GitHub repo"
6. Select repo `kasir-api`
7. Railway auto-detect Go → auto-deploy!
8. Dapatkan URL production

## Struktur Kode

- `main.go` - File utama dengan seluruh logic API
  - Package & Import
  - Struct Produk
  - In-memory storage
  - Health check endpoint
  - CRUD endpoints
  - Main routing function

## Data Struktur

```json
{
  "id": 1,
  "nama": "Indomie Godog",
  "harga": 3500,
  "stok": 10
}
```

## Status HTTP

- `200 OK` - Success
- `201 Created` - Resource created
- `400 Bad Request` - Invalid input
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

## Catatan

- Data disimpan in-memory (hilang saat restart)
- Sesi 2 akan menambahkan SQLite database untuk persistent storage
- ID otomatis increment berdasarkan jumlah data

---

Dokumentasi lengkap: [Kodingworks Tutorial](https://docs.kodingworks.io/s/01e57b74-74e6-44df-ac02-7e30a2478528)
