# x402cli

CLI tool for interacting with x402-protected resources and facilitator services.

## Usage

```
x402cli <command> [flags]
```

## Commands

### check

Check if a resource requires x402 payment.

```
x402cli check --resource <url>
x402cli check -r <url>
```

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

### payload

Generate a payment payload with EIP-3009 authorization. Optionally signs with a private key.

```
x402cli payload --to <address> --value <amount> [options]
```

Flags:
- `--to` — recipient address (required)
- `--value` — amount in smallest unit (required)
- `--private-key` — hex private key for EIP-712 signing
- `--from` — payer address (derived from key if omitted)
- `--valid-after` — unix timestamp (default: now)
- `--valid-before` — unix timestamp (default: now + 10min)
- `--valid-duration` — seconds, alternative to --valid-before
- `--nonce` — hex bytes32 nonce (default: random)
- `-o`, `--output` — file path to write output (default: stdout)
