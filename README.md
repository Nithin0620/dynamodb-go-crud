# dynamodb-go-crud

A small Go REST API for CRUD operations on products, backed by Amazon DynamoDB.

## Features

- Create, list, read, update, and delete products
- Health check endpoint
- Automatic `products` table creation on startup
- Request logging, timeouts, request IDs, panic recovery, and CORS middleware
- Support for DynamoDB Local via `DYNAMODB_ENDPOINT`

## Project structure

```text
cmd/app/                     Application entry point
config/                      Environment-based configuration
internal/controllers/        Business logic / CRUD operations
internal/entities/           Shared base entity and Product model
internal/handlers/           HTTP handlers for health and products
internal/repository/         DynamoDB adapter and connection setup
internal/routes/             Router, middleware, and route registration
internal/rules/              Validation and table migration rules
utils/                       Environment, HTTP response, and logger helpers
```

## Requirements

- Go 1.27.4 or newer
- DynamoDB access (AWS account or DynamoDB Local)
- Docker (only for the DynamoDB Local option)
- `curl` (for the manual API tests)

Install dependencies:

```bash
go mod tidy
```

## Configuration

| Variable | Default | Description |
| --- | --- | --- |
| `PORT` | `8080` | HTTP port the server listens on |
| `TIMEOUT` | `30` | Request timeout in seconds |
| `AWS_REGION` | `us-east-1` | AWS region for DynamoDB |
| `AWS_ACCESS_KEY_ID` | - | AWS access key |
| `AWS_SECRET_ACCESS_KEY` | - | AWS secret key |
| `DYNAMODB_ENDPOINT` | - | Optional custom endpoint, e.g. DynamoDB Local |
| `DIALECT` | `dynamodb` | Reserved configuration value |
| `DATABASE_URI` | - | Reserved configuration value |

On startup, the application creates the `products` table if it does not already exist. The table uses `_id` (string) as its partition key.

---

## Option A: Run against DynamoDB Local (recommended for development)

### 1. Start DynamoDB Local

```bash
docker run -d --rm --name ddb-local -p 8000:8000 amazon/dynamodb-local
```

### 2. Start the application

```bash
export AWS_ACCESS_KEY_ID=dummy
export AWS_SECRET_ACCESS_KEY=dummy
export AWS_REGION=us-east-1
export DYNAMODB_ENDPOINT=http://localhost:8000
export PORT=8080

go run ./cmd/app
```

Expected output:

```text
Waiting for the service to start on port: 8080
Table found: products
Server is running on port: :8080
```

### 3. Stop DynamoDB Local when done

```bash
docker rm -f ddb-local
```

> DynamoDB Local stores data in memory by default. All products are lost when the container is removed.

---

## Option B: Run against real AWS

### 1. Configure AWS credentials

```bash
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=your_access_key
export AWS_SECRET_ACCESS_KEY=your_secret_key
```

You can also use an existing shared AWS profile instead of exporting keys.

### 2. Start the application

```bash
go run ./cmd/app
```

The application attempts to create the `products` table automatically. If your credentials do not allow `CreateTable`, create the table manually with:

- Table name: `products`
- Partition key: `_id` (String)
- Capacity mode: On-demand or provisioned (the app defaults to 10 read/write units)
- Billing mode: Provisioned (default) or `PAY_PER_REQUEST`

### 3. Remove the environment variables when done

```bash
unset AWS_REGION AWS_ACCESS_KEY_ID AWS_SECRET_ACCESS_KEY
```

---

## Build, format, and static checks

```bash
go mod tidy
go build ./...
go vet ./...
gofmt -l .
```

`gofmt -l .` should print nothing. To automatically format files:

```bash
gofmt -w .
```

There are currently no automated test files, so `go test ./...` will report `no test files`. Use the manual API tests below to verify behavior.

---

## API overview

All responses use this envelope:

```json
{
  "status": 200,
  "result": {}
}
```

Errors use the same envelope:

```json
{
  "status": 400,
  "result": {
    "error": "message"
  }
}
```

| Method | Path | Description | Success |
| --- | --- | --- | --- |
| `GET` | `/health` | Check DynamoDB connectivity | `200` |
| `GET` | `/product/` | List all products | `200` |
| `GET` | `/product/{ID}` | Get one product | `200` |
| `POST` | `/product/` | Create a product | `200` |
| `PUT` | `/product/{ID}` | Update a product | `204` |
| `DELETE` | `/product/{ID}` | Delete a product | `204` |
| `OPTIONS` | `/product/` | CORS preflight | `204` |

Health endpoint behavior:

- `GET /health` or `GET /health/` returns `200` when DynamoDB is reachable.
- `POST`, `PUT`, and `DELETE` on `/health` return `405`.
- `OPTIONS` on `/health` returns `204`.

Product validation rules:

