# x402cli

CLI tool for interacting with x402-protected resources and facilitator services.

## Usage

```
x402cli <command> [flags]
```

## Commands

```
x402cli
├── browse       Fetch the /.well-known/x402 discovery document
├── check        Check if a resource requires payment (resource server)
├── pay          Pay for a resource with a payment payload (resource server)
├── supported    Query facilitator for supported schemes/networks
├── verify       Verify a payment payload (facilitator)
├── settle       Settle a payment payload (facilitator)
├── payload      Generate a payment payload with EIP-3009 authorization
├── req          Generate a payment requirements object
└── proof        Generate an ownership proof signature for a resource URL
```

### browse

Fetch the `/.well-known/x402` discovery document from a server. Returns the list of x402-protected endpoint URLs and optional metadata.

```
x402cli browse -u https://api.example.com
x402cli browse -u https://api.example.com -o discovery.json
```

Flags:
- `-u`, `--url` — base URL of the server (required)
- `-o`, `--output` — file path to write JSON output (default: stdout)

Example output:

```json
{
  "version": 1,
  "resources": [
    "https://api.example.com/api/data",
    "https://api.example.com/api/premium"
  ],
  "ownershipProofs": [
    "0xabc123...",
    "0xdef456..."
  ],
  "instructions": "This API provides premium data. Pay per request."
}
```

> NOTE: Since a resource may have multiple accepted `payTo` addresses (e.g. EVM, Solana, etc.) there may be multiple proofs associated with one resource.

### check

Check if a resource requires x402 payment. Outputs the full PaymentRequired JSON response if the resource returns 402.

```
x402cli check -r <url>
x402cli check -r <url> -m POST
x402cli check -r <url> -o requirements.json
```

Flags:
- `-r`, `--resource` — URL of the resource to check (required)
- `-m`, `--method` — HTTP method, GET or POST (default: GET)
- `-d`, `--data` — request body data (optional)
- `-o`, `--output` — file path to write JSON output (default: stdout)

### pay

Pay for a resource by sending a request with a `PAYMENT-SIGNATURE` header. Constructs the full PaymentPayload from the inner payload and requirements, base64 encodes it, and sends it to the resource server.

```
x402cli pay -r <url> -p <json|file> --req <json|file>
x402cli pay -r <url> -m POST -p payload.json --req requirements.json -d '{"key":"value"}'
```

Flags:
- `-r`, `--resource` — URL of the resource (required)
- `-m`, `--method` — HTTP method, GET or POST (default: GET)
- `-p`, `--payload` — inner payload as JSON or file path (required, output of `payload` command)
- `--req`, `--requirements` — PaymentRequirements as JSON or file path (required)
- `-d`, `--data` — request body as JSON string or file path (optional, sets Content-Type: application/json)
- `-o`, `--output` — file path to write response body (default: stdout)

On success (200), prints the response body and decodes the `PAYMENT-RESPONSE` settlement header to stderr. On 402, prints the PaymentRequired JSON.

### supported

Query a facilitator for its supported schemes, networks, and signers.

```
x402cli supported --facilitator <url>
x402cli supported -f <url>
```

### verify

Verify a payment payload against a facilitator. Takes a payload object and requirements object (as JSON strings or file paths).

```
x402cli verify -f <facilitator-url> -p <json|file> -r <json|file>
```

Flags:
- `-f`, `--facilitator` — facilitator URL (required)
- `-p`, `--payload` — payload object as JSON or file path (required)
- `-r`, `--requirement` — payment requirements as JSON or file path (required)

### settle

Settle a payment payload via a facilitator. Same interface as verify, but calls `/settle`.

```
x402cli settle -f <facilitator-url> -p <payload-json|file> -r <requirements-json|file>
```

Flags:
- `-f`, `--facilitator` — facilitator URL (required)
- `-p`, `--payload` — payload object as JSON or file path (required)
- `-r`, `--requirement` — payment requirements as JSON or file path (required)

