# CoAP Webhooks Router

## To Run
- generate your ADMIN_BEARER (`openssl rand -base64 27` can do it) and export ADMIN_BEARER="generatedbearertokenvalue" in your shell
- `mkdir -f dbdata && make dev` for local run
- `make docker-run` for local docker run (functionally the same as the previous)
- `make docker-build` to build the docker image
- the Admin Hooks API runs on 8081, and the CoAP server runs on 5683

## Using the Hooks API

### Create a new hook
- `curl -v -d '{"owner":"{OWNER_NAME}","name":"{HOOK_NAME}","destination":"{HTTPS_ADDRESS}"}' -H "Content-Type:application/json" -H "Authorization: Bearer $ADMIN_BEARER"  http://localhost:8081/api/hooks`

### List all hooks for an owner
- `curl -v -H "Accept:application/json" -H "Authorization: Bearer $ADMIN_BEARER"  http://localhost:8081/api/hooks/{OWNER_NAME}`

### List all hooks existing in the system
- `curl -v -H "Accept:application/json" -H "Authorization: Bearer $ADMIN_BEARER"  http://localhost:8081/api/hooks`

### Delete an existing hook by name
- `curl -v -X DELETE -H "Content-Type:application/json" -H "Authorization: Bearer $ADMIN_BEARER"  http://localhost:8081/api/hooks/{OWNER_NAME}/{HOOK_NAME}`

### Delete all hooks for an owner
- `curl -v -X DELETE -H "Content-Type:application/json" -H "Authorization: Bearer $ADMIN_BEARER"  http://localhost:8081/api/hooks/{OWNER_NAME}`

## Making CoAP requests

Only CoAP POST requests are currently supported
