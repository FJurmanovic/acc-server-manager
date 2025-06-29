# ACC Server Manager

A comprehensive web-based management system for Assetto Corsa Competizione (ACC) dedicated servers. This application provides a modern, secure interface for managing multiple ACC server instances with advanced features like automated Steam integration, firewall management, and real-time monitoring.

## 🚀 Features

### Core Server Management
- **Multi-Server Support**: Manage multiple ACC server instances from a single interface
- **Configuration Management**: Web-based configuration editor with validation
- **Service Integration**: Windows Service management via NSSM
- **Port Management**: Automatic port allocation and firewall rule creation
- **Real-time Monitoring**: Live server status and performance metrics

### Steam Integration
- **Automated Installation**: Automatic ACC server installation via SteamCMD
- **Credential Management**: Secure Steam credential storage with AES-256 encryption
- **Update Management**: Automated server updates and maintenance

### Security Features
- **JWT Authentication**: Secure token-based authentication system
- **Role-Based Access**: Granular permission system with user roles
- **Rate Limiting**: Protection against brute force and DoS attacks
- **Input Validation**: Comprehensive input sanitization and validation
- **Security Headers**: OWASP-compliant security headers
- **Password Security**: Bcrypt password hashing with strength validation

### Monitoring & Analytics
- **State History**: Track server state changes and player activity
- **Performance Metrics**: Server performance and usage statistics
- **Activity Logs**: Comprehensive logging and audit trails
- **Dashboard**: Real-time overview of all managed servers

## 🏗️ Architecture

### Technology Stack
- **Backend**: Go 1.23.0 with Fiber web framework
- **Database**: SQLite with GORM ORM
- **Authentication**: JWT tokens with bcrypt password hashing
- **API Documentation**: Swagger/OpenAPI integration
- **Dependency Injection**: Uber Dig container

### Project Structure
```
acc-server-manager/
├── cmd/
│   └── api/                    # Application entry point
├── local/
│   ├── api/                    # API route definitions
│   ├── controller/             # HTTP request handlers
│   ├── middleware/             # Authentication and security middleware
│   ├── model/                  # Database models and business logic
│   ├── repository/             # Data access layer
│   ├── service/                # Business logic services
│   └── utl/                    # Utilities and shared components
│       ├── cache/              # Caching utilities
│       ├── command/            # Command execution utilities
│       ├── common/             # Common utilities
│       ├── configs/            # Configuration management
│       ├── db/                 # Database connection and migration
│       ├── jwt/                # JWT token management
│       ├── logging/            # Logging utilities
│       ├── network/            # Network utilities
│       ├── password/           # Password hashing utilities
│       ├── regex_handler/      # Regular expression utilities
│       ├── server/             # HTTP server configuration
│       └── tracking/           # Server state tracking
├── docs/                       # Documentation
├── logs/                       # Application logs
└── vendor/                     # Go dependencies
```

## 📋 Prerequisites

### System Requirements
- **Operating System**: Windows 10/11 or Windows Server 2016+
- **Go**: Version 1.23.0 or later
- **SteamCMD**: For ACC server installation and updates
- **NSSM**: Non-Sucking Service Manager for Windows services
- **PowerShell**: Version 5.0 or later

### Dependencies
- ACC Dedicated Server files
- Valid Steam account (for server installation)
- Administrative privileges (for service and firewall management)

## ⚙️ Installation

### 1. Clone the Repository
```bash
git clone <repository-url>
cd acc-server-manager
```

### 2. Install Dependencies
```bash
go mod download
```

### 3. Generate Environment Configuration
We provide scripts to automatically generate secure secrets and create your `.env` file:

**Windows (PowerShell):**
```powershell
.\scripts\generate-secrets.ps1
```

**Linux/macOS (Bash):**
```bash
./scripts/generate-secrets.sh
```

**Manual Setup:**
If you prefer to set up manually:
```bash
copy .env.example .env
```

Then generate secure secrets:
```bash
# JWT Secret (64 bytes, base64 encoded)
openssl rand -base64 64

# Application secrets (32 bytes, hex encoded)
openssl rand -hex 32

# Encryption key (16 bytes, hex encoded = 32 characters)
openssl rand -hex 16
```

