# API Quick Reference

This document provides a complete reference for all available endpoints in the E-Commerce backend service, including HTTP methods, access permissions, descriptions, sample request payloads, and response formats.

---

## Base URL & Headers

* **Base URL**: `http://localhost:8080` (or configured `PORT`)
* **Content-Type**: `application/json`
* **Protected Routes**: Add header `Authorization: Bearer <JWT_TOKEN>`

---

## 1. Authentication Endpoints

| Method | Endpoint | Access | Description | Input Format (Request Body) | Output Format (Response Body) |
|---|---|---|---|---|---|
| `POST` | `/auth/register` | Public | Register a new user or store owner | ```json<br>{<br>  "name": "John Doe",<br>  "email": "john@example.com",<br>  "password": "password123",<br>  "role": "user"<br>}<br>``` <br>*(role: `"user"` or `"admin"`)* | ```json<br>{<br>  "status": 200,<br>  "result": {<br>    "_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",<br>    "name": "John Doe",<br>    "email": "john@example.com",<br>    "role": "user",<br>    "createdAt": "2026-11-15T20:00:00.0000",<br>    "updatedAt": "2026-11-15T20:00:00.0000"<br>  }<br>}<br>``` |
| `POST` | `/auth/login` | Public | Authenticate user and receive JWT token | ```json<br>{<br>  "email": "john@example.com",<br>  "password": "password123"<br>}<br>``` | ```json<br>{<br>  "status": 200,<br>  "result": {<br>    "token": "eyJhbGciOiJIUzI1Ni...",<br>    "user": {<br>      "_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",<br>      "name": "John Doe",<br>      "email": "john@example.com",<br>      "role": "user"<br>    }<br>  }<br>}<br>``` |

---

## 2. Public Product Endpoints

| Method | Endpoint | Access | Description | Input Format (Request Body) | Output Format (Response Body) |
|---|---|---|---|---|---|
| `GET` | `/products` | Public | List all available products | *None* | ```json<br>{<br>  "status": 200,<br>  "result": [<br>    {<br>      "_id": "d3b07384-d113-4672-886d-3f0e8f85f340",<br>      "name": "Wireless Mouse",<br>      "description": "Ergonomic 2.4G mouse",<br>      "price": 29.99,<br>      "stock": 45,<br>      "ownerId": "a1b2c3d4-...",<br>      "createdAt": "2026-11-15T19:50:00.0000",<br>      "updatedAt": "2026-11-15T19:50:00.0000"<br>    }<br>  ]<br>}<br>``` |
| `GET` | `/products/{ID}` | Public | Get single product details by UUID | *None* (Path param `ID`) | ```json<br>{<br>  "status": 200,<br>  "result": {<br>    "_id": "d3b07384-d113-4672-886d-3f0e8f85f340",<br>    "name": "Wireless Mouse",<br>    "description": "Ergonomic 2.4G mouse",<br>    "price": 29.99,<br>    "stock": 45,<br>    "ownerId": "a1b2c3d4-...",<br>    "createdAt": "2026-11-15T19:50:00.0000",<br>    "updatedAt": "2026-11-15T19:50:00.0000"<br>  }<br>}<br>``` |

---

## 3. Customer Order & Puzzle Payment Endpoints

All endpoints below require header: `Authorization: Bearer <USER_JWT>`

