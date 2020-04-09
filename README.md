# To Run
- `cp .env.example .env`
- generate your ADMIN_BEARER
- `make run`

# To create a new hook
- `curl -v -d '{"owner":"{OWNER_NAME}","name":"{HOOK_NAME}","destination":"{HTTPS_ADDRESS}"}' -H "Content-Type:application/json" -H "Authorization: Bearer {BEARER}"  http://localhost:8081/api/hooks/`

# To list all hooks for an owner
- `curl -v -H "Accept:application/json" -H "Authorization: Bearer {BEARER}"  http://localhost:8081/api/hooks/{OWNER_NAME}`

