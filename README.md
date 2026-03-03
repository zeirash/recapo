# Recapo - Order Management System

A modern order management system for Jastipers (Indonesian cross-border social media sellers), built with a microservices architecture.

## 🏗️ Project Structure

```
recapo/
├── arion/           # Backend API Service (Go)
├── oncius/          # Frontend Application (Next.js)
└── README.md        # This file
```

## 🚀 Quick Start

### Prerequisites
- Go 1.20+
- Node.js 18+
- PostgreSQL 15+
- Docker

### Backend (Arion)
```bash
cd arion
go mod download
go run main.go
```
Server runs on: `http://localhost:4000`

### Frontend (Oncius)
```bash
cd oncius
npm install
npm run dev
```
Application runs on: `http://localhost:3000`

## 📋 Features

### Core Features
- **User Authentication**: JWT-based login/register system
- **User Management**: Profile management and user data
- **Order Management**: Complete CRUD operations for orders
- **Product Management**: Product catalog and inventory
- **Customer Management**: Customer database and relationships
- **Multi-currency Support**: International transaction support

### Technical Features
- **RESTful API**: Clean, RESTful backend architecture
- **Real-time Data**: React Query for efficient data fetching
- **Mobile-First**: Responsive design optimized for mobile
- **Type Safety**: Full TypeScript implementation
- **Error Handling**: Comprehensive error handling and logging
- **Security**: JWT authentication, password hashing, CORS protection

## 🚀 Deployment

### Backend (Arion)
```bash
cd arion
docker build -t arion .
docker run -p 4000:4000 arion
```

### Frontend (Oncius)
```bash
cd oncius
npm run build
npm start
```

### Database (Postgres)
```bash
cd recapo
docker compose up -d postgres
```

### Docker Compose (Full Stack)
```bash
docker compose up -d
```

## 📋 Logs

```bash
# View logs for a specific service
docker compose logs backend
docker compose logs frontend

# Follow logs in real time
docker compose logs -f backend

# View last N lines
docker compose logs --tail=100 backend
```

## 🔒 Security Features

- **JWT Authentication**: Secure token-based authentication
- **Password Hashing**: bcrypt for password security
- **CORS Protection**: Configured cross-origin resource sharing
- **Input Validation**: Server-side validation for all inputs
- **Error Handling**: Secure error messages without data leakage

## 🧪 Development

### Backend Development
```bash
cd arion
go mod download
go run main.go
```

### Frontend Development
```bash
cd oncius
npm install
npm run dev
```

### Database Setup
```bash
# Using Docker Compose
docker compose up -d postgres
```

## 📊 Performance

- **Backend**: High-performance Go with efficient database queries
- **Frontend**: Next.js optimizations with React Query caching
- **Database**: Optimized PostgreSQL queries with proper indexing
- **Caching**: React Query for client-side data caching

## 📄 License

This project is licensed under the MIT License.

## 🆘 Support

For support and questions:
1. Check the documentation
2. Search existing issues
3. Create a new issue with detailed information

## 🔗 Related Projects

- [Arion Backend](./arion/) - Go REST API service
- [Oncius Frontend](./oncius/) - Next.js frontend application
