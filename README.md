# API for retrieving info and managing DNS server

## Usage:
1. Set DNS_SERVER environment variable in `<HOST>:<PORT>` format. Example: `export DNS_SERVER=dns.local.domain:53`
2. Run server with `go run .`

## API Endpoints
`/records`: Currently only returns all A records for a domain

| Parameter | Required | Description |
| :--- | :--- | :--- |
| `domain` | `Yes` | Domain to lookup |

Example:
```http
GET /records?domain=local.domain
```