Edit `.env` with your generated secrets:
```env
# Security Settings (REQUIRED)
JWT_SECRET=your-generated-jwt-secret-here
APP_SECRET=your-generated-app-secret-here
APP_SECRET_CODE=your-generated-secret-code-here
ENCRYPTION_KEY=your-generated-32-character-hex-key

# Core Application Settings
PORT=3000
CORS_ALLOWED_ORIGIN=http://localhost:5173
DB_NAME=acc.db
PASSWORD=change-this-default-admin-password
```

### 4. Build the Application
```bash
go build -o api.exe cmd/api/main.go
```

### 5. Run the Application
```bash
./api.exe
```

The application will be available at `http://localhost:3000`

## 🔧 Configuration

### Environment Variables

The application uses minimal environment variables, with most settings managed through the web interface:

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `JWT_SECRET` | Yes | - | JWT signing secret (64+ chars, base64) |
| `APP_SECRET` | Yes | - | Application secret key (32 bytes, hex) |
| `APP_SECRET_CODE` | Yes | - | Application secret code (32 bytes, hex) |
| `ENCRYPTION_KEY` | Yes | - | AES-256 encryption key (32 hex chars) |
| `PORT` | No | 3000 | HTTP server port |
| `DB_NAME` | No | acc.db | SQLite database filename |
| `CORS_ALLOWED_ORIGIN` | No | http://localhost:5173 | CORS allowed origin |
| `PASSWORD` | No | - | Default admin password for initial setup |

**⚠️ Important**: All required secrets are automatically generated by the provided scripts in `scripts/` directory.

### System Configuration (Web Interface)

Advanced settings are managed through the web interface and stored in the database:
- **Steam Integration**: SteamCMD path and credentials
- **Service Management**: NSSM path and service settings  
- **Server Settings**: Default ports, firewall rules
- **Security Policies**: Rate limits, session timeouts
- **Monitoring**: Logging levels, performance tracking
- **Backup Settings**: Automatic backup configuration

Access these settings through the admin panel after initial setup.

## 🔒 Security

This application implements comprehensive security measures:

### Authentication & Authorization
- **JWT Tokens**: Secure token-based authentication
- **Password Security**: Bcrypt hashing with strength validation
- **Role-Based Access**: Granular permission system
- **Session Management**: Configurable timeouts and lockouts

### Protection Mechanisms
- **Rate Limiting**: Multiple layers of rate limiting
- **Input Validation**: Comprehensive input sanitization
- **Security Headers**: OWASP-compliant HTTP headers
- **CORS Protection**: Configurable cross-origin restrictions
- **Request Limits**: Size and timeout limitations

### Monitoring & Logging
- **Security Events**: Authentication and authorization logging
- **Audit Trail**: Comprehensive activity logging
- **Threat Detection**: Suspicious activity monitoring

For detailed security information, see [SECURITY.md](docs/SECURITY.md).

## 📚 API Documentation

The application includes comprehensive API documentation via Swagger UI:
- **Local Development**: http://localhost:3000/swagger/
- **Interactive Testing**: Test API endpoints directly from the browser
- **Schema Documentation**: Complete request/response schemas

### Key API Endpoints

#### Authentication
- `POST /api/v1/auth/login` - User authentication
- `POST /api/v1/auth/register` - User registration
- `GET /api/v1/auth/me` - Get current user

#### Server Management
- `GET /api/v1/servers` - List all servers
- `POST /api/v1/servers` - Create new server
- `GET /api/v1/servers/{id}` - Get server details
- `PUT /api/v1/servers/{id}` - Update server
- `DELETE /api/v1/servers/{id}` - Delete server

#### Configuration
- `GET /api/v1/servers/{id}/config/{file}` - Get configuration file
- `PUT /api/v1/servers/{id}/config/{file}` - Update configuration
- `POST /api/v1/servers/{id}/restart` - Restart server

## 🖥️ Frontend Integration

