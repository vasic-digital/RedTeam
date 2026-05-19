# test-coverage.md — digital.vasic.redteam

Round 281 symbol → test / Challenge ledger. Every exported symbol of
`digital.vasic.redteam` MUST appear here with the test(s) and
Challenge(s) that exercise it AND the anti-bluff dimension each
proves. Adding an exported symbol without updating this ledger is a
CONST-048 violation. Per Article XI §11.9, every PASS row MUST carry
positive runtime evidence — the "Evidence" column documents what to
capture during a release-gate sweep.

## Exported symbols

| Symbol                                | Kind        | Unit test(s)                                  | Challenge(s)                       | Anti-bluff dimension                                                                 | Evidence (runtime)                                |
|---------------------------------------|-------------|-----------------------------------------------|------------------------------------|--------------------------------------------------------------------------------------|---------------------------------------------------|
| `AttackClass`                         | type        | `TestLoadByClass_Jailbreak_LoadsSuccessfully` | `runner -all`                      | Type identity flows from YAML through loader to consumer.                            | `go test -v` line `PASS`.                         |
| `AttackClassJailbreak`                | const       | `TestLoadByClass_Jailbreak_LoadsSuccessfully` | `runner -all`                      | Constant resolves to a file present on the embed FS.                                 | Loader returns non-empty `[]Fixture`.             |
| `AttackClassAbliterationProbe`        | const       | `TestLoadAll_LoadsEverySupportedClass`        | `runner -all`                      | Constant resolves to a file present on the embed FS.                                 | `LoadAll()` key set contains the constant.        |
| `AttackClassFilterBypass`             | const       | `TestLoadAll_LoadsEverySupportedClass`        | `runner -all`                      | Constant resolves to a file present on the embed FS.                                 | `LoadAll()` key set contains the constant.        |
| `AttackClassStegoMutation`            | const       | `TestLoadAll_LoadsEverySupportedClass`        | `runner -all`                      | Constant resolves to a file present on the embed FS.                                 | `LoadAll()` key set contains the constant.        |
| `AttackClassGeneticSeed`              | const       | `TestLoadAll_LoadsEverySupportedClass`        | `runner -all`                      | Constant resolves to a file present on the embed FS.                                 | `LoadAll()` key set contains the constant.        |
| `AttackClassSystemPromptExtraction`   | const       | `TestLoadAll_LoadsEverySupportedClass`        | `runner -all`                      | Constant resolves to a file present on the embed FS.                                 | `LoadAll()` key set contains the constant.        |
| `AttackClassRoleReversal`             | const       | `TestLoadAll_LoadsEverySupportedClass`        | `runner -all`                      | Constant resolves to a file present on the embed FS.                                 | `LoadAll()` key set contains the constant.        |
| `SupportedAttackClasses`              | func        | `TestLoadAll_LoadsEverySupportedClass`        | `runner -all`                      | Source-of-truth list matches the embed-FS surface.                                   | Coverage assertion in runner.                     |
| `Fixture`                             | struct      | `TestLoadByClass_Jailbreak_LoadsSuccessfully` | `runner -all`                      | Schema invariants (non-empty ID, valid Severity) enforced at parse time.             | Per-fixture assertion lines.                      |
| `LoadByClass`                         | func        | `TestLoadByClass_Jailbreak_LoadsSuccessfully`, `TestLoadByClass_UnknownClass_ReturnsError` | `runner -all`, `runner -class=<C>` | Real YAML parse on embed FS; unknown class returns explicit error (not silent miss). | `go test` PASS + Challenge stdout `class=<C> fixtures=<N>`. |
| `LoadAll`                             | func        | `TestLoadAll_LoadsEverySupportedClass`        | `runner -all`                      | Every supported class has a file; no orphan files exist.                             | Challenge stdout `classes=7`.                     |

## Anti-bluff dimensions covered

| Dimension                                                           | Where proved                                       |
|---------------------------------------------------------------------|----------------------------------------------------|
| Real I/O (embed FS, not hardcoded list)                             | `LoadAll` iterates `fixtureFS.ReadDir`             |
| Real YAML parse (not regex)                                         | `yaml.Unmarshal` in `LoadByClass`                  |
| Schema invariants enforced (not metadata-only)                      | Runner asserts ID/Severity/Prompt/Trigger non-empty|
| 5-locale bilingual UX (CONST-046)                                   | Runner prints `en/sr/ja/es/de` summary lines       |
| Paired-mutation evidence (CONST-035)                                | `--mutate` flag in describe-Challenge              |
| Defensive-use boundary preserved (no inverse helpers)               | Grep gate in describe-Challenge                    |
| Error path covered (unknown class fails loudly, not silently)       | `TestLoadByClass_UnknownClass_ReturnsError`        |

## Maintenance

Every CL that touches `fixtures.go` (adds/removes/renames an exported
symbol, alters `Fixture` shape, changes loader behaviour) MUST update
this file in the SAME commit. Drift is a CONST-048 violation. The
Challenge runner walks the exported-symbol set at runtime via the
runner's reflection-free assertion plan — adding a symbol without
extending the runner asserts is a paired CONST-035 violation.
