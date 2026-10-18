package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/TuanKietTran/fraud-detection-go/internal/domain"
)

const MaxStartupRulesBytes = 8192

// The fixed error intentionally omits raw input, field names and decoder errors.
var ErrStartupRules = errors.New("invalid FRAUD_RULES_JSON: expected a bounded JSON object with canonical integer rule fields and valid thresholds")

func parseStartupRules(raw string) (domain.RuleConfig, error) {
	rules := domain.DefaultRuleConfig()
	if len(raw) > MaxStartupRulesBytes {
		return domain.RuleConfig{}, ErrStartupRules
	}
	if strings.TrimSpace(raw) == "" {
		return rules, nil
	}
	fields := map[string]any{
		"ReviewAt": &rules.ReviewAt, "DenyAt": &rules.DenyAt,
		"HighAmountCents": &rules.HighAmountCents, "VeryHighAmountCents": &rules.VeryHighAmountCents,
		"HighAmountWeight": &rules.HighAmountWeight, "VeryHighAmountWeight": &rules.VeryHighAmountWeight,
		"VelocityLimit": &rules.VelocityLimit, "VelocityWeight": &rules.VelocityWeight,
		"GeoMismatchWeight": &rules.GeoMismatchWeight, "DeviceMismatchWeight": &rules.DeviceMismatchWeight,
		"NewPayeeWeight": &rules.NewPayeeWeight, "StructuringWeight": &rules.StructuringWeight,
		"StructuringBandCents": &rules.StructuringBandCents,
	}
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.UseNumber()
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return domain.RuleConfig{}, ErrStartupRules
	}
	seen := make(map[string]bool)
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil {
			return domain.RuleConfig{}, ErrStartupRules
		}
		name, ok := key.(string)
		field, known := fields[name]
		if !ok || !known || seen[name] {
			return domain.RuleConfig{}, ErrStartupRules
		}
		seen[name] = true
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return domain.RuleConfig{}, ErrStartupRules
		}
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return domain.RuleConfig{}, ErrStartupRules
		}
		if err := json.Unmarshal(value, field); err != nil {
			return domain.RuleConfig{}, ErrStartupRules
		}
	}
	if token, err := decoder.Token(); err != nil || token != json.Delim('}') {
		return domain.RuleConfig{}, ErrStartupRules
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return domain.RuleConfig{}, ErrStartupRules
	}
	if err := rules.Validate(); err != nil {
		return domain.RuleConfig{}, ErrStartupRules
	}
	return rules, nil
}
