# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 4.x.x   | :white_check_mark: |
| < 4.0   | :x:                |

## Reporting a Vulnerability

If you discover a security vulnerability, please report it responsibly:

1. **Do not** open a public GitHub issue
2. Email the maintainer directly or use GitHub's private vulnerability reporting
3. Include details about the vulnerability and steps to reproduce
4. Allow reasonable time for a fix before public disclosure

## Security Considerations

### API Token Handling

- Your ProductPlan API token is stored locally and never transmitted to third parties
- The token is only sent to ProductPlan's official API endpoints (`app.productplan.com`)
- Tokens are read from environment variables, not stored in config files

### Network Security

- All API calls use HTTPS
- No data is cached to disk by default
- The server runs locally as a subprocess of your AI assistant

### Best Practices

- Keep your API token private
- Use environment variables for token storage
- Regularly rotate your ProductPlan API token
- Run the latest version for security updates
