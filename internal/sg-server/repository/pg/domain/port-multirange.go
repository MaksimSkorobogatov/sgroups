package domain

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

// MarshalJSON serializes PortMultirange to PostgreSQL multirange text format.
// Example output: "{[80,81),[443,444),[8080,9091)}"
func (pm PortMultirange) MarshalJSON() ([]byte, error) {
	var sb strings.Builder
	sb.WriteByte('{')
	for i, r := range pm.Multirange {
		if i > 0 {
			sb.WriteByte(',')
		}
		if !r.Valid {
			sb.WriteString("empty")
			continue
		}
		switch r.LowerType {
		case pgtype.Inclusive:
			sb.WriteByte('[')
		case pgtype.Exclusive:
			sb.WriteByte('(')
		}
		fmt.Fprintf(&sb, "%d,%d", r.Lower, r.Upper)
		switch r.UpperType {
		case pgtype.Inclusive:
			sb.WriteByte(']')
		case pgtype.Exclusive:
			sb.WriteByte(')')
		}
	}
	sb.WriteByte('}')
	return json.Marshal(sb.String())
}

// UnmarshalJSON parses PostgreSQL multirange text format.
// Example input: "{[80,81),[443,444),[8080,9091)}"
func (pm *PortMultirange) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	s = strings.TrimSpace(s)
	if s == "" {
		pm.Multirange = nil
		return nil
	}

	// Remove outer braces: {[80,81),[443,444)} → [80,81),[443,444)
	s = strings.TrimPrefix(s, "{")
	s = strings.TrimSuffix(s, "}")
	if s == "" {
		pm.Multirange = nil
		return nil
	}

	pm.Multirange = nil
	for _, part := range splitRanges(s) {
		r, err := parseRange(part)
		if err != nil {
			return fmt.Errorf("parse port range %q: %w", part, err)
		}
		pm.Multirange = append(pm.Multirange, r)
	}
	return nil
}

// splitRanges splits "[80,81),[443,444)" into ["[80,81)", "[443,444)"].
func splitRanges(s string) []string {
	var result []string
	depth := 0
	start := 0
	for i, ch := range s {
		switch ch {
		case '[', '(':
			if depth == 0 {
				start = i
			}
			depth++
		case ']', ')':
			depth--
			if depth == 0 {
				result = append(result, s[start:i+1])
			}
		}
	}
	return result
}

// parseRange parses "[80,81)" into a PortRange.
func parseRange(s string) (PortRange, error) {
	var r PortRange

	if len(s) < 3 {
		return r, fmt.Errorf("range too short: %q", s)
	}

	switch s[0] {
	case '[':
		r.LowerType = pgtype.Inclusive
	case '(':
		r.LowerType = pgtype.Exclusive
	default:
		return r, fmt.Errorf("unexpected lower bound char: %c", s[0])
	}

	switch s[len(s)-1] {
	case ']':
		r.UpperType = pgtype.Inclusive
	case ')':
		r.UpperType = pgtype.Exclusive
	default:
		return r, fmt.Errorf("unexpected upper bound char: %c", s[len(s)-1])
	}

	inner := s[1 : len(s)-1]
	parts := strings.SplitN(inner, ",", 2)
	if len(parts) != 2 {
		return r, fmt.Errorf("expected 'lower,upper', got %q", inner)
	}

	lower, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 32)
	if err != nil {
		return r, fmt.Errorf("parse lower bound: %w", err)
	}

	upper, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 32)
	if err != nil {
		return r, fmt.Errorf("parse upper bound: %w", err)
	}

	r.Lower = PortNumber(lower)
	r.Upper = PortNumber(upper)
	r.Valid = true
	return r, nil
}
