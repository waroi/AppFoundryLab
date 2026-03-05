# Frontend Notes

## 1. Scope

`frontend/` is the Astro shell plus Svelte islands that expose the local diagnostics UI, runtime incident surfaces, and the restore-drill preview page.

## 2. File map

- `src/layouts/BaseLayout.astro`: document shell, pre-paint theme bootstrap, shared toolbar mount
- `src/lib/ui/preferences.ts`: normalized locale/theme store, document sync, and theme persistence
- `src/lib/ui/copy.ts`: EN/TR copy dictionary plus locale-aware formatting helpers
- `src/lib/ui/routes.ts`: canonical mapping between logical pages and localized URLs
- `src/components/Page/`: shared Astro page shells reused by both default and localized routes
- `src/components/Layout/PreferenceToolbar.svelte`: shared language/theme controls
- `src/components/Static/`: shell copy that must react to locale changes
- `src/components/Interactive/`: diagnostics and restore-drill surfaces
- `src/styles/global.scss`: semantic surface/text/control tokens and theme-specific values

## 3. Locale and theme contract

- Supported locales: `en`, `tr`
- Supported themes: `light`, `dark`
- Route contract: `/` and `/test` resolve to `en`, `/tr` and `/tr/test` resolve to `tr`
- Persistence key: `appfoundrylab.theme`
- `BaseLayout.astro` writes route-correct `html[lang]` and `html[data-theme]`, then an inline bootstrap script reconciles only the saved theme before the main UI hydrates
- Locale is URL-authoritative on entry; toolbar language switches navigate to the localized route variant for the current logical page
- `src/styles/global.scss` keeps the light theme airy while mapping dark theme surfaces to charcoal and CTA accents to vivid orange tokens
- Interactive components must read the shared store instead of keeping their own locale/theme state

## 4. Contributor rules

- Do not hardcode new user-facing strings in components; add them to `src/lib/ui/copy.ts`
- Do not invent ad-hoc locale paths; route changes must go through `src/lib/ui/routes.ts`
- Prefer semantic shared classes from `src/styles/global.scss` over one-off light-only utility colors
- If UI state matters to tests, add stable `data-testid` or `data-*` attributes instead of relying on visible translated text
- When adding a new page title, route-level shell string, or dynamic formatter, update both locales in the same change set

## 5. Validation

```bash
cd frontend
./node_modules/.bin/astro check
./node_modules/.bin/astro build
node ./scripts/smoke.mjs
./scripts/run-playwright.sh
```
