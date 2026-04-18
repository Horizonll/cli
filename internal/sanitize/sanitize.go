package sanitize

import (
	"regexp"
	"strings"
)

type Level string

const (
	LevelStrict   Level = "strict"
	LevelBalanced Level = "balanced"
	LevelOff      Level = "off"
)

type Rule struct {
	Name        string
	Pattern     *regexp.Regexp
	Replacement string
}

type RuleResult struct {
	Name         string `json:"name"`
	Replacements int    `json:"replacements"`
}

type Report struct {
	Level              Level        `json:"level"`
	Rules              []RuleResult `json:"rules"`
	SuspiciousUnmasked bool         `json:"suspicious_unmasked"`
}

type Sanitizer struct {
	level Level
	rules []Rule
}

func New(level Level) Sanitizer {
	if level == "" {
		level = LevelBalanced
	}
	return Sanitizer{
		level: level,
		rules: []Rule{
			{Name: "github_token", Pattern: regexp.MustCompile(`(?i)\b(ghp|github_pat)_[A-Za-z0-9_]+\b`), Replacement: "[REDACTED_GITHUB_TOKEN]"},
			{Name: "bearer_auth", Pattern: regexp.MustCompile(`(?i)(authorization\s*:\s*bearer\s+)[^\s]+`), Replacement: "$1[REDACTED_BEARER]"},
			{Name: "aws_access_key", Pattern: regexp.MustCompile(`\b(AKIA|ASIA)[A-Z0-9]{16}\b`), Replacement: "[REDACTED_AWS_ACCESS_KEY]"},
			{Name: "private_key_block", Pattern: regexp.MustCompile(`(?s)-----BEGIN [^-]*PRIVATE KEY-----.*?-----END [^-]*PRIVATE KEY-----`), Replacement: "[REDACTED_PRIVATE_KEY_BLOCK]"},
			{Name: "kv_secret", Pattern: regexp.MustCompile(`(?i)\b(token|password|apikey|api_key|secret)\s*[:=]\s*([^\s"']+)`), Replacement: "$1=[REDACTED]"},
		},
	}
}

func ParseLevel(v string) Level {
	v = strings.ToLower(strings.TrimSpace(v))
	switch Level(v) {
	case LevelStrict, LevelBalanced, LevelOff:
		return Level(v)
	default:
		return LevelBalanced
	}
}

func (s Sanitizer) Apply(input string) (string, Report) {
	report := Report{Level: s.level, Rules: make([]RuleResult, 0, len(s.rules))}
	if s.level == LevelOff {
		return input, report
	}
	out := input
	for _, rule := range s.rules {
		count := len(rule.Pattern.FindAllStringIndex(out, -1))
		if count > 0 {
			out = rule.Pattern.ReplaceAllString(out, rule.Replacement)
		}
		report.Rules = append(report.Rules, RuleResult{Name: rule.Name, Replacements: count})
	}
	if s.level == LevelStrict {
		jwtLike := regexp.MustCompile(`\beyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9._-]{10,}\b`)
		if jwtLike.MatchString(out) {
			report.SuspiciousUnmasked = true
			out = jwtLike.ReplaceAllString(out, "[REDACTED_JWT]")
		}
	}
	return out, report
}

func MergeReports(reports ...Report) Report {
	final := Report{Level: LevelBalanced}
	counts := map[string]int{}
	for _, rep := range reports {
		if rep.Level != "" {
			final.Level = rep.Level
		}
		if rep.SuspiciousUnmasked {
			final.SuspiciousUnmasked = true
		}
		for _, rr := range rep.Rules {
			counts[rr.Name] += rr.Replacements
		}
	}
	for name, count := range counts {
		final.Rules = append(final.Rules, RuleResult{Name: name, Replacements: count})
	}
	return final
}
