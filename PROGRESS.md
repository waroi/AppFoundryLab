# PROGRESS

Bu dosya repo-owned backlog'un tek kanonik kaynagidir.
Stratejik faz sirasi [docs/gelistirmePlanı.md](/mnt/d/w/AppFoundryLab/docs/gelistirmePlanı.md) icinde tutulur; aktif faz satiri bu dosya ile birebir ayni kalmalidir.

Durum: 2026-03-06 itibariyla bu dosyada izlenen son uc faz tamamlandi.
Aktif faz: Yok - repo-owned acik faz kalmadi

## Son kapanan fazlar

### Faz 1 - Runtime Knob Transparency

- `GET /api/v1/admin/runtime-config` ve `runtime-report` artik request logging trusted proxy CIDR'larini ve logger timing knob'larini yayinlar.
- Admin diagnostics paneli runtime knob kartlarini tarayicida gosterir.
- Operator dokumanlari runtime-config ve admin diagnostics ile ayni dilde guncellendi.

### Faz 2 - Browser Coverage Depth

- Mock-backed Playwright regresyonlari keyboard/focus akisi, degraded admin diagnostics ve runtime knob fallback gorunumunu kapsar.
- Live-stack browser smoke runtime knob panelini assert eder.
- Governance script regresyonlari duplicate `# PROGRESS` heading ve release-policy drift fixture'larini kapsar.

### Faz 3 - Live Smoke Cost Governance

- `e2e:live` tam Docker-backed admin smoke olarak nightly ve on-demand release-confidence lane'i kabul edildi.
- `RUN_LIVE_STACK_BROWSER_SMOKE=true` olmadikca PR ve merge kapilarinda kosmaz.
- Policy, checklist ve workflow matrix'leri branch-protection disinda ama release evidence icinde tek karara sabitlendi.

## Governance notlari

- `check-doc-drift.sh --mode strict` artik `PROGRESS.md` icinde tam olarak bir adet `# PROGRESS` heading ister.
- `docs/gelistirmePlanı.md` icindeki `Sonraki aktif hedef:` satiri bu dosyadaki `Aktif faz:` degeriyle birebir ayni kalmalidir.
- Yeni repo backlog'u ancak taze analiz ve test kaniti ile yeniden acilmalidir.