| Method | Endpoint | Access | Description | Input Format (Request Body) | Output Format (Response Body) |
|---|---|---|---|---|---|
| `POST` | `/orders/checkout-challenge` | User | Step 1: Validate stock, get total amount & puzzle challenge | ```json<br>{<br>  "items": [<br>    {<br>      "productId": "d3b07384-d113-4672-886d-3f0e8f85f340",<br>      "quantity": 2<br>    }<br>  ]<br>}<br>``` | ```json<br>{<br>  "status": 200,<br>  "result": {<br>    "challengeId": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",<br>    "question": "Solve to confirm order: 7 + 6 = ?",<br>    "totalAmount": 59.98,<br>    "items": [<br>      {<br>        "productId": "d3b07384-d113-4672-886d-3f0e8f85f340",<br>        "productName": "Wireless Mouse",<br>        "price": 29.99,<br>        "quantity": 2,<br>        "ownerId": "a1b2c3d4-..."<br>      }<br>    ]<br>  }<br>}<br>``` |
| `POST` | `/orders/confirm` | User | Step 2: Answer puzzle, decrement stock & create paid order | ```json<br>{<br>  "challengeId": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",<br>  "answer": "13",<br>  "items": [<br>    {<br>      "productId": "d3b07384-d113-4672-886d-3f0e8f85f340",<br>      "quantity": 2<br>    }<br>  ]<br>}<br>``` | ```json<br>{<br>  "status": 200,<br>  "result": {<br>    "message": "Order successfully placed and paid!",<br>    "order": {<br>      "_id": "e8b9195b-21d7-40c2-b7e5-1c3905cf83d2",<br>      "userId": "f47ac10b-58cc-4372-a567-0e02b2c3d479",<br>      "totalAmount": 59.98,<br>      "status": "PAID",<br>      "items": [...],<br>      "createdAt": "2026-11-15T20:05:00.0000",<br>      "updatedAt": "2026-11-15T20:05:00.0000"<br>  }<br>}<br>``` |
| `GET` | `/orders/my-orders` | User | View orders placed by authenticated user | *None* | ```json<br>{<br>  "status": 200,<br>  "result": [<br>    {<br>      "_id": "e8b9195b-21d7-40c2-b7e5-1c3905cf83d2",<br>      "userId": "f47ac10b-58cc-4372-a567-0e02b2c3d479",<br>      "totalAmount": 59.98,<br>      "status": "PAID",<br>      "items": [...]<br>    }<br>  ]<br>}<br>``` |

---

## 4. Admin / Store Owner Endpoints

All endpoints below require header: `Authorization: Bearer <ADMIN_JWT>`

| Method | Endpoint | Access | Description | Input Format (Request Body) | Output Format (Response Body) |
|---|---|---|---|---|---|
| `GET` | `/admin/products` | Admin | List products belonging to authenticated store owner | *None* | ```json<br>{<br>  "status": 200,<br>  "result": [<br>    {<br>      "_id": "d3b07384-d113-4672-886d-3f0e8f85f340",<br>      "name": "Wireless Mouse",<br>      "price": 29.99,<br>      "stock": 43,<br>      "ownerId": "store-owner-uuid"<br>    }<br>  ]<br>}<br>``` |
| `POST` | `/admin/products` | Admin | Create a new product in store inventory | ```json<br>{<br>  "name": "Mechanical Keyboard",<br>  "description": "RGB Hot-swappable",<br>  "price": 89.99,<br>  "stock": 20<br>}<br>``` | ```json<br>{<br>  "status": 200,<br>  "result": {<br>    "id": "c1f7b98d-2e45-4209-8652-32a22be581e2"<br>  }<br>}<br>``` |
| `PUT` | `/admin/products/{ID}` | Admin | Update product details / stock | ```json<br>{<br>  "name": "Mechanical Keyboard Pro",<br>  "description": "RGB Hot-swappable with wrist rest",<br>  "price": 99.99,<br>  "stock": 25<br>}<br>``` | *204 No Content* |
| `DELETE` | `/admin/products/{ID}` | Admin | Remove product from store | *None* (Path param `ID`) | *204 No Content* |
| `GET` | `/admin/orders` | Admin | View orders containing items from this store owner | *None* | ```json<br>{<br>  "status": 200,<br>  "result": [<br>    {<br>      "_id": "e8b9195b-21d7-40c2-b7e5-1c3905cf83d2",<br>      "status": "PAID",<br>      "totalAmount": 59.98,<br>      "items": [<br>        {<br>          "productId": "d3b07384-...",<br>          "productName": "Wireless Mouse",<br>          "quantity": 2,<br>          "price": 29.99,<br>          "ownerId": "store-owner-uuid"<br>        }<br>      ]<br>    }<br>  ]<br>}<br>``` |
| `PUT` | `/admin/orders/{id}/status` | Admin | Update status of an order (`SHIPPED`, `DELIVERED`, etc.) | ```json<br>{<br>  "status": "SHIPPED"<br>}<br>``` | ```json<br>{<br>  "status": 200,<br>  "result": {<br>    "message": "order status updated"<br>  }<br>}<br>``` |

---

## 5. Health Check Endpoint

| Method | Endpoint | Access | Description | Input Format | Output Format |
|---|---|---|---|---|---|
| `GET` | `/health` | Public | Check if DynamoDB connection is alive | *None* | ```json<br>{<br>  "status": 200,<br>  "result": "Relational database is alive"<br>}<br>``` |
