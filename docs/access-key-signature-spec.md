# Access Key Signature Spec

This document defines the canonical AK/SK signing flow used by SDK clients and `aegisctl` to exchange an access key for a short-lived bearer token.

## Portal Workflow

Recommended operator flow:

1. Sign in to the Portal with your normal human account.
2. Open the access-key management page and create an access key for the specific automation use case.
3. Copy the returned `access_key` and one-time `secret_key` immediately and store them in your secret manager.
4. Use the signed-header token exchange flow in SDKs, `aegisctl`, CI, or other automation.
5. Rotate or disable the key from Portal when the automation changes or is no longer needed.

Portal manages the credential lifecycle, while runtime callers only use:

- `access_key`
- `secret_key`
- `POST /api/v2/auth/access-key/token`

## Endpoint

- `POST /api/v2/auth/access-key/token`

This endpoint is the only place where `secret_key` is used directly. All normal business APIs still use:

- `Authorization: Bearer <token>`

## Required Headers

Every token exchange request must include these headers:

- `X-Access-Key`: the access key identifier, for example `ak_xxx`
- `X-Timestamp`: unix timestamp in seconds, for example `1713333333`
- `X-Nonce`: caller-generated unique nonce, max length `128`
- `X-Signature`: lowercase hex `HMAC-SHA256`

## Canonical String

The signature payload is the following newline-joined canonical string:

```text
METHOD
PATH
ACCESS_KEY
TIMESTAMP
NONCE
```

For the token exchange endpoint, the canonical string looks like:

```text
POST
/api/v2/auth/access-key/token
ak_demo
1713333333
abc123
```

Rules:

- `METHOD` must be uppercase, for example `POST`
- `PATH` is the request path only, without scheme, host, or query string
- `ACCESS_KEY`, `TIMESTAMP`, and `NONCE` must exactly match the transmitted headers

## Signature Algorithm

Compute the signature as:

```text
signature = hex(hmac_sha256(secret_key, canonical_string))
```

Details:

- hash: `SHA-256`
- MAC: `HMAC`
- output encoding: lowercase hexadecimal
- secret material: raw `secret_key`

## Verification Rules

The server currently enforces:

- timestamp must be within `+- 5 minutes`
- nonce is single-use inside the validity window
- repeated nonce submissions are rejected as replay attempts
- disabled, deleted, or expired access keys cannot issue tokens

Replay protection is implemented with Redis-backed nonce reservation.

## Request Example

```http
POST /api/v2/auth/access-key/token HTTP/1.1
Host: aegislab.example.com
Accept: application/json
Content-Type: application/json
X-Access-Key: ak_demo
X-Timestamp: 1713333333
X-Nonce: abc123
X-Signature: 4cf2f2cbb93d...
```

The request body is empty.

## curl Example

The following example shows a full shell flow from `access_key` / `secret_key` to bearer token:

```bash
ACCESS_KEY="ak_demo"
SECRET_KEY="sk_demo"
SERVER="http://localhost:8082"
PATH_URI="/api/v2/auth/access-key/token"
TIMESTAMP="$(date +%s)"
NONCE="$(openssl rand -hex 16)"
CANONICAL="POST\n${PATH_URI}\n${ACCESS_KEY}\n${TIMESTAMP}\n${NONCE}"
SIGNATURE="$(printf '%b' "${CANONICAL}" | openssl dgst -sha256 -hmac "${SECRET_KEY}" -hex | awk '{print $2}')"

curl -X POST "${SERVER}${PATH_URI}" \
  -H "Accept: application/json" \
  -H "X-Access-Key: ${ACCESS_KEY}" \
  -H "X-Timestamp: ${TIMESTAMP}" \
  -H "X-Nonce: ${NONCE}" \
  -H "X-Signature: ${SIGNATURE}"
```

After receiving the response, extract `data.token` and use it as:

```http
Authorization: Bearer <jwt>
```

## Response Usage

On success, the endpoint returns a bearer token payload similar to:

```json
{
  "code": 0,
  "message": "Access key token issued successfully",
  "data": {
    "token": "<jwt>",
    "token_type": "Bearer",
    "expires_at": "2026-04-17T12:00:00Z",
    "auth_type": "access_key",
    "access_key": "ak_demo"
  }
}
```

Clients must use the returned JWT for subsequent business API calls:

```http
Authorization: Bearer <jwt>
```

Do not send `X-Access-Key` / `X-Signature` headers to normal business endpoints.

## aegisctl Debug Helpers

`aegisctl` provides two local debugging commands for signature issues:

- `aegisctl auth inspect`: inspect the stored auth context, token source, expiry, and access key metadata
- `aegisctl auth sign-debug --access-key ... --secret-key ...`: print the canonical string, signed headers, and a ready-to-run curl example
- `aegisctl auth sign-debug --execute`: execute the signed token exchange request immediately and print the API response
- `aegisctl auth sign-debug --execute --save-context`: execute the signed request and persist the returned bearer token into the current CLI context
