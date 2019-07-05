# Description
Dockerized authentication proxy for sidecar container pattern.

## Currently
1. Requires `Authorization` in header of type `Bearer`
2. Supports AWS Cognito as an authentication source
3. Logs requests & responses -- including headers, bodies
4. Formats incoming request body to `camelCase` (convenient for NodeJS applications)
5. Formats outgoing response body to `snake_case` (convenient for readability)
6. Validates JSON in request body
7. Supports `gzip` compression for responses if client supports (although turned off for now)

## To-Do
1. Automated tests
