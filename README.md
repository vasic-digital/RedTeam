# digital.vasic.redteam

YAML-driven adversarial prompt fixture harness for **defensive** LLM
guardrail regression testing.

Each fixture is an adversarial prompt (jailbreak, filter-bypass,
stego-mutation, system-prompt extraction, role-reversal, abliteration
probe, genetic seed) bundled with the guardrail detector the consumer
is expected to trigger. Consumers replay the fixture set against their
guardrail pipeline and assert that every fixture is blocked.

This harness is **defensive**: it verifies that known attack patterns
get caught. It does NOT generate new attacks.

## Attack classes

Seven classes, one YAML file each under `fixtures/`:

| Class | File |
|-------|------|
| `jailbreak` | `fixtures/jailbreak.yaml` |
| `abliteration_probe` | `fixtures/abliteration_probe.yaml` |
| `filter_bypass` | `fixtures/filter_bypass.yaml` |
| `stego_mutation` | `fixtures/stego_mutation.yaml` |
| `genetic_seed` | `fixtures/genetic_seed.yaml` |
| `system_prompt_extraction` | `fixtures/system_prompt_extraction.yaml` |
| `role_reversal` | `fixtures/role_reversal.yaml` |

## Usage

```go
import "digital.vasic.redteam"

fixtures, err := redteam.LoadByClass(redteam.AttackClassJailbreak)
if err != nil {
    return err
}
for _, f := range fixtures {
    results, err := myGuardrailPipeline.CheckInput(ctx, f.Prompt, nil)
    if err != nil {
        t.Errorf("pipeline error on %s: %v", f.ID, err)
        continue
    }
    blocked := false
    for _, r := range results {
        if r != nil && r.Triggered && r.Action == ActionBlock {
            blocked = true
            break
        }
    }
    if !blocked {
        t.Errorf("fixture %s (expected %s) slipped through",
            f.ID, f.ExpectedGuardrailTrigger)
    }
}
```

`LoadAll()` returns every class keyed by `AttackClass`.

## Schema

See [`fixtures/README.md`](fixtures/README.md) for per-fixture field
semantics and the policy statement.

## Defensive-use policy

This module exists so that defenders can regression-test their
guardrails against documented attack patterns. It does not generate
new attacks, it does not produce obfuscated payloads from plain text,
and it does not expose helpers that would invert the direction.
Integrating this into any red-team or bypass tooling violates the
stated use case.

## Tests

```bash
GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -count=1 -p 1 -race ./...
```

## Primary consumer

HelixAgent (`dev.helix.agent`) — consumed via `go.mod` replace
directive pointing at the `RedTeam/` submodule.
`DeepTeamRedTeamer.RunFixtureSuite(ctx, class)` in
`internal/security/redteam_fixtures.go` wires the fixture set through
`StandardGuardrailPipeline` and asserts 47/47 blocked.

## License

Apache-2.0
