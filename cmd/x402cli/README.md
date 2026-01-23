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

Verify a payment payload against a facilitator. Accepts input as a JSON file or inline JSON string.

```
x402cli verify --facilitator <url> --file <path>
x402cli verify -f <url> -d '<json>'
```

See `verify.example.json` for an example request structure.
