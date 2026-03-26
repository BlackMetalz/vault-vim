# Vault API Reference

How vault-vim interacts with the HashiCorp Vault HTTP API.

## Authentication

All requests include the header:
```
X-Vault-Token: <token>
```

Token is resolved from `VAULT_TOKEN` env var or `~/.vault-token` file.

## API Calls

### List KV v2 Mounts

```
GET /v1/sys/mounts
```

Response contains all mounted secret engines. vault-vim filters for entries where `type == "kv"` and `options.version == "2"`.

Example response (trimmed):
```json
{
  "secret/": {
    "type": "kv",
    "options": { "version": "2" }
  },
  "team-secrets/": {
    "type": "kv",
    "options": { "version": "2" }
  }
}
```

### List Secrets at Path

```
LIST /v1/{mount}metadata/{path}
```

Note: Uses HTTP method `LIST` (not GET). Returns directory entries (ending with `/`) and secret names.

Example: `LIST /v1/secret/metadata/myapp/`
```json
{
  "data": {
    "keys": ["dev", "prod", "staging"]
  }
}
```

Directories end with `/`:
```json
{
  "data": {
    "keys": ["myapp/", "infra/"]
  }
}
```

### Read Secret

```
GET /v1/{mount}data/{path}
```

Example: `GET /v1/secret/data/myapp/prod`
```json
{
  "data": {
    "data": {
      "db_host": "prod-db.example.com",
      "db_port": "5432",
      "db_user": "admin",
      "db_pass": "supersecret123"
    },
    "metadata": {
      "created_time": "2024-01-01T00:00:00Z",
      "version": 3
    }
  }
}
```

### Patch Secret (Update Single Key)

```
PATCH /v1/{mount}data/{path}
Content-Type: application/merge-patch+json

{
  "data": {
    "db_host": "new-host.example.com"
  }
}
```

This ONLY updates `db_host`. All other keys remain unchanged. This is the key operation that makes vault-vim safe — it never overwrites keys you didn't touch.

### Put Secret (Full Write)

```
POST /v1/{mount}data/{path}
Content-Type: application/json

{
  "data": {
    "key1": "value1",
    "key2": "value2"
  }
}
```

This REPLACES all keys. Used internally for:
- **Delete key**: Read all keys → remove one → write back
- **Rename key**: Read all keys → add new name → remove old name → write back
- **Create secret**: Write initial key-value pairs

### Delete Key (Composite Operation)

No native Vault API for deleting a single key. vault-vim does:

1. `GET /v1/{mount}data/{path}` — read all keys
2. Remove the target key from the map
3. `POST /v1/{mount}data/{path}` — write back remaining keys

### Rename Key (Composite Operation)

No native Vault API for renaming a key. vault-vim does:

1. `GET /v1/{mount}data/{path}` — read all keys
2. Copy value from old key to new key name
3. Delete old key from the map
4. `POST /v1/{mount}data/{path}` — write back

## Required Vault Policies

For vault-vim to work, the authenticated user needs these capabilities:

```hcl
# List available mounts
path "sys/mounts" {
  capabilities = ["read"]
}

# Full access to secrets (adjust paths as needed)
path "secret/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
```

Minimum for read-only browsing:
```hcl
path "sys/mounts" {
  capabilities = ["read"]
}

path "secret/*" {
  capabilities = ["read", "list"]
}
```