This backend is designed to work with a modern web frontend. Recommended stack:
- **React/Vue/Angular**: Modern JavaScript framework
- **TypeScript**: Type safety and better development experience
- **Axios/Fetch**: HTTP client for API communication
- **WebSocket**: Real-time server status updates

### CORS Configuration
Configure `CORS_ALLOWED_ORIGIN` to match your frontend URL:
```env
CORS_ALLOWED_ORIGIN=http://localhost:3000,https://yourdomain.com
```

## 🛠️ Development

### Running in Development Mode
```bash
# Install air for hot reloading (optional)
go install github.com/cosmtrek/air@latest

# Run with hot reload
air

# Or run directly with go
go run cmd/api/main.go
```

### Database Management
```bash
# View database schema
sqlite3 acc.db ".schema"

# Backup database
copy acc.db acc_backup.db
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test package
go test ./local/service/...
```

## 🚀 Production Deployment

### 1. Generate Production Secrets
```bash
# Use the secret generation script for production
.\scripts\generate-secrets.ps1  # Windows
./scripts/generate-secrets.sh   # Linux/macOS
```

### 2. Build for Production
```bash
# Build optimized binary
go build -ldflags="-w -s" -o acc-server-manager.exe cmd/api/main.go
```

### 3. Security Checklist
- [ ] Generate unique production secrets (use provided scripts)
- [ ] Configure production CORS origins in `.env`
- [ ] Change default admin password immediately after first login
- [ ] Enable HTTPS with valid certificates
- [ ] Set up proper firewall rules
- [ ] Configure system paths via web interface
- [ ] Set up monitoring and alerting
- [ ] Test all security configurations

### 3. Service Installation
```bash
# Create Windows service using NSSM
nssm install "ACC Server Manager" "C:\path\to\acc-server-manager.exe"
nssm set "ACC Server Manager" DisplayName "ACC Server Manager"
nssm set "ACC Server Manager" Description "Assetto Corsa Competizione Server Manager"
nssm start "ACC Server Manager"
```

### 4. Monitoring Setup
- Configure log rotation
- Set up health check monitoring
- Configure alerting for critical errors
- Monitor resource usage and performance

## 🔧 Troubleshooting

### Common Issues

#### "JWT_SECRET environment variable is required"
**Solution**: Set the JWT_SECRET environment variable with a secure 32+ character string.

#### "Failed to connect database"
**Solution**: Ensure the application has write permissions to the database directory.

#### "SteamCMD not found"
**Solution**: Install SteamCMD and update the `STEAMCMD_PATH` environment variable.

#### "Permission denied creating firewall rule"
**Solution**: Run the application as Administrator for firewall management.

### Log Locations
- **Application Logs**: `./logs/app.log`
- **Error Logs**: `./logs/error.log`
- **Security Logs**: `./logs/security.log`

### Debug Mode
Enable debug logging:
```env
LOG_LEVEL=debug
DEBUG_MODE=true
```

## 🤝 Contributing

### Development Setup
1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes and add tests
4. Ensure all tests pass: `go test ./...`
5. Commit your changes: `git commit -m 'Add amazing feature'`
6. Push to the branch: `git push origin feature/amazing-feature`
7. Open a Pull Request

### Code Style
- Follow Go best practices and conventions
- Use `gofmt` for code formatting
- Add comprehensive comments for public functions
- Include tests for new functionality

### Security Considerations
- Never commit secrets or credentials
- Follow secure coding practices
- Test security features thoroughly
- Report security issues privately

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **Fiber Framework**: High-performance HTTP framework
- **GORM**: Powerful ORM for Go
- **Assetto Corsa Competizione**: The amazing racing simulation
- **Community**: Contributors and users who make this project possible

## 📞 Support

### Documentation
- [Security Guide](docs/SECURITY.md)
- [API Documentation](http://localhost:3000/swagger/)
- [Configuration Guide](docs/CONFIGURATION.md)

### Community
- **Issues**: Report bugs and request features via GitHub Issues
- **Discussions**: Join community discussions
- **Wiki**: Community-maintained documentation and guides

### Professional Support
For professional support, consulting, or custom development, please contact the maintainers.

---

**Happy Racing! 🏁**