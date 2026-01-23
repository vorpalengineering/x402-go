# x402cli

CLI tool for interacting with x402-protected resources and facilitator services.

## Usage

```
x402cli <command> [flags]
```

## Commands

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
- `-o`, `--output` — file path to write JSON output (default: stdout)

### pay

Pay for a resource by sending a request with a `PAYMENT-SIGNATURE` header. Constructs the full PaymentPayload from the inner payload and requirements, base64 encodes it, and sends it to the resource server.

```
x402cli pay -r <url> -p <payload-json|file> --req <requirements-json|file>
x402cli pay -r <url> -p payload.json --req requirements.json -o response.txt
```

Flags:
- `-r`, `--resource` — URL of the resource (required)
- `-m`, `--method` — HTTP method, GET or POST (default: GET)
- `-p`, `--payload` — inner payload as JSON or file path (required, output of `payload` command)
- `--req`, `--requirements` — PaymentRequirements as JSON or file path (required)
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
x402cli verify -f <facilitator-url> -p <payload-json|file> -r <requirements-json|file>
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
- `--req`, `--requirements` — PaymentRequirements as JSON or file path (populates defaults)
- `--valid-after` — unix timestamp (default: now)
- `--valid-before` — unix timestamp (default: now + 10min)
- `--valid-duration` — seconds, alternative to --valid-before
- `--nonce` — hex bytes32 nonce (default: random)
- `-o`, `--output` — file path to write output (default: stdout)

### req

Generate a payment requirements object. All field flags are optional.

```
x402cli req [options]
```

Flags:
- `--scheme` — payment scheme (e.g. exact)
- `--network` — CAIP-2 network (e.g. eip155:84532)
- `--amount` — amount in smallest unit
- `--asset` — token contract address
- `--pay-to` — recipient address
- `--max-timeout` — max timeout in seconds
- `--extra-name` — EIP-712 domain name (e.g. USD Coin)
- `--extra-version` — EIP-712 domain version (e.g. 2)
- `-o`, `--output` — file path to write output (default: stdout)
