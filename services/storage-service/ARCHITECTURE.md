# Storage Service Architecture

## 🏗️ Microservices Architecture

### Correct Request Flow

```
┌─────────────┐
│   Client    │ (Browser, Mobile App, etc.)
│  (Frontend) │
└──────┬──────┘
       │ HTTP/HTTPS
       │ POST http://localhost:8000/api/upload
       ▼
┌─────────────────────────────────────────────────────────────┐
│                    Kong API Gateway                         │
│                      (Port 8000)                           │
│  • Authentication (JWT for protected endpoints)            │
│  • Rate Limiting                                           │
│  • CORS                                                    │
│  • Request Size Limiting (100MB)                           │
│  • Routing                                                 │
└──────┬──────────────────────────────────────────────────────┘
       │
       │ Routes /api/upload to:
       │ http://storage-service:8059/upload
       │
       ▼
┌─────────────────────────────────────────────────────────────┐
│              Storage Service (HTTP Server)                  │
│                    (Port 8059)                             │
│  • Receives multipart/form-data                            │
│  • Handles chunk uploads                                   │
│  • Tracks progress                                         │
└──────┬──────────────────────────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────────────────────────────┐
│                  Chunk Manager                              │
│  • Creates upload sessions                                 │
│  • Stores chunks temporarily                               │
│  • Assembles complete files                                │
│  • Auto-cleanup (24 hours)                                 │
└──────┬──────────────────────────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────────────────────────────┐
│                   FTP Client                                │
│  • Uploads to FTP server (production)                      │
│  • Or saves to local disk (testing/development)            │
└──────┬──────────────────────────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────────────────────────────┐
│              FTP Storage / File System                      │
│  • Organized by MIME type and date                         │
│  • upload/{mime-type}/{YYYY-MM-DD}/filename                │
└─────────────────────────────────────────────────────────────┘
```

## 🚫 Important: Direct Service Access

**❌ CLIENTS SHOULD NEVER ACCESS SERVICES DIRECTLY**

```
❌ Wrong:  http://localhost:8059/upload
✅ Correct: http://localhost:8000/api/upload
```

### Why?

1. **Security**: Kong provides authentication, rate limiting, and request validation
2. **Routing**: Services can be moved/scaled without changing client code
3. **Monitoring**: All traffic goes through a central point
4. **Flexibility**: Easy to add new policies (caching, transformation, etc.)

## 🔧 Kong Configuration

The Kong API Gateway is configured in `kong/kong.yml`:

```yaml
# Public upload endpoint (NO authentication required)
- name: storage-service-upload
  url: http://storage-service:8059
  protocol: http
  routes:
    - name: upload-route
      paths: ["/api/upload"]
      methods: ["POST", "OPTIONS"]
      strip_path: true  # Removes /api from path before forwarding
  plugins:
    - name: cors
    - name: request-size-limiting
      config:
        allowed_payload_size: 100  # 100MB
    - name: rate-limiting
      config:
        minute: 50
        hour: 1000
```

### How Kong Routes the Request

1. **Client sends:** `POST http://localhost:8000/api/upload`
2. **Kong receives:** Request on port 8000
3. **Kong matches:** Route with path `/api/upload`
4. **Kong strips:** Removes `/api` (because `strip_path: true`)
5. **Kong forwards:** `POST http://storage-service:8059/upload`
6. **Storage Service:** Processes the upload
7. **Kong returns:** Response to client

## 🎯 Dual Server Architecture

The Storage Service runs TWO servers simultaneously:

### 1. HTTP Server (Port 8059)
- **Purpose**: Handle REST API requests from Kong
- **Endpoints**: 
  - `POST /upload` - Chunk upload
  - `POST /api/upload` - Alternative path
  - `GET /health` - Health check
- **Access**: Via Kong API Gateway only

### 2. gRPC Server (Port 50059)
- **Purpose**: Handle internal microservice-to-microservice communication
- **Services**: 
  - `FileStorageService` - File operations
  - `ImageService` - Image management
