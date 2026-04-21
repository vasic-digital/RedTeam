// Package redteam provides a YAML-driven adversarial prompt fixture
// harness for defensive LLM guardrail regression testing.
//
// Each fixture is an adversarial prompt with an expected guardrail
// trigger. Consumers replay the fixture set against their guardrail
// pipeline and assert that every fixture is blocked. The harness is
// DEFENSIVE — it does not generate new attacks, only verifies that
// known attack patterns get caught.
package redteam
