# HomeWizard API Bridge

Small Go service that talks to a HomeWizard P1 Meter over the local v2 API and re-exposes selected values without authentication, so other consumers like Loxone can use them with very simple HTTP inputs.

## Disclaimer

This project was developed with AI-assisted tooling and reviewed manually.

## What it does

- Pairs with the HomeWizard P1 Meter through the documented v2 `POST /api/user` flow.
- Stores the returned bearer token locally in `.data/token.json`.
- Keeps HomeWizard-specific API details inside a small client layer.
- Exposes explicit, stable electricity, gas, and water endpoints that return plain text values.
- Keeps a few raw/debug routes available as JSON or plain text.

## Route contract

System and debug:

- `GET /`
- `GET /docs`
- `GET /docs.json`
- `GET /ui`
- `GET /healthz`
- `GET /auth/status`
- `POST /pair`
- `GET /status`
- `GET /api/device`
- `GET /api/measurement`
- `GET /api/telegram`

Electricity:

- `GET /electricity/usage`
- `GET /electricity/import-usage`
- `GET /electricity/export-usage`
- `GET /electricity/total-usage`
- `GET /electricity/total-export`
- `GET /electricity/tariff`
- `GET /electricity/usage-l1`
- `GET /electricity/usage-l2`
- `GET /electricity/usage-l3`
- `GET /electricity/current-total`
- `GET /electricity/current-l1`
- `GET /electricity/current-l2`
- `GET /electricity/current-l3`
- `GET /electricity/voltage`
- `GET /electricity/voltage-l1`
- `GET /electricity/voltage-l2`
- `GET /electricity/voltage-l3`
- `GET /electricity/frequency`
- `GET /electricity/average-demand-15m`
- `GET /electricity/monthly-peak`
- `GET /electricity/monthly-peak-timestamp`

Gas:

- `GET /gas/total-usage`
- `GET /gas/timestamp`
- `GET /gas/unit`

Water:

- `GET /water/total-usage`
- `GET /water/timestamp`
- `GET /water/unit`

## Environment

```bash
export HOMEWIZARD_HOST=192.168.1.50
export HOMEWIZARD_USERNAME=local/loxone-bridge
export HOMEWIZARD_INSECURE_SKIP_VERIFY=true
export BIND_ADDR=:8080
```

Optional:

```bash
export HOMEWIZARD_TOKEN=YOUR_EXISTING_TOKEN
export DATA_DIR=.data
```

## Run

```bash
go run .
```

The first time, pair it:

1. Start the bridge.
2. Open `http://localhost:8080/ui` in a browser.
3. Press `Start / Retry Pairing`.
4. If the page tells you pairing is not enabled yet, press the button on the P1 meter.
5. Press `Start / Retry Pairing` again within 30 seconds.

Once pairing succeeds, the token is saved and reused on restart.

If you prefer scripts, `POST http://localhost:8080/pair` still works too.

Built-in pages:

- `http://localhost:8080/ui` for pairing and auth monitoring
- `http://localhost:8080/docs` for Swagger UI docs
- `http://localhost:8080/docs.json` for the OpenAPI document

## Loxone examples

Get values as plain text:

```text
http://bridge-host:8080/electricity/usage
http://bridge-host:8080/electricity/total-usage
http://bridge-host:8080/gas/total-usage
http://bridge-host:8080/water/total-usage
```

Get one combined JSON snapshot:

```text
http://bridge-host:8080/status
```

## Notes

- HomeWizard documents v2 as HTTPS with bearer token auth. This bridge intentionally removes auth on its own HTTP routes for local-network use with Loxone.
- The bridge defaults to `HOMEWIZARD_INSECURE_SKIP_VERIFY=true` because HomeWizard uses device-local TLS certificates. If you want stricter validation later, we can extend the client with the HomeWizard CA certificate and expected certificate hostname.
- HomeWizard measurement fields are optional. If your smart meter does not provide a field, the endpoint returns `404` instead of inventing a value.
- Gas and water are taken from HomeWizard's `external` measurement array and only exist if your setup provides those meters.