- `name` is required.
- `name` must be between 3 and 50 characters.
- `ID` must be a valid UUID when supplied in the URL.

---

## How to test everything

The examples below assume the server is running on `http://localhost:8080`.

### 1. Health check

```bash
curl -i http://localhost:8080/health
```

Expected: `HTTP/1.1 200 OK` and a body containing `"DynamoDB is alive"`.

Check the unsupported health methods:

```bash
curl -i -X POST http://localhost:8080/health
curl -i -X PUT http://localhost:8080/health
curl -i -X DELETE http://localhost:8080/health
curl -i -X OPTIONS http://localhost:8080/health
```

Expected: `405` for `POST`/`PUT`/`DELETE`, `204` for `OPTIONS`.

### 2. Create a product

```bash
curl -i -X POST http://localhost:8080/product/ \
  -H 'Content-Type: application/json' \
  -d '{"name":"widget"}'
```

Expected: `HTTP/1.1 200 OK` and a body similar to:

```json
{
  "status": 200,
  "result": {
    "ID": "f90ed7ec-9db4-41f0-902f-53b2211ca442"
  }
}
```

Copy the returned `ID` value. The commands below use `<ID>` as a placeholder.

On Linux/macOS you can capture it automatically with `jq`:

```bash
ID=$(curl -s -X POST http://localhost:8080/product/ \
  -H 'Content-Type: application/json' \
  -d '{"name":"widget"}' | jq -r '.result.ID')

echo "$ID"
```

### 3. List products

```bash
curl -i http://localhost:8080/product/
```

Expected: `200` with an array of products in `result`.

### 4. Get one product

```bash
curl -i http://localhost:8080/product/<ID>
```

Expected: `200` with the product in `result`.

### 5. Update a product

```bash
curl -i -X PUT http://localhost:8080/product/<ID> \
  -H 'Content-Type: application/json' \
  -d '{"name":"gadget"}'
```

Expected: `HTTP/1.1 204 No Content`.

Verify the update:

```bash
curl -s http://localhost:8080/product/<ID>
```

Expected: `name` is now `gadget`, and `updatedAt` has changed.

### 6. Delete a product

```bash
curl -i -X DELETE http://localhost:8080/product/<ID>
```

Expected: `HTTP/1.1 204 No Content`.

Verify it is gone:

```bash
curl -i http://localhost:8080/product/<ID>
```

Expected: `HTTP/1.1 404 Not Found`.

### 7. Test error cases

Invalid UUID:

```bash
curl -i http://localhost:8080/product/not-a-uuid
```

Expected: `400 Bad Request`.

Missing or invalid JSON body:

```bash
curl -i -X POST http://localhost:8080/product/ \
  -H 'Content-Type: application/json' \
  -d '{}'

curl -i -X POST http://localhost:8080/product/ \
  -H 'Content-Type: application/json' \
  -d 'not-json'
```

Expected: `400 Bad Request`.

Name too short:

```bash
curl -i -X POST http://localhost:8080/product/ \
  -H 'Content-Type: application/json' \
  -d '{"name":"ab"}'
```

Expected: `400 Bad Request` with a validation error.

Update or delete a missing product:

```bash
curl -i -X PUT http://localhost:8080/product/00000000-0000-4000-8000-000000000000 \
  -H 'Content-Type: application/json' \
  -d '{"name":"gadget"}'

curl -i -X DELETE http://localhost:8080/product/00000000-0000-4000-8000-000000000000
```

Expected: `404 Not Found`.

### 8. Full smoke-test script

With DynamoDB Local or AWS configured and the server running, run:

```bash
set -e

BASE_URL=http://localhost:8080

echo "Health:"
curl -s "$BASE_URL/health"
echo

echo "Create:"
CREATE=$(curl -s -X POST "$BASE_URL/product/" \
  -H 'Content-Type: application/json' \
  -d '{"name":"widget"}')
echo "$CREATE"

ID=$(echo "$CREATE" | grep -o '"ID":"[^"]*"' | cut -d'"' -f4)
echo "ID: $ID"

echo "List:"
curl -s "$BASE_URL/product/"
echo

echo "Get:"
curl -s "$BASE_URL/product/$ID"
echo

echo "Update:"
curl -s -o /dev/null -w "%{http_code}\n" -X PUT "$BASE_URL/product/$ID" \
  -H 'Content-Type: application/json' \
  -d '{"name":"gadget"}'

echo "Get after update:"
curl -s "$BASE_URL/product/$ID"
echo

echo "Delete:"
curl -s -o /dev/null -w "%{http_code}\n" -X DELETE "$BASE_URL/product/$ID"

echo "Get after delete (expect 404):"
curl -s -o /dev/null -w "%{http_code}\n" "$BASE_URL/product/$ID"
```

Expected final status codes: `204` for update, `204` for delete, and `404` for the final get.
