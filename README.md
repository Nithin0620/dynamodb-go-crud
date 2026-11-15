# Saltwake E-Commerce Setup & Run Guide

This project is a full-stack E-Commerce application with:
* **Backend**: Go with Chi router and AWS DynamoDB (Single-table compatible or multi-table).
* **Frontend**: Modern React + Vite application with the animated **Saltwake Pirate Ship Two-Scene Hero**.

---

## 1. Prerequisites
- **Go** (1.20+)
- **Node.js** (v18+) & **npm**
- **DynamoDB**: Either AWS credentials configured (`~/.aws/credentials`) or a local DynamoDB instance (e.g. LocalStack or DynamoDB Local).

---

## 2. Running the Backend

From the project root:

```bash
# 1. Download Go dependencies
go mod tidy

# 2. Start the Go backend server (runs on http://localhost:8080)
go run cmd/app/main.go
```

The backend automatically creates tables (`products`, `users`, `orders`) on startup if they do not exist.

---

## 3. Running the React Frontend

Open a new terminal window:

```bash
# 1. Navigate to the frontend directory
cd frontend

# 2. Install dependencies (already installed if done previously)
npm install

# 3. Start the Vite development server (runs on http://localhost:5173)
npm run dev
```

Open your browser to:
👉 **`http://localhost:5173`**

*(Note: The Vite development server automatically proxies all API requests `/auth`, `/products`, `/orders`, and `/admin` to your Go backend at `http://localhost:8080`)*.

---

## 4. Building Frontend for Production (Optional)

If you prefer to serve the entire app from **a single port (8080)**:

```bash
# In frontend/ directory:
npm run build
```

Then simply start the Go server:
```bash
go run cmd/app/main.go
```
Now visiting **`http://localhost:8080`** serves the built React app and the API together!

---

## 5. Walkthrough & Features to Try

1. **Animated Hero Scene**:
   - Move your cursor around to test the hand-drawn custom cursor.
   - Click **"Set sail"** (or scroll the mouse wheel) to see the transition: the sea rises, the ship sails into the horizon, and the pirate island slides in.
2. **Store Owner / Admin**:
   - Click **"Account"** $\rightarrow$ **"Sign articles"** $\rightarrow$ select role **"Store Owner (Quartermaster)"** and register.
   - Click **"Captain's Deck"** $\rightarrow$ add trade goods (e.g. Compass, Telescope, Cutlass) with price and stock quantity.
3. **Customer Browsing & Riddle Payment**:
   - Open **"Bazaars"** to browse the available products.
   - Click **"Stow in Cargo"** to add items to your hold.
   - Click **"Cargo Hold"** $\rightarrow$ click **"Pay & Unlock Cargo with Riddle"**.
   - Solve the math challenge (e.g., `8 + 5 = 13`) to confirm the order and watch the store inventory decrement in real-time!

---

## 6. Docker Deployment (No Cloud Provider Required)

The repository includes everything needed to run the full stack as Docker containers:

| File | Purpose |
| --- | --- |
| `Dockerfile` | Multi-stage build: builds the React frontend, builds the Go backend, and produces one runtime image |
| `docker-compose.yml` | Runs the API image plus **DynamoDB Local** with a persistent volume |
| `.env.example` | Example environment values for Docker Compose |
| `.dockerignore` | Keeps build contexts small |

The API container serves both the REST API and the built React frontend from `http://localhost:8080`. DynamoDB Local runs in its own container and stores data in the `dynamodb-data` volume.

### 6.1 Quick start

```bash
cp .env.example .env

# Optional: change the JWT secret and ports in .env before starting
docker compose up -d --build
```

Check the containers:

```bash
docker compose ps
```

The `api` service should become `healthy` within about 20 seconds. Then open:

- Application: `http://localhost:8080`
- Health check: `http://localhost:8080/health`
- DynamoDB Local endpoint: `http://localhost:8000` (development access only)

### 6.2 Use custom ports

If ports `8080` or `8000` are already in use, change them in `.env`:

```env
API_PORT=18080
DYNAMODB_PORT=18000
```

Then restart:

```bash
docker compose up -d --build
```

The app is now available at `http://localhost:18080`.

### 6.3 Test the Docker deployment

```bash
# Health
curl -s http://localhost:8080/health

# Public product list
curl -s http://localhost:8080/products

# Frontend
curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/

# Register an admin user
curl -s -X POST http://localhost:8080/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"name":"Docker Admin","email":"admin@example.com","password":"secret123","role":"admin"}'

# Login and capture the token
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@example.com","password":"secret123"}' \
  | python3 -c 'import sys,json;print(json.load(sys.stdin)["result"]["token"])')

# Create a product as the admin
curl -s -X POST http://localhost:8080/admin/products \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Docker Compass","description":"Works in a container","price":19.99,"stock":7}'

# Verify the product is publicly visible
curl -s http://localhost:8080/products
```

Expected results:

- `/health` returns `200` with `"DynamoDB is alive"`.
- `/` returns `200` with `text/html`.
- Registration and login return `200`.
- Product creation returns `200` with an `id`.
- `/products` includes the created product.

### 6.4 Data persistence

```bash
# Stop containers but keep data
docker compose down

# Start again; previous data is still there
docker compose up -d

# Stop containers and delete the DynamoDB volume
docker compose down -v
```

### 6.5 Logs and troubleshooting

```bash
docker compose logs -f api
docker compose logs -f dynamodb
```

The API waits up to 60 seconds for DynamoDB to become reachable before starting. Tables (`products`, `users`, `orders`) are created automatically on startup.

If you see a Docker error like `Bind for 0.0.0.0:8000 failed: port is already allocated`, another process is using that port. Change `API_PORT` / `DYNAMODB_PORT` in `.env` or stop the other process.

### 6.6 Where you can deploy this

This setup is fully self-contained and does **not** require any managed cloud service. You can run it on any machine with Docker and Docker Compose:

- A local laptop or desktop
- A home server or mini PC
- An on-premises VM (Proxmox, VMware, Hyper-V, etc.)
- A NAS that supports Docker (Synology, QNAP, TrueNAS, Unraid)
- A Raspberry Pi or other ARM board (DynamoDB Local runs via `linux/amd64` emulation)
- A self-hosted Kubernetes/k3s cluster
- Self-hosted Docker management platforms such as Portainer, Coolify, Dokploy, CapRover, or Docker Swarm

Typical deployment flow on a remote machine:

```bash
git clone <your-repository-url>
cd dynamodb-go-crud
cp .env.example .env
# Edit .env: set JWT_SECRET and any port changes
docker compose up -d --build
```

Put a reverse proxy such as Caddy, Traefik, or Nginx in front of port `8080` if you need HTTPS or a domain name.

### 6.7 Important limitations

- **DynamoDB Local is intended for development and testing**, not production. It does not provide the durability, backups, replication, IAM, or scaling guarantees of real DynamoDB.
- Data lives in the Docker volume on a single host. Back up the volume if the data matters.
- `JWT_SECRET` must be changed from the development default before exposing the app to anyone.
- Do not expose port `8000` (DynamoDB Local) publicly. The API reaches it over the internal Compose network.
- The API container serves the frontend from `/app/frontend/dist`; rebuild the image after frontend changes.
- To deploy only the API image elsewhere, push the built image to any container registry or load it with `docker save` / `docker load`, then set `DYNAMODB_ENDPOINT` to a reachable DynamoDB endpoint.
