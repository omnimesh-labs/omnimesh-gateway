# Security Policy

## Supported Versions

We release patches for security vulnerabilities in the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take the security of Omnimesh AI Gateway seriously. If you believe you have found a security vulnerability, please report it to us as described below.

### How to Report a Security Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please report them via email to: **team@wraithscan.com**

If you don't receive a response within 48 hours, please follow up via GitHub by mentioning @michaelmcclelland.

### What to Include in Your Report

To help us better understand the nature and scope of the possible issue, please include as much of the following information as possible:

- Type of issue (e.g. buffer overflow, SQL injection, cross-site scripting, etc.)
- Full paths of source file(s) related to the manifestation of the issue
- The location of the affected source code (tag/branch/commit or direct URL)
- Any special configuration required to reproduce the issue
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact of the issue, including how an attacker might exploit it

### Response Process

1. **Acknowledgment**: We will acknowledge receipt of your vulnerability report within 48 hours.

2. **Initial Assessment**: We will perform an initial assessment within 5 business days and respond with our evaluation.

3. **Investigation**: We will investigate the issue and determine its validity and severity.

4. **Fix Development**: If confirmed, we will develop and test a fix.

5. **Release**: We will release the fix as soon as possible, typically within 30 days for high-severity issues.

6. **Disclosure**: After the fix is released, we will publicly disclose the vulnerability (with credit to you if desired).

## Security Features

Omnimesh AI Gateway implements several security measures:

### Authentication & Authorization
- JWT-based authentication with configurable expiry
- Role-based access control (RBAC)
- API key authentication for service-to-service communication
- Token blacklisting for secure logout

### Rate Limiting & Protection
- IP-based rate limiting with Redis backing
- Per-user and per-organization rate limits
- DDoS protection with configurable thresholds
- Request size limiting

### Data Protection
- Secure password hashing with bcrypt
- Environment variable-based configuration for secrets
- SQL injection prevention with parameterized queries
- Input validation and sanitization

### Network Security
- HTTPS/TLS support with configurable certificates
- CORS protection
- Security headers middleware
- Request timeout protection

### Audit & Monitoring
- Comprehensive audit logging
- Security event tracking
- Failed authentication monitoring
- Anomaly detection capabilities

## Security Best Practices for Users

When deploying Omnimesh AI Gateway:

1. **Use HTTPS in production** - Always enable TLS encryption
2. **Secure your JWT secret** - Use a strong, random secret key
3. **Configure rate limiting** - Set appropriate limits for your use case
4. **Enable audit logging** - Monitor security events
5. **Regular updates** - Keep Omnimesh AI Gateway updated to the latest version
6. **Environment variables** - Never commit secrets to version control
7. **Database security** - Use secure database credentials and SSL connections
8. **Network isolation** - Deploy in secure network environments

## Known Security Considerations

- **Development Configuration**: The development config includes example credentials that must be changed in production
- **Database Access**: Ensure PostgreSQL is properly secured and not exposed publicly
- **Redis Security**: If using Redis for rate limiting, ensure it's properly secured
- **Log Files**: Audit logs may contain sensitive information - secure appropriately

## Dependencies

We regularly monitor our dependencies for known vulnerabilities and update them promptly. Our CI/CD pipeline includes:

- Automated dependency vulnerability scanning
- Go security analyzer (gosec)
- Regular dependency updates
- Security-focused linting rules

## Contact

For general security questions or concerns, please contact: team@wraithscan.com

For urgent security matters, please include "URGENT" in the subject line.

---

Thank you for helping keep Omnimesh AI Gateway and our users safe!
