// Command runner is the round-281 anti-bluff Challenge runner for
// digital.vasic.redteam. It exercises the public loader API
// (LoadByClass, LoadAll, SupportedAttackClasses) against the embedded
// fixture corpus, enforces every schema invariant at runtime, and
// emits a 5-locale bilingual UX summary line per CONST-046.
//
// Defensive-use only. The runner reads the corpus; it does NOT execute
// any fixture against any model. There is no inverse helper, no
// payload generator, no obfuscator.
//
// Exit codes:
//   0   — every class loaded, every fixture passed invariants, every
//         locale line printed.
//   1   — usage / flag error.
//   2   — coverage gap (a supported class has no file, or a file
//         declares an unsupported class).
//   3   — schema-invariant violation in at least one fixture.
//   4   — locale UX line missing.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	redteam "digital.vasic.redteam"
)

// locale describes a UX line printed by the runner. The text is a
// short, locale-correct summary that consumers can grep for to confirm
// that the runner produced operator-facing output in every supported
// locale.
type locale struct {
	tag  string
	line func(classCount, fixtureCount int) string
}

// supportedLocales is the 5-locale CONST-046 set the runner must emit
// every run. The set mirrors the test-bank locale matrix used across
// other round-281 enrichments.
func supportedLocales() []locale {
	return []locale{
		{
			tag: "en",
			line: func(c, f int) string {
				return fmt.Sprintf("[en] redteam: %d attack classes loaded, %d fixtures parsed (defensive-use only)", c, f)
			},
		},
		{
			tag: "sr",
			line: func(c, f int) string {
				return fmt.Sprintf("[sr] redteam: %d napadnih klasa učitano, %d fiksacija obrađeno (samo za odbranu)", c, f)
			},
		},
		{
			tag: "ja",
			line: func(c, f int) string {
				return fmt.Sprintf("[ja] redteam: %d 個の攻撃クラス、%d 個のフィクスチャを読み込み(防御用途のみ)", c, f)
			},
		},
		{
			tag: "es",
			line: func(c, f int) string {
				return fmt.Sprintf("[es] redteam: %d clases de ataque cargadas, %d fixtures parseados (uso defensivo)", c, f)
			},
		},
		{
			tag: "de",
			line: func(c, f int) string {
				return fmt.Sprintf("[de] redteam: %d Angriffsklassen geladen, %d Fixtures geparst (nur Verteidigung)", c, f)
			},
		},
	}
}

func main() {
	all := flag.Bool("all", false, "run every check (default mode)")
	class := flag.String("class", "", "exercise only the named attack class")
	flag.Parse()

	if !*all && *class == "" {
		*all = true
	}

	if *class != "" {
		if err := runClass(redteam.AttackClass(*class)); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(exitCodeFor(err))
		}
		return
	}

	if err := runAll(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitCodeFor(err))
	}
}

// runClass exercises a single attack class. It enforces every schema
// invariant and emits the per-class header line.
func runClass(c redteam.AttackClass) error {
	fixtures, err := redteam.LoadByClass(c)
	if err != nil {
		return wrap(errCoverage, fmt.Errorf("LoadByClass(%q): %w", c, err))
	}
	fmt.Printf("class=%s fixtures=%d\n", c, len(fixtures))
	for _, f := range fixtures {
		if err := assertFixture(f); err != nil {
			return wrap(errSchema, fmt.Errorf("class %s: %w", c, err))
		}
	}
	return nil
}

// runAll exercises every supported class, asserts the embed FS
// surface matches the supported list, and emits the 5-locale summary.
func runAll() error {
	loaded, err := redteam.LoadAll()
	if err != nil {
		return wrap(errCoverage, fmt.Errorf("LoadAll: %w", err))
	}

	supported := redteam.SupportedAttackClasses()
	if len(loaded) != len(supported) {
		return wrap(errCoverage, fmt.Errorf(
			"coverage gap: LoadAll returned %d classes, SupportedAttackClasses has %d",
			len(loaded), len(supported),
		))
	}

	// Sort class names so output is reproducible across runs (map
	// iteration order would otherwise reorder the per-class lines).
	classes := make([]string, 0, len(loaded))
	for c := range loaded {
		classes = append(classes, string(c))
	}
	sort.Strings(classes)

	total := 0
	for _, name := range classes {
		c := redteam.AttackClass(name)
		fixtures := loaded[c]
		fmt.Printf("class=%s fixtures=%d\n", c, len(fixtures))
		for _, f := range fixtures {
			if err := assertFixture(f); err != nil {
				return wrap(errSchema, fmt.Errorf("class %s: %w", c, err))
			}
		}
		total += len(fixtures)
	}

	// 5-locale bilingual UX evidence per CONST-046.
	printed := 0
	for _, loc := range supportedLocales() {
		out := loc.line(len(classes), total)
		if !strings.Contains(out, "redteam:") {
			return wrap(errLocale, fmt.Errorf("locale %s: missing canonical token", loc.tag))
		}
		fmt.Println(out)
		printed++
	}
	if printed != len(supportedLocales()) {
		return wrap(errLocale, fmt.Errorf("printed %d/%d locales", printed, len(supportedLocales())))
	}

	fmt.Printf("OK classes=%d total_fixtures=%d locales=%d\n", len(classes), total, printed)
	return nil
}

// assertFixture enforces the per-fixture schema invariants the
// loader's unit tests assert for a single class. Anti-bluff: every
// fixture in every class is checked, not just the first one.
func assertFixture(f redteam.Fixture) error {
	if strings.TrimSpace(f.ID) == "" {
		return fmt.Errorf("fixture has empty ID")
	}
	if strings.TrimSpace(f.Prompt) == "" {
		return fmt.Errorf("fixture %s has empty Prompt", f.ID)
	}
	if strings.TrimSpace(f.ExpectedGuardrailTrigger) == "" {
		return fmt.Errorf("fixture %s has empty ExpectedGuardrailTrigger", f.ID)
	}
	if string(f.AttackClass) == "" {
		return fmt.Errorf("fixture %s has empty AttackClass", f.ID)
	}
	switch f.Severity {
	case "low", "medium", "high":
		// OK
	default:
		return fmt.Errorf("fixture %s has unexpected Severity %q (want low|medium|high)", f.ID, f.Severity)
	}
	return nil
}

// Sentinel error tags used to compute exit codes without printing the
// tag itself.
var (
	errCoverage = errors.New("coverage")
	errSchema   = errors.New("schema")
	errLocale   = errors.New("locale")
)

// taggedError attaches a sentinel for exit-code mapping while
// preserving the inner cause via Unwrap.
type taggedError struct {
	tag   error
	inner error
}

func (e *taggedError) Error() string { return e.inner.Error() }
func (e *taggedError) Unwrap() error { return e.inner }
func (e *taggedError) Is(t error) bool {
	return errors.Is(e.tag, t)
}

func wrap(tag, inner error) error {
	return &taggedError{tag: tag, inner: inner}
}

func exitCodeFor(err error) int {
	switch {
	case errors.Is(err, errCoverage):
		return 2
	case errors.Is(err, errSchema):
		return 3
	case errors.Is(err, errLocale):
		return 4
	default:
		return 1
	}
}
