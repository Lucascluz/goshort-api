# GoShort API

GoShort is a URL shortening service built with Go, Gin, and SQLite. It provides user authentication (JWT), password reset via email (with OTP), and URL management (shorten, list, redirect).

## Features

- User registration and login with JWT authentication
- Password reset via email with OTP verification
- Shorten URLs and get a short link
- List all shortened URLs for a user
- Redirect from short URL to original URL
- Built-in SQLite database (no external DB required)
- Environment-based configuration via `.env` file

## Project Structure

```
.
├── main.go
├── configs/
├── internal/
│   ├── database/
│   ├── handlers/
│   ├── middleware/
│   ├── models/
│   └── routes/
├── pkg/
├── scripts/
├── .env
├── go.mod
└── go.sum
```

## Getting Started

### Prerequisites

- Go 1.18 or newer
- [Git](https://git-scm.com/)
- An SMTP account (e.g., Gmail) for password reset emails

### Setup

1. **Clone the repository**

   ```sh
   git clone https://github.com/yourusername/goshort-api.git
   cd goshort-api
   ```

2. **Configure environment variables**

   Copy `.env` and update values as needed:

   ```
   BASE_URL=localhost:8080
   DB_DSN=file:app.db
   JWT_SECRET=your_jwt_secret
   SMTP_HOST=smtp.gmail.com
   SMTP_PORT=587
   SMTP_ACCOUNT=your_email@gmail.com
   SMTP_PASSWORD=your_smtp_password
   ```

3. **Install dependencies**

   ```sh
   go mod tidy
   ```

4. **Run the application**

   ```sh
   go run main.go
   ```

   The server will start on the port specified in `.env` (default: 8080).

## API Endpoints

### Auth

- `POST /auth/register` — Register a new user
- `POST /auth/login` — Login and receive JWT
- `POST /auth/logout` — Logout (JWT required)

### Password Reset

- `POST /auth/password-reset/request` — Request OTP via email
- `GET /auth/password-reset/confirm` — Confirm OTP
- `POST /auth/password-reset/submit` — Set new password

### URL Management

- `POST /shorten` — Shorten a URL (JWT required)
- `GET /list` — List all URLs (JWT required)
- `GET /:shortKey` — Redirect to original URL

## Notes

- The SQLite database file will be created automatically.
- For production, set a strong `JWT_SECRET` and configure SMTP credentials securely.

## License

MIT