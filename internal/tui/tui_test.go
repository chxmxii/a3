package tui

import (
	"testing"

	"github.com/chxmxii/a3/internal/storage"
)

func TestParseSGRulesOpenCIDR(t *testing.T) {
	perms := []any{
		map[string]any{
			"ip_protocol": "tcp",
			"from_port":   float64(443),
			"to_port":     float64(443),
			"ip_ranges": []any{
				map[string]any{"cidr_ip": "0.0.0.0/0", "description": "world"},
				map[string]any{"cidr_ip": "10.0.0.0/8"},
			},
		},
	}

	for _, egress := range []bool{false, true} {
		rules := parseSGRules(perms, egress)
		if len(rules) != 2 {
			t.Fatalf("egress=%v: got %d rules, want 2", egress, len(rules))
		}
		if !rules[0].isOpen {
			t.Errorf("egress=%v: 0.0.0.0/0 rule not marked open", egress)
		}
		if !rules[0].isCIDR {
			t.Errorf("egress=%v: 0.0.0.0/0 rule not marked as CIDR", egress)
		}
		if rules[0].target != "0.0.0.0/0" || rules[0].desc != "world" {
			t.Errorf("egress=%v: got target=%q desc=%q", egress, rules[0].target, rules[0].desc)
		}
		if rules[0].portRange != "443" || rules[0].proto != "tcp" {
			t.Errorf("egress=%v: got proto=%q ports=%q, want tcp/443", egress, rules[0].proto, rules[0].portRange)
		}
		if rules[1].isOpen {
			t.Errorf("egress=%v: 10.0.0.0/8 rule wrongly marked open", egress)
		}
	}
}

func TestParseSGRulesOpenIPv6(t *testing.T) {
	perms := []any{
		map[string]any{
			"ip_protocol": "-1",
			"ip_ranges":   []any{map[string]any{"cidr_ip": "::/0"}},
		},
	}
	rules := parseSGRules(perms, true)
	if len(rules) != 1 {
		t.Fatalf("got %d rules, want 1", len(rules))
	}
	if !rules[0].isOpen {
		t.Error("::/0 egress rule not marked open")
	}
	if rules[0].proto != "ALL" || rules[0].portRange != "ALL" {
		t.Errorf("got proto=%q ports=%q, want ALL/ALL", rules[0].proto, rules[0].portRange)
	}
}

func TestParseSGRulesFallbackAndSGRefs(t *testing.T) {
	perms := []any{
		map[string]any{
			"ip_protocol":         "tcp",
			"from_port":           float64(80),
			"to_port":             float64(80),
			"user_id_group_pairs": []any{map[string]any{"group_id": "sg-123"}},
		},
	}

	// Ingress renders the SG reference.
	rules := parseSGRules(perms, false)
	if len(rules) != 1 {
		t.Fatalf("ingress: got %d rules, want 1", len(rules))
	}
	if rules[0].target != "sg:sg-123" || rules[0].isFallback {
		t.Errorf("ingress: got target=%q isFallback=%v, want sg:sg-123/false", rules[0].target, rules[0].isFallback)
	}

	// Egress skips SG references and falls back to the "(all)" placeholder.
	rules = parseSGRules(perms, true)
	if len(rules) != 1 {
		t.Fatalf("egress: got %d rules, want 1", len(rules))
	}
	if rules[0].target != "(all)" || !rules[0].isFallback {
		t.Errorf("egress: got target=%q isFallback=%v, want (all)/true", rules[0].target, rules[0].isFallback)
	}

	// An ingress rule with no ranges or refs uses "(self/all)".
	rules = parseSGRules([]any{map[string]any{"ip_protocol": "tcp"}}, false)
	if len(rules) != 1 || rules[0].target != "(self/all)" || !rules[0].isFallback {
		t.Errorf("got %+v, want a single (self/all) fallback rule", rules)
	}
}

func TestResourceMatchesSearch(t *testing.T) {
	r := storage.Resource{
		ResourceID:   "i-0abc123def",
		ResourceType: "ec2_instance",
		Name:         "Web-Server",
		Region:       "us-east-1",
	}

	for _, query := range []string{"web", "SERVER", "0ABC", "ec2_inst", "east", "i-0abc123def"} {
		if !resourceMatchesSearch(r, query) {
			t.Errorf("query %q should match %+v", query, r)
		}
	}
	for _, query := range []string{"zzz", "west", "rds"} {
		if resourceMatchesSearch(r, query) {
			t.Errorf("query %q should not match %+v", query, r)
		}
	}
}

func TestClampScroll(t *testing.T) {
	tests := []struct {
		name    string
		total   int
		offset  int
		maxRows int
		want    int
	}{
		{"beyond max", 10, 20, 5, 5},
		{"at max", 10, 5, 5, 5},
		{"within range", 10, 3, 5, 3},
		{"negative", 10, -3, 5, 0},
		{"content fits", 3, 2, 5, 0},
		{"empty", 0, 4, 5, 0},
	}
	for _, tt := range tests {
		offset := tt.offset
		clampScroll(tt.total, &offset, tt.maxRows)
		if offset != tt.want {
			t.Errorf("%s: clampScroll(%d, %d, %d) = %d, want %d", tt.name, tt.total, tt.offset, tt.maxRows, offset, tt.want)
		}
	}
}

func TestClampCursorScroll(t *testing.T) {
	tests := []struct {
		name    string
		cursor  int
		offset  int
		maxRows int
		want    int
	}{
		{"cursor above window", 2, 5, 5, 2},
		{"cursor below window", 10, 0, 5, 6},
		{"cursor at window bottom edge", 4, 0, 5, 0},
		{"cursor just past window", 5, 0, 5, 1},
		{"cursor inside window", 7, 5, 5, 5},
		{"negative offset", 0, -2, 5, 0},
	}
	for _, tt := range tests {
		offset := tt.offset
		clampCursorScroll(tt.cursor, &offset, tt.maxRows)
		if offset != tt.want {
			t.Errorf("%s: clampCursorScroll(%d, %d, %d) = %d, want %d", tt.name, tt.cursor, tt.offset, tt.maxRows, offset, tt.want)
		}
	}
}
