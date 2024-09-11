# API for retrieving info from and managing DNS server

## Usage:
1. Set DNS_SERVER environment variable in `<HOST>:<PORT>` format. Example: `export DNS_SERVER=dns.local.domain:53`
2. Set TSIG_FILE environment variable to tsig file location. Example: `export TSIG_FILE=/var/tsig.json`
3. Run server with `go run .`

## TSIG_FILE Format:
| Key | Description | Example
| :--- | :--- | :--- |
| `"name"` | Name in cononical name format | `"tsig-key."` |
| `"algorithm"` | Name of HMAC algorithm in cononical name format | `"hmac-sha256."` |
| `"secret"` | Base64 encoded secret | `"c2VjcmV0c2VjcmV0c2VjcmV0Cg=="`

Example tsig.json:
```json
{
    "name": "tsig-key.",
    "algorithm": "hmac-sha256.",
    "secret": "c2VjcmV0c2VjcmV0c2VjcmV0Cg=="
}
```

## API Endpoints
### GET /api/v1/records/{zone}
Get all records for a zone

| Path | Required | Description |
| :--- | :--- | :--- |
| `zone` | `Yes` | Zone to lookup |

#### Example curl:
`curl http://dns-manager.example.com:8080/api/v1/records/local.domain.`

### POST /api/v1/records/{zone}
Create a record in a zone

| Path | Required | Description |
| :--- | :--- | :--- |
| `zone` | `Yes` | Zone to create the record in |

#### Example curl:
`curl http://dns-manager.example.com:8080/api/v1/records/local.domain. -X POST -d @./record.json`
#### Example record.json:
```json
{
  "name": "test.local.domain.",
  "type": "TypeA",
  "class": "ClassINET",
  "ttl": 300,
  "data": {
    "address": "10.10.10.10"
  }
}
```

### DELETE /api/v1/records/{zone}
Delete a record from a zone

| Path | Required | Description |
| :--- | :--- | :--- |
| `zone` | `Yes` | Zone to delete the record from |

#### Example curl:
`curl http://dns-manager.example.com:8080/api/v1/records/local.domain. -X DELETE -d @./record.json`
#### Example record.json:
```json
{
  "name": "test.local.domain.",
  "type": "TypeA",
  "class": "ClassINET",
  "ttl": 300,
  "data": {
    "address": "10.10.10.10"
  }
}
```
