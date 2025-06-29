# Security Guide for ACC Server Manager

## Overview

This document outlines the security features, best practices, and requirements for the ACC Server Manager application. Following these guidelines is essential for maintaining a secure deployment.

## 🔐 Authentication & Authorization

### JWT Token Security

- **Secret Key**: Must be at least 32 bytes long and cryptographically secure
- **Token Expiration**: Default 24 hours, configurable via environment
- **Refresh Strategy**: Implement token refresh before expiration
- **Storage**: Store tokens securely (httpOnly cookies recommended for web)

### Password Security

- **Hashing**: Uses bcrypt with cost factor 12
- **Requirements**: Minimum 8 characters, must include uppercase, lowercase, digit, and special character
- **Validation**: Real-time strength validation during registration/update
- **Storage**: Never store plain text passwords

### Rate Limiting

- **Global**: 100 requests per minute per IP
- **Authentication**: 5 attempts per 15 minutes per IP+User-Agent
- **API Endpoints**: 60 requests per minute per IP
- **Customizable**: Configurable via environment variables

## 🛡️ Security Headers

The application automatically sets the following security headers:

- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`
- `Referrer-Policy: strict-origin-when-cross-origin`
- `Content-Security-Policy: [configured policy]`
- `Permissions-Policy: [restricted permissions]`

## 🔒 Data Protection

### Encryption

- **Algorithm**: AES-256-GCM for sensitive data
- **Key Management**: 32-byte keys from environment variables
- **Usage**: Steam credentials and other sensitive configuration data

### Database Security

- **SQLite**: Default database with file-level security
- **Migrations**: Automatic password security upgrades
- **Backup**: Encrypted backups with retention policies

## 🌐 Network Security

### HTTPS

- **Production**: HTTPS enforced in production environments
- **Certificates**: Use valid SSL/TLS certificates
- **Redirection**: Automatic HTTP to HTTPS redirect

### CORS Configuration

- **Origins**: Configured per environment
- **Headers**: Properly configured for API access
- **Credentials**: Enabled for authenticated requests

### Firewall Rules

- **Automatic**: Creates Windows Firewall rules for server ports
- **Management**: Centralized firewall rule management
- **Cleanup**: Automatic rule removal when servers are deleted

## 🚨 Input Validation & Sanitization

### Request Validation

- **Content-Type**: Validates expected content types
- **Size Limits**: 10MB request body limit
- **User-Agent**: Blocks suspicious user agents
- **Timeout**: 30-second request timeout

### Input Sanitization

- **XSS Prevention**: Removes dangerous HTML/JavaScript patterns
- **SQL Injection**: Uses parameterized queries
- **Path Traversal**: Validates file paths and names

## 📊 Monitoring & Logging

### Security Events

- **Authentication**: All login attempts (success/failure)
- **Authorization**: Permission checks and violations
- **Rate Limiting**: Blocked requests and patterns
- **Suspicious Activity**: Automated threat detection

### Log Security

- **Sensitive Data**: Never logs passwords or tokens
- **Format**: Structured logging with security context
- **Retention**: Configurable log retention policies
- **Access**: Restricted access to log files

## ⚙️ Environment Configuration

### Required Environment Variables

```bash
# Critical Security Settings
JWT_SECRET=<64-character-base64-string>
APP_SECRET=<32-character-hex-string>
APP_SECRET_CODE=<32-character-hex-string>
ENCRYPTION_KEY=<32-character-hex-string>

# Security Features
FORCE_HTTPS=true
RATE_LIMIT_GLOBAL=100
RATE_LIMIT_AUTH=5
SESSION_TIMEOUT=60
MAX_LOGIN_ATTEMPTS=5
LOCKOUT_DURATION=15
```

### Secret Generation

Generate secure secrets using:

```bash
# JWT Secret (Base64, 64 bytes)
openssl rand -base64 64

# Application Secrets (Hex, 32 bytes)
openssl rand -hex 32

# Encryption Key (Hex, 32 bytes)
openssl rand -hex 32
```

## 🔄 Security Migrations

### Password Security Upgrade

The application includes an automatic migration that:

1. Upgrades old encrypted passwords to bcrypt hashes
2. Maintains data integrity during the process
3. Provides rollback protection
4. Logs migration status and errors

### Migration Safety

- **Backup**: Automatically creates password backups
- **Validation**: Verifies password strength requirements
- **Recovery**: Handles corrupted or invalid passwords
- **Logging**: Detailed migration logs for auditing

## 🚀 Deployment Security

### Production Checklist

- [ ] Generate unique secrets for production
- [ ] Enable HTTPS with valid certificates
- [ ] Configure appropriate CORS origins
- [ ] Set up proper firewall rules
- [ ] Enable security monitoring and alerting
- [ ] Configure secure backup strategies
- [ ] Review and adjust rate limits
- [ ] Set up log monitoring and analysis
- [ ] Test security configurations
- [ ] Document security procedures

### Container Security (if applicable)

- **Base Images**: Use official, minimal base images
- **User Privileges**: Run as non-root user
- **Secrets**: Use container secret management
- **Network**: Isolate containers appropriately

## 🔍 Security Testing

### Automated Testing

- **Dependencies**: Regular security scanning of dependencies
- **SAST**: Static application security testing
- **DAST**: Dynamic application security testing
- **Penetration Testing**: Regular security assessments

### Manual Testing

- **Authentication Bypass**: Test authentication mechanisms
- **Authorization**: Verify permission controls
- **Input Validation**: Test input sanitization
- **Rate Limiting**: Verify rate limiting effectiveness

## 🚨 Incident Response

### Security Incident Procedures

1. **Detection**: Monitor logs and alerts
2. **Assessment**: Evaluate impact and scope
3. **Containment**: Isolate affected systems
4. **Eradication**: Remove threats and vulnerabilities
5. **Recovery**: Restore normal operations
6. **Lessons Learned**: Document and improve

### Emergency Contacts

- **Security Team**: [Configure your security team contacts]
- **System Administrators**: [Configure admin contacts]
- **Management**: [Configure management contacts]

## 📋 Security Maintenance

### Regular Tasks

- **Weekly**: Review security logs and alerts
- **Monthly**: Update dependencies and security patches
- **Quarterly**: Security configuration review
- **Annually**: Comprehensive security assessment

### Monitoring

- **Failed Logins**: Monitor authentication failures
- **Rate Limit Hits**: Track rate limiting events
- **Error Patterns**: Identify suspicious error patterns
- **Performance**: Monitor for DoS attacks

## 🔗 Additional Resources

### Security Standards

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [NIST Cybersecurity Framework](https://www.nist.gov/cyberframework)
- [CIS Controls](https://www.cisecurity.org/controls/)

### Go Security

- [Go Security Policy](https://golang.org/security)
- [Secure Coding Practices](https://github.com/OWASP/Go-SCP)

### Dependencies

- [Fiber Security](https://docs.gofiber.io/api/middleware/helmet)
- [GORM Security](https://gorm.io/docs/security.html)

## 📞 Support

For security questions or concerns:

- **Security Issues**: Report via private channels
- **Documentation**: Refer to this guide and code comments
- **Updates**: Monitor security advisories for dependencies

## 🔄 Version History

- **v1.0.0**: Initial security implementation
- **v1.1.0**: Added password security migration
- **v1.2.0**: Enhanced rate limiting and monitoring

---

**Important**: This security guide should be reviewed and updated regularly as the application evolves and new security threats emerge.