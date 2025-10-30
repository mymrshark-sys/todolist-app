# Todo List Application

A full-stack todo list application built with Go, PostgreSQL, and Bootstrap 5.

## Features

- User Authentication (Register/Login)
- Create, Read, Update, Delete (CRUD) todo items
- Mark todos as complete/pending
- Filter todos by status
- Responsive design
- Session-based authentication
- Per-user todo items

## Prerequisites

- Go 1.20 or later
- PostgreSQL 12 or later
- Git (optional)

## Setup Instructions

1. Clone the repository (if using Git):
```bash
git clone <repository-url>
cd todolist-app
```

2. Set up the database:
```bash
# Connect to PostgreSQL
psql -U postgres

# Create the database
CREATE DATABASE todolist_db;

# Connect to the new database
\c todolist_db

# Run the schema file
\i schema.sql
```

3. Configure the database connection:
   - Open `main.go`
   - Update the `connStr` variable with your PostgreSQL credentials:
```go
connStr := "host=localhost port=5432 user=postgres password=your_password dbname=todolist_db sslmode=disable"
```

4. Install dependencies:
```bash
go mod tidy
```

5. Run the application:
```bash
go run main.go
```

6. Access the application:
   - Open http://localhost:8080 in your browser
   - Default login credentials:
     - Username: admin
     - Password: admin123

## Project Structure

```
todolist-app/
├── main.go              # Main application file
├── schema.sql           # Database schema
├── hash_generator.go    # Utility for password hashing
├── static/
│   ├── css/
│   │   └── style.css   # Custom styles
│   └── js/
│       └── app.js      # Frontend JavaScript
└── templates/
    ├── index.html      # Main application page
    ├── login.html      # Login page
    └── register.html   # Registration page
```

## API Endpoints

- Authentication:
  - POST `/api/login` - Login user
  - POST `/api/register` - Register new user
  - POST `/api/logout` - Logout user
  - GET `/api/user` - Get current user info

- Todo Operations:
  - GET `/api/notes` - Get all todos for current user
  - POST `/api/notes` - Create new todo
  - PUT `/api/notes/{id}` - Update todo
  - DELETE `/api/notes/{id}` - Delete todo
  - PATCH `/api/notes/{id}/toggle` - Toggle todo status

## Libraries Used

- Backend:
  - `github.com/gorilla/mux` - HTTP router
  - `github.com/gorilla/sessions` - Session management
  - `github.com/lib/pq` - PostgreSQL driver
  - `golang.org/x/crypto/bcrypt` - Password hashing

- Frontend:
  - Bootstrap 5 - UI framework
  - Bootstrap Icons - Icons

## Security Features

- Password hashing using bcrypt
- Session-based authentication
- CSRF protection via HTTP-only cookies
- SQL injection prevention using parameterized queries
- XSS prevention using HTML escaping
- Per-user data isolation

## Error Handling

The application includes comprehensive error handling:
- Database connection errors
- Authentication errors
- Input validation
- Not found errors
- Server errors

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.# todolist-app
