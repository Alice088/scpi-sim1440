# TEST_RESULTS

## Формат ведения

Каждый прогон добавляется сверху, сразу под этим разделом.
Заголовок: `## Прогон: YYYY-MM-DD HH:MM TZ (±HHMM)`.
Результаты — только таблицей, затем блоки `### Сводка` и `### Падения`.

---

## Прогон: 2026-09-19 15:04 MSK (+0300)

Команда: `go test ./internal/noise/ -v -count=1`, `go vet ./...`, e2e `/tmp/scpi-manager -action up` + `-action down`
Ветка: `master`
Go: `go1.27.1 linux/amd64`
OR-Tools: не используется

| # | Тест | Пакет | Результат | Время |
|---|------|-------|-----------|-------|
| 1 | TestPlanCombinesDelayAndGarbageInDeviceChain | scpi-sim1440/internal/noise | PASS | 0.00s |
| 2 | TestPlanEmptyChainsAtZeroChance | scpi-sim1440/internal/noise | PASS | 0.00s |
| 3 | go vet ./... | scpi-sim1440 | PASS | 0.20s |
| 4 | e2e up (bat-1, AutoRemove=true, tcp up, без parse error) | cmd/manager | PASS | 1.10s |
| 5 | e2e down (контейнер и сеть stand удалены) | cmd/manager | PASS | 0.30s |

### Сводка

Всего: 5 / PASS: 5 / FAIL: 0 / SKIP: 0
Время: 1.60s

### Падения

| Тест | Причина |
|------|---------|
| нет | нет |

---

## Прогон: 2026-09-19 15:03 MSK (+0300)

Команда: `go test ./internal/noise/ -v -count=1`
Ветка: `master`
Go: `go1.27.1 linux/amd64`
OR-Tools: не используется

| # | Тест | Пакет | Результат | Время |
|---|------|-------|-----------|-------|
| 1 | TestPlanCombinesDelayAndGarbageInDeviceChain | scpi-sim1440/internal/noise | PASS | 0.00s |
| 2 | TestPlanEmptyChainsAtZeroChance | scpi-sim1440/internal/noise | PASS | 0.00s |

### Сводка

Всего: 2 / PASS: 2 / FAIL: 0 / SKIP: 0
Время: 0.001s

### Падения

| Тест | Причина |
|------|---------|
| нет | нет |
