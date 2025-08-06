# 📚 Go Library School API

A RESTful API built with Golang for managing a school library system. It handles book and user CRUD operations, book loans and returns, JWT-based authentication, and email notifications for overdue returns.

![Go Version](https://img.shields.io/badge/go-%3E%3D1.20-blue)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![Last Commit](https://img.shields.io/github/last-commit/adityamaulanazidqy/go-libraryschool-api)

## 🚀 Key Features

- 📖 Book management (add, edit, delete, search)
- 👤 User authentication (login, register, logout, password update)
- 🧑‍🏫 User profile management (Student, Manager, Librarian)
- 🔐 JWT middleware for endpoint protection
- 🧰 Redis for session/logout handling
- 📄 Swagger documentation

## 🌐 API Documentation

Live Swagger documentation is automatically available when the server is running:

```docs-swagger
http://localhost:8080/swagger/index.html
```

## 🛠️ Tech Stack

| Component      | Technology |
|----------------|------------|
| Language       | Go 1.20    |
| Framework      | Native     |
| Database       | MySQL 8.0  |
| Cache          | Redis 7    |
| Email Sending    | Gomail     |
| Documentation  | Swagger    |
| Logging        | Logrus     |
| Auth  | JWT Token      |

## 📁 Folder Structure

```bash
go-libraryschool/
│
├── controllers/         # Business logic for each feature
├── routes/              # HTTP routing per feature
├── middlewares/         # Middleware (JWT, Logging, Redis)
├── models/              # Request and response structs
├── repository/          # Database access functions
├── config/              # DB, Redis, Logger initialization
├── docs/                # Swagger documentation
├── helpers/             # Common utility functions (JSON response, etc)
├── assets/              # Static files (if any)
├── main.go              # Application entry point
└── .env                 # Environment configuration    
```

# 📦 Installation

## Prerequisites
- Go 1.20+
- MySQL 8.0+
- Redis 7+

## 1. Clone repository
```bash
git clone https://github.com/adityamaulanazidqy/go-libraryschool-api.git
cd go-libraryschool-api
```

## 2. Setup ``.env``

Create .env file based on your database & redis connection

```env
JWT_KEY = "your_key_jwt_token"
REDIS_ADDR = "localhost:6379"
REDIS_PASSWORD = "your_redis_password"
```

## 3. Install dependencies ``command prompt``

```dependencies
go mod tidy
```

## 4. Generate Swagger docs

```swagger
swag init
```

## 5. Run server

```running
go run main.go
```

## 6. Import Database

*Make sure you have MySQL running. Then import the SQL file to the database*

## 📄 License

Distributed under the MIT License. See LICENSE for more information.

## 📧 Contact

Aditya Maulana Zidqy  
Email: adityamaullana234@gmail.com  
GitHub: @adityamaulanazidqy

Project Link: https://github.com/adityamaulanazidqy/go-libraryschool-api
