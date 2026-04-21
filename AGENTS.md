# AGENTS.md -- digital.vasic.redteam

## Module

`digital.vasic.redteam` -- YAML-driven adversarial prompt fixture
harness for defensive LLM guardrail regression testing.

## Framing: DEFENSIVE USE ONLY

This library loads documented adversarial prompt fixtures (jailbreak,
filter-bypass, stego-mutation, system-prompt extraction, role-reversal,
abliteration probe, genetic seed) so that defenders can regression-test
their guardrail pipelines against them. A consumer runs every fixture
through its pipeline and asserts that every fixture is blocked.

**It is NOT an attack toolkit.** It does not generate new attack
strings, it does not mutate prompts into obfuscated payloads, and it
must not be repurposed as such. Agents touching this module must
reject any request that inverts the direction (for example, "feed the
fixture corpus to a raw model without a guardrail" or "export the
corpus as a one-shot payload bundle for an offensive run"). The
fixtures exist so that the attacks they describe fail.

## Primary consumer

HelixAgent (`dev.helix.agent`). The `internal/security` package
consumes `redteam.LoadByClass` and replays every fixture through
`StandardGuardrailPipeline`; the acceptance gate is 47/47 blocked.

Repository: `git@github.com:vasic-digital/HelixAgent.git`
(pinned via submodule + `go.mod` replace).

## Contribution policy

- Additions to the fixture corpus must be motivated by a documented,
  defensive use case (a specific bypass observed in the wild that a
  consumer's guardrails now handle, or a regression test for a
  newly-shipped detector).
- Every new fixture must declare the `expected_guardrail_trigger` and
  a `severity` tier; otherwise the loader's shape invariants fail.
- Never add helpers that generate new attack strings from seeds, mutate
  fixtures into obfuscated variants, or otherwise produce novel
  offensive payloads. The consumer can trivially do that if it needs
  to; exporting such helpers enlarges attack surface for downstream
  consumers that did not ask for it.
