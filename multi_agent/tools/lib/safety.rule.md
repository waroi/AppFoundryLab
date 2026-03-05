# Safety and Redaction Rule

Purpose:
- Prevent sensitive content leakage in brief/summary artifacts.

Apply these rules to task text and extracted memo fields before writing outputs.

## Redaction Patterns
1. `private_key_block`  
   Pattern: `-----BEGIN [^-]*PRIVATE KEY-----...-----END [^-]*PRIVATE KEY-----`
2. `openai_key`  
   Pattern: `sk-[A-Za-z0-9]{20,}`
3. `aws_access_key`  
   Pattern: `AKIA[0-9A-Z]{16}`
4. `jwt`  
   Pattern: `<header>.<payload>.<sig>` with long base64url-like parts
5. `bearer_token`  
   Pattern: `bearer <token>`
6. `credential_assignment`  
   Pattern: `(password|passwd|pwd|secret|api[_-]?key|token)\s*[:=]\s*<value>`
7. `connection_string`  
   Pattern: `(postgres|postgresql|mysql|mongodb|redis|amqp)://...`

## Replacement Rule
- Replace each match with `<REDACTED:<pattern_name>>`.

## Reporting Rule
- Track per-pattern count.
- Emit:
  - `redaction_count` (total)
  - `pattern_hits[]` entries with `{pattern, count}`

## Output Requirements
- Brief and summary must contain a `## Safety` section.
- Json payloads must expose redaction totals and pattern hit details.
