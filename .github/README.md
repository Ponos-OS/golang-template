# Golang Template

Base project setup for any Go program.

I build this template on top of what i found here: https://youtu.be/gCOAhSSjNMU

## Running Locally

1. Start PostgreSQL:

   ```bash
   docker compose up -d --wait
   ```

2. Initialize the project

   ```bash
   task init
   ```

3. Start the server in dev mode:

   ```bash
   task start:dev
   ```

4. Check that it is up:

   ```bash
   curl localhost:8080/health   # prints "ok"
   ```

> [!TIP]
>
> For OpenTelemetry in `.env`, set `OTEL_ENABLED=true`. Then the service can export **traces** and **logs** over OTLP/HTTP.
>
> Then if you `curl localhost:8080/api/v1/hello/` and `docker logs base-otel-collector` shows the span and the log records, with the same trace ID as the line on stdout.
>
> **Everything else is configured with the standard OTel environment variables (`OTEL_B*`).** The code does not parse them: the OTel SDK reads them itself, You can see them in the [OTel SDK configuration spec](https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/).
