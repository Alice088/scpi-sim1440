# TEST_RESULTS

## Формат ведения

Каждый прогон добавляется сверху, сразу под этим разделом.
Заголовок: `## Прогон: YYYY-MM-DD HH:MM TZ (±HHMM)`.
Результаты — только таблицей, затем блоки `### Сводка` и `### Падения`.

---

## Прогон: 2026-09-19 15:36 MSK (+0300)

Команда: `go test ./... -count=1`, `go vet ./...`, e2e `/tmp/scpi-manager` (down/down/up/up/`work`/down, nc-проверка)
Ветка: `master`
Go: `go1.27.1 linux/amd64`
OR-Tools: не используется

| # | Тест | Пакет | Результат | Время |
|---|------|-------|-----------|-------|
| 1 | TestPlanCombinesDelayAndGarbageInDeviceChain | scpi-sim1440/internal/noise | PASS | 0.00s |
| 2 | TestPlanEmptyChainsAtZeroChance | scpi-sim1440/internal/noise | PASS | 0.00s |
| 3 | go build ./... | scpi-sim1440 | PASS | 0.90s |
| 4 | go vet ./... | scpi-sim1440 | PASS | 0.20s |
| 5 | e2e down повторно: errdefs.IsNotFound игнорируется | cmd/manager | PASS | 0.20s |
| 6 | e2e up после down: сеть + bat-1, IP == 172.30.0.11 | cmd/manager | PASS | 1.00s |
| 7 | e2e up идемпотентен (already running, skip) | cmd/manager | PASS | 0.20s |
| 8 | e2e nc MEAS:VOLT? -> ответ от bat-1 | cmd/manager | PASS | 0.40s |
| 9 | e2e work: busybox, подсказка nc <ip> <port> внутри | cmd/manager | PASS | 0.60s |
| 10 | e2e down: контейнер и сеть удалены | cmd/manager | PASS | 0.20s |

### Сводка

Всего: 10 / PASS: 10 / FAIL: 0 / SKIP: 0
Время: 3.70s

### Падения

| Тест | Причина |
|------|---------|
| нет | нет |

---

## Прогон: 2026-09-19 15:22 MSK (+0300)

Команда: `go test ./... -count=1`, `go vet ./...`, e2e `/tmp/scpi-manager` (up/up/down/down/up, ручной `docker stop`)
Ветка: `master`
Go: `go1.27.1 linux/amd64`
OR-Tools: не используется

| # | Тест | Пакет | Результат | Время |
|---|------|-------|-----------|-------|
| 1 | TestPlanCombinesDelayAndGarbageInDeviceChain | scpi-sim1440/internal/noise | PASS | 0.00s |
| 2 | TestPlanEmptyChainsAtZeroChance | scpi-sim1440/internal/noise | PASS | 0.00s |
| 3 | go build ./... | scpi-sim1440 | PASS | 0.90s |
| 4 | go vet ./... | scpi-sim1440 | PASS | 0.20s |
| 5 | e2e up: AutoRemove=true, stand IP == 172.30.0.11 | cmd/manager | PASS | 1.00s |
| 6 | e2e up идемпотентен (already running, skip) | cmd/manager | PASS | 0.20s |
| 7 | e2e down: контейнер и сеть stand удалены | cmd/manager | PASS | 0.20s |
| 8 | e2e down повторно (not found игнорируется) | cmd/manager | PASS | 0.10s |
| 9 | e2e up после down: пересоздание с .11 | cmd/manager | PASS | 1.00s |
| 10 | e2e ручной docker stop -> авто-удаление контейнера | cmd/manager | PASS | 1.10s |

### Сводка

Всего: 10 / PASS: 10 / FAIL: 0 / SKIP: 0
Время: 4.90s

### Падения

| Тест | Причина |
|------|---------|
| нет | нет |

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
