# Branch Protection Required Checks

Recommended required checks for protected branches:

- `ci-required-jobs-green`
- `boilerplate-quality-report`

Notes:
- Keep required checks aligned with fast CI and governance proof.
- Do not add `live-stack-browser-smoke-review` as a required branch-protection check; `e2e:live` remains nightly or on-demand release evidence.
