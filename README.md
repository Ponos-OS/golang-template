# Golang Template

Base project setup for any Go program

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
