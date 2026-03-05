# HTTP Interface Helpers

Optional helpers for HTTP-based services can live here.

Recommended approach:
- Keep middleware opt-in.
- Add only the pieces your service actually needs.
- Prefer small `net/http` compatible adapters first.

Included example:
- `load_shedding.go.example`: optional in-flight load shedding middleware snippet.