- **Access**: Direct service-to-service (not through Kong)

## 🌍 Environment-Based Configuration

### Production (Docker Compose)

```yaml
# docker-compose.yml
storage-service:
  ports:
    - "50059:50059"  # gRPC (internal)
    - "8059:8059"    # HTTP (via Kong only)
  environment:
    HTTP_PORT: 8059
    GRPC_PORT: 50059
```

**Client Access:** `http://api.metarang.com/api/upload` → Kong → Storage Service

### Development (Local Testing)

**Option 1: Through Kong (Recommended)**
```bash
# Start full stack
docker-compose up -d

# Client accesses
http://localhost:8000/api/upload
```

**Option 2: Direct Access (Testing Only)**
```bash
# Start service standalone
go run test_server.go

# Direct access (ONLY for testing)
http://localhost:8059/upload
```

## 📋 Testing Scenarios

### 1. Production/Integration Testing (Through Kong)

```javascript
// HTML/JavaScript
const response = await fetch('http://localhost:8000/api/upload', {
    method: 'POST',
    body: formData
});
```

```bash
# cURL
curl -X POST http://localhost:8000/api/upload \
  -F "file=@test.jpg"
```

**Flow:** Client → Kong (8000) → Storage Service (8059)

### 2. Unit Testing (Direct Access)

```bash
# Start test server
go run test_server.go

# Test directly (bypassing Kong)
curl -X POST http://localhost:8059/upload \
  -F "file=@test.jpg"
```

**Flow:** Test → Storage Service (8059)

**Note:** Direct access is ONLY for development/testing. Never expose port 8059 in production.

## 🔒 Security Layers

### Layer 1: Network (Kong)
- ✅ Rate limiting (50 req/min, 1000 req/hour)
- ✅ Request size limiting (100MB max)
- ✅ CORS headers
- ✅ Request/response logging

### Layer 2: Service (Storage Service)
- ✅ Input validation
- ✅ Chunk size validation
- ✅ File type detection
- ✅ Error handling

### Layer 3: Storage (FTP/Filesystem)
- ✅ Directory permissions
- ✅ Unique filenames (MD5 hash)
- ✅ Organized structure
- ✅ Automatic cleanup

## 📊 Port Allocation

| Service | gRPC Port | HTTP Port | Access |
|---------|-----------|-----------|--------|
| Kong Gateway | - | 8000 | Public |
| Storage Service | 50059 | 8059 | Internal |
| Auth Service | 50051 | - | Internal |
| Commercial Service | 50052 | - | Internal |
| Features Service | 50053 | - | Internal |

**Public Endpoint:** Only Kong (8000) should be accessible from outside

## 🚀 Deployment

### Docker Compose

```bash
# Start all services including Kong
cd metarang-microservices
docker-compose up -d

# Verify Kong is routing correctly
curl http://localhost:8000/api/upload

# Check service health
curl http://localhost:8001/routes  # Kong Admin API
```

### Kubernetes

```yaml
# Kong will use service discovery
- name: storage-service-upload
  url: http://storage-service.metarang.svc.cluster.local:8059
```

## 📝 Summary

### ✅ Correct Architecture

```
Client → Kong (8000) → Storage Service (8059) → FTP/Storage
         ↓
      - Auth
      - Rate Limit
      - CORS
      - Logging
```

### ❌ Wrong Architecture

```
Client → Storage Service (8059) → FTP/Storage
         ↓
      NO security layers!
```

**Always use the API Gateway!** The only exception is during local development/testing when you need to debug the service directly.

## 🔗 Related Documentation

- [Kong Configuration](../../kong/kong.yml)
- [Docker Compose Setup](../../docker-compose.yml)
- [Upload Endpoint API](../../docs/UPLOAD_ENDPOINT.md)
- [Storage Service Implementation](../../docs/STORAGE_CHUNK_UPLOAD.md)

---

**Remember:** In production, clients should NEVER have direct access to microservices. All traffic MUST go through the API Gateway.

