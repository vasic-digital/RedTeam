# Red-Team Fixtures — Defensive Use Only

These YAML fixtures are adversarial prompt samples used by consumers
of `digital.vasic.redteam` to verify that their guardrail pipelines
block the attack classes the fixtures describe.

## Policy

- **Defensive use only.** Never feed a fixture to a live model without
  passing it through a guardrail pipeline first. The primary consumer
  (HelixAgent) wraps every replay in `StandardGuardrailPipeline` and
  asserts `GuardrailActionBlock`.
- **No offensive repurposing.** The fixtures are documented attack
  patterns included here so that defenders can regression-test against
  them. Repurposing this module as an attack payload generator (for
  example, feeding fixtures straight to an unprotected model) violates
  the stated use case.

## Schema (per fixture)

| Field | Purpose |
|-------|---------|
| `id` | Stable identifier `redteam.<class>.<seq>` |
| `attack_class` | Attack category (optional per-fixture; defaults to the file-level `attack_class`) |
| `prompt` | Adversarial input (text) |
| `expected_guardrail_trigger` | Name of the guardrail detector that must flag this fixture |
| `severity` | `low` / `medium` / `high` |
| `source` | Upstream reference (for audit) |
| `notes` | Free-form commentary |

## Consumer

`fixtures.go` (package `redteam`) parses these files. Consumers call
`redteam.LoadByClass(class)` / `redteam.LoadAll()` and replay the
fixtures through their own guardrail pipeline. HelixAgent's
`DeepTeamRedTeamer.RunFixtureSuite(ctx, class)` is the reference
integration.