### payload

Generate a payment payload with EIP-3009 authorization. Optionally signs with a private key.

```
x402cli payload --to <address> --value <amount> [options]
x402cli payload --req requirements.json --private-key 0x...
```

Flags:
- `--to` — recipient address (required)
- `--value` — amount in smallest unit (required)
- `--private-key` — hex private key for EIP-712 signing
- `--from` — payer address (derived from key if omitted)
- `--asset` — token contract address (required with --private-key)
- `--name` — EIP-712 domain name (required with --private-key)
- `--version` — EIP-712 domain version (required with --private-key)
- `--chain-id` — chain ID (required with --private-key)
- `--req`, `--requirements` — PaymentRequirements as JSON or file path (see note below)
- `--valid-after` — unix timestamp (default: now)
- `--valid-before` — unix timestamp (default: now + 10min)
- `--valid-duration` — seconds, alternative to --valid-before
- `--nonce` — hex bytes32 nonce (default: random)
- `-o`, `--output` — file path to write output (default: stdout)

When `--req` is provided, the requirements object is used to populate default values for `--to` (from `payTo`), `--value` (from `amount`), `--asset`, `--name` (from `extra.name`), `--version` (from `extra.version`), and `--chain-id` (parsed from `network`). Individual flags always override values from requirements.

### req

Generate a payment requirements object, either from individual flags or by fetching from a resource server.

```
x402cli req --scheme exact --network eip155:84532 --amount 10000
x402cli req -r http://localhost:3000/api/data
x402cli req -r http://localhost:3000/api/data -m POST -d '{"key":"value"}'
x402cli req -r http://localhost:3000/api/data -i 1 --amount 5000 -o requirements.json
```

Flags:
- `-r`, `--resource` — URL of resource to fetch requirements from (hits server, parses 402 response)
- `-m`, `--method` — HTTP method to use when fetching requirements (default: GET)
- `-d`, `--data` — request body data (optional)
- `-i`, `--index` — index into the accepts array (default: 0)
- `--scheme` — payment scheme (e.g. exact)
- `--network` — CAIP-2 network (e.g. eip155:84532)
- `--amount` — amount in smallest unit
- `--asset` — token contract address
- `--pay-to` — recipient address
- `--max-timeout` — max timeout in seconds
- `--extra-name` — EIP-712 domain name (e.g. USD Coin)
- `--extra-version` — EIP-712 domain version (e.g. 2)
- `-o`, `--output` — file path to write output (default: stdout)

When `-r` is provided, the command fetches the PaymentRequired response from the resource server and uses `accepts[index]` as the base. Individual flags override fields from the fetched requirements.

## Docker

A multi-stage Dockerfile is provided at `cmd/x402cli/Dockerfile`. It produces a minimal Alpine-based image containing only the `x402cli` binary.

### Building

```bash
# From the project root
docker build -f cmd/x402cli/Dockerfile -t x402cli .
```

### Running

```bash
# Run any CLI command
docker run --rm x402cli supported -f http://host.docker.internal:4020

# Pass a private key for signing operations
docker run --rm -e X402_FACILITATOR_PRIVATE_KEY=0x... x402cli payload \
  --req requirements.json --private-key 0x...
```

### Docker Compose

The CLI is defined as a service in the project-level `docker-compose.yml` under the `cli` profile. It won't start with `docker compose up` by default.

```bash
# Run CLI commands via compose (has network access to the facilitator service)
docker compose run --rm x402cli supported -f http://facilitator:4020
docker compose run --rm x402cli check -r https://api.example.com/data
```

### proof

Generate an EIP-191 personal sign ownership proof for a resource URL. The signature can be used in the `ownershipProofs` field of a `.well-known/x402` discovery response.

```
x402cli proof -r https://api.example.com --private-key 0x...
```

Flags:
- `-r`, `--resource` — URL to sign (required)
- `--private-key` — hex-encoded private key for signing (required)
