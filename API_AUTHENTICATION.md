# 🔐 API Authentication Guide

## Overview

The WhatsApp API Gateway uses API Key authentication to secure all endpoints. There are multiple ways to provide your API key for authentication.

## 🔑 API Key Types

### 1. Global API Key
- **Purpose**: Access to all sessions and session management endpoints
- **Configuration**: Set in `.env` file as `GLOBAL_API_KEY`
- **Default**: `your-super-secret-global-api-key-here`
- **Scope**: Full access to all API endpoints

### 2. Session API Key
- **Purpose**: Access to specific session endpoints
- **Configuration**: Generated when creating a session
- **Scope**: Limited to specific session operations

## 🛡️ Authentication Method

### Authorization Header (Only Method Accepted)
```bash
curl -X GET "http://localhost:8080/sessions/list" \
  -H "Authorization: your-super-secret-global-api-key-here"
```

**Note**: The API uses a simplified authentication method with only the `Authorization` header containing your API key directly (no Bearer or ApiKey prefix).

## 📚 Swagger UI Authentication

### Using Swagger UI (http://localhost:8080/swagger/index.html)

1. **Open Swagger UI** in your browser
2. **Click the "Authorize" button** (🔒 icon) at the top right
3. **Enter your API key**:
   - Field: `Authorization`
   - Value: `your-super-secret-global-api-key-here` (no prefix needed)
4. **Click "Authorize"**
5. **Test endpoints** - they will now include your API key automatically

## 🚨 Security Best Practices

### Production Setup
```bash
# Generate a strong API key
GLOBAL_API_KEY=$(openssl rand -hex 32)

# Update your .env file
echo "GLOBAL_API_KEY=$GLOBAL_API_KEY" >> .env

# Set production mode
echo "GIN_MODE=release" >> .env
```

### Environment Variables
```bash
# Required
GLOBAL_API_KEY=your-super-secret-global-api-key-here

# Optional (defaults)
SERVER_PORT=8080
GIN_MODE=debug
```

## 📋 API Response Examples

### ✅ Successful Authentication
```json
{
  "success": true,
  "code": 200,
  "data": {
    "sessions": [...]
  }
}
```

### ❌ Missing API Key
```json
{
  "error": "API key required"
}
```

### ❌ Invalid API Key
```json
{
  "error": "Invalid API key"
}
```

## 🔧 Testing Authentication

### Test with curl
```bash
# Test health endpoint (no auth required)
curl http://localhost:8080/health

# Test sessions endpoint (auth required)
curl -H "Authorization: your-super-secret-global-api-key-here" \
     http://localhost:8080/sessions/list
```

### Test with Postman
1. **Create new request**
2. **Set URL**: `http://localhost:8080/sessions/list`
3. **Go to Authorization tab**
4. **Select "API Key"**
5. **Set**:
   - Key: `Authorization`
   - Value: `your-super-secret-global-api-key-here`
   - Add to: `Header`

## 🎯 Quick Start

1. **Update your API key**:
   ```bash
   # Edit .env file
   GLOBAL_API_KEY=my-secret-api-key-123
   ```

2. **Start the server**:
   ```bash
   make up && make run
   ```

3. **Test authentication**:
   ```bash
   curl -H "Authorization: my-secret-api-key-123" \
        http://localhost:8080/sessions/list
   ```

4. **Use Swagger UI**:
   - Open: http://localhost:8080/swagger/index.html
   - Click "Authorize" 🔒
   - Enter your API key
   - Test endpoints!

## 🔗 Related Documentation

- [Main README](README.md) - General setup and usage
- [API Reference](API.md) - Complete API documentation
- [Architecture](ARCHITECTURE.md) - System architecture
