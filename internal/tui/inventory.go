package tui

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/chxmxii/a3/internal/storage"
)

type inventoryView struct {
	resources    []storage.Resource
	regions      []string
	regionIdx    int    // -1 means "all regions"
	typeFilter   string // empty means all types
	typeIdx      int
	cursor       int
	offset       int
	showDetail   bool // true when viewing a resource's details
	detailScroll int
	searchInput  bool   // true while the search input line is active
	searchQuery  string // case-insensitive substring filter, empty means off
}

// capturesInput reports whether the view is in text-entry mode and must
// receive every key press, including global ones.
func (v *inventoryView) capturesInput() bool {
	return v.searchInput
}

// handleKey handles a key press for the inventory view. Returns true if the
// key was handled. The scroll offset is clamped to the cursor at render time.
func (v *inventoryView) handleKey(key string) bool {
	if v.searchInput {
		return v.handleSearchKey(key)
	}
	switch key {
	case "up", "k":
		if v.showDetail {
			if v.detailScroll > 0 {
				v.detailScroll--
			}
		} else if v.cursor > 0 {
			v.cursor--
		}
	case "down", "j":
		if v.showDetail {
			v.detailScroll++
		} else if v.cursor < len(v.filteredResources())-1 {
			v.cursor++
		}
	case "enter":
		v.toggleDetail()
	case "esc":
		if v.showDetail {
			v.showDetail = false
			v.detailScroll = 0
		}
	case "r":
		v.nextRegion()
	case "R":
		v.prevRegion()
	case "t":
		v.nextType()
	case "T":
		v.prevType()
	case "x":
		v.clearFilters()
	case "/":
		if !v.showDetail {
			v.searchInput = true
		}
	default:
		return false
	}
	return true
}

// handleSearchKey handles a key press while the search input is active.
// Every key is considered handled.
func (v *inventoryView) handleSearchKey(key string) bool {
	switch key {
	case "enter":
		// Confirm: keep the filter, leave input mode.
		v.searchInput = false
	case "esc":
		// Cancel: clear the filter and leave input mode.
		v.searchInput = false
		v.searchQuery = ""
		v.cursor = 0
		v.offset = 0
	case "backspace":
		if v.searchQuery != "" {
			runes := []rune(v.searchQuery)
			v.searchQuery = string(runes[:len(runes)-1])
			v.cursor = 0
			v.offset = 0
		}
	default:
		if key == "space" {
			key = " "
		}
		if utf8.RuneCountInString(key) == 1 {
			v.searchQuery += key
			v.cursor = 0
			v.offset = 0
		}
	}
	return true
}

func (v *inventoryView) nextRegion() {
	if v.showDetail {
		return
	}
	v.regionIdx++
	if v.regionIdx >= len(v.regions) {
		v.regionIdx = -1
	}
	v.cursor = 0
	v.offset = 0
}

func (v *inventoryView) prevRegion() {
	if v.showDetail {
		return
	}
	v.regionIdx--
	if v.regionIdx < -1 {
		v.regionIdx = len(v.regions) - 1
	}
	v.cursor = 0
	v.offset = 0
}

func (v *inventoryView) nextType() {
	if v.showDetail {
		return
	}
	types := v.availableTypes()
	v.typeIdx++
	if v.typeIdx >= len(types) {
		v.typeIdx = -1
		v.typeFilter = ""
	} else {
		v.typeFilter = types[v.typeIdx]
	}
	v.cursor = 0
	v.offset = 0
}

func (v *inventoryView) prevType() {
	if v.showDetail {
		return
	}
	types := v.availableTypes()
	v.typeIdx--
	if v.typeIdx < -1 {
		v.typeIdx = len(types) - 1
		v.typeFilter = types[v.typeIdx]
	} else if v.typeIdx == -1 {
		v.typeFilter = ""
	} else {
		v.typeFilter = types[v.typeIdx]
	}
	v.cursor = 0
	v.offset = 0
}

func (v *inventoryView) clearFilters() {
	if v.showDetail {
		v.showDetail = false
		v.detailScroll = 0
		return
	}
	v.regionIdx = -1
	v.typeFilter = ""
	v.typeIdx = -1
	v.searchQuery = ""
	v.cursor = 0
	v.offset = 0
}

func (v *inventoryView) toggleDetail() {
	if v.showDetail {
		v.showDetail = false
		v.detailScroll = 0
	} else {
		v.showDetail = true
		v.detailScroll = 0
	}
}

func (v *inventoryView) availableTypes() []string {
	typeSet := make(map[string]bool)
	for _, r := range v.resources {
		typeSet[r.ResourceType] = true
	}
	var types []string
	for t := range typeSet {
		types = append(types, t)
	}
	sort.Strings(types)
	return types
}

func (v *inventoryView) currentRegion() string {
	if v.regionIdx < 0 || v.regionIdx >= len(v.regions) {
		return ""
	}
	return v.regions[v.regionIdx]
}

func (v *inventoryView) filteredResources() []storage.Resource {
	region := v.currentRegion()
	var filtered []storage.Resource
	for _, r := range v.resources {
		if region != "" && r.Region != region {
			continue
		}
		if v.typeFilter != "" && r.ResourceType != v.typeFilter {
			continue
		}
		if v.searchQuery != "" && !resourceMatchesSearch(r, v.searchQuery) {
			continue
		}
		filtered = append(filtered, r)
	}
	return filtered
}

// resourceMatchesSearch reports whether the resource matches the search
// query as a case-insensitive substring of its name, ID, type or region.
func resourceMatchesSearch(r storage.Resource, query string) bool {
	query = strings.ToLower(query)
	for _, field := range []string{r.Name, r.ResourceID, r.ResourceType, r.Region} {
		if strings.Contains(strings.ToLower(field), query) {
			return true
		}
	}
	return false
}

func (v *inventoryView) selectedResource() *storage.Resource {
	filtered := v.filteredResources()
	if v.cursor < 0 || v.cursor >= len(filtered) {
		return nil
	}
	r := filtered[v.cursor]
	return &r
}

func (v *inventoryView) render(width, height int) string {
	if v.showDetail {
		return v.renderDetail(width, height)
	}
	return v.renderList(width, height)
}

func (v *inventoryView) renderList(width, height int) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  📦 Resource Inventory"))
	b.WriteString("\n\n")

	// Filter status bar.
	region := v.currentRegion()
	filterParts := []string{}
	if region != "" {
		filterParts = append(filterParts, regionBadgeStyle.Render("Region: "+region))
	} else {
		filterParts = append(filterParts, dimNavStyle.Render("Region: all"))
	}
	if v.typeFilter != "" {
		filterParts = append(filterParts, regionBadgeStyle.Render("Type: "+v.typeFilter))
	} else {
		filterParts = append(filterParts, dimNavStyle.Render("Type: all"))
	}
	if v.searchInput {
		filterParts = append(filterParts, regionBadgeStyle.Render("/"+v.searchQuery+"█"))
	} else if v.searchQuery != "" {
		filterParts = append(filterParts, regionBadgeStyle.Render("Search: "+v.searchQuery))
	}
	b.WriteString("  " + strings.Join(filterParts, "  "))
	b.WriteString("\n")

	filtered := v.filteredResources()
	b.WriteString(dimNavStyle.Render(fmt.Sprintf("  %d of %d resources  (Enter: details)", len(filtered), len(v.resources))))
	b.WriteString("\n\n")

	// Table header.
	header := fmt.Sprintf("  %-18s %-35s %-18s %s", "TYPE", "NAME", "REGION", "ID")
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n")
	b.WriteString(dimNavStyle.Render("  " + strings.Repeat("─", min(width-4, 95))))
	b.WriteString("\n")

	// Visible rows.
	maxRows := height - 11
	if maxRows < 5 {
		maxRows = 5
	}

	clampCursorScroll(v.cursor, &v.offset, maxRows)

	end := min(v.offset+maxRows, len(filtered))

	for i := v.offset; i < end; i++ {
		r := filtered[i]
		shortID := r.ResourceID
		if len(shortID) > 20 {
			shortID = "..." + shortID[len(shortID)-17:]
		}
		name := r.Name
		if name == "" {
			name = "(unnamed)"
		}
		if len(name) > 33 {
			name = name[:30] + "..."
		}

		line := fmt.Sprintf("  %-18s %-35s %-18s %s", r.ResourceType, name, r.Region, shortID)
		if i == v.cursor {
			b.WriteString(selectedStyle.Render(line))
		} else {
			b.WriteString(normalStyle.Render(line))
		}
		b.WriteString("\n")
	}

	b.WriteString(scrollFooter(len(filtered), v.offset, maxRows))

	return b.String()
}

func (v *inventoryView) renderDetail(width, height int) string {
	r := v.selectedResource()
	if r == nil {
		v.showDetail = false
		return v.renderList(width, height)
	}

	lines := v.buildDetailLines(r, width)

	// Apply scroll.
	maxRows := height - 4
	if maxRows < 10 {
		maxRows = 10
	}

	var b strings.Builder
	b.WriteString(renderScrollable(lines, &v.detailScroll, maxRows))

	if len(lines) > maxRows {
		b.WriteString(dimNavStyle.Render(fmt.Sprintf("  ↕ %d%%  (Esc/x: back)", scrollPct(len(lines), v.detailScroll, maxRows))))
	} else {
		b.WriteString(dimNavStyle.Render("  Esc/x: back to list"))
	}

	return b.String()
}

func (v *inventoryView) buildDetailLines(r *storage.Resource, width int) []string {
	// Type-aware rendering for specific resource types.
	switch r.ResourceType {
	case "security_group":
		return v.buildSGDetail(r, width)
	case "iam_policy":
		return v.buildPolicyDetail(r, width)
	case "route_table":
		return v.buildRouteTableDetail(r, width)
	default:
		return v.buildGenericDetail(r, width)
	}
}

// sgRule is one parsed security group rule row.
type sgRule struct {
	proto      string
	portRange  string
	target     string // CIDR, "sg:<id>" reference, or placeholder label
	desc       string
	isOpen     bool // CIDR is 0.0.0.0/0 or ::/0
	isCIDR     bool // rule came from an IP range (vs SG ref or placeholder)
	isFallback bool // synthesized "(self/all)"/"(all)" placeholder row
}

// parseSGRules parses the ip_permissions entries of a security group into
// flat rule rows. For egress rules, security group references are skipped
// and the placeholder label is "(all)" instead of "(self/all)".
func parseSGRules(perms []any, egress bool) []sgRule {
	var rules []sgRule
	for _, perm := range perms {
		permMap, ok := perm.(map[string]any)
		if !ok {
			continue
		}
		proto := getDetailStr(permMap, "ip_protocol", "IpProtocol")
		if proto == "-1" {
			proto = "ALL"
		}

		fromPort := getFloat(permMap, "from_port", "FromPort")
		toPort := getFloat(permMap, "to_port", "ToPort")
		portRange := "ALL"
		if proto != "ALL" && fromPort >= 0 {
			if fromPort == toPort {
				portRange = fmt.Sprintf("%.0f", fromPort)
			} else {
				portRange = fmt.Sprintf("%.0f-%.0f", fromPort, toPort)
			}
		}

		// IP ranges.
		ipRanges := getSlice(permMap, "ip_ranges", "IpRanges")
		for _, ipr := range ipRanges {
			iprMap, ok := ipr.(map[string]any)
			if !ok {
				continue
			}
			cidr := getDetailStr(iprMap, "cidr_ip", "CidrIp")
			rules = append(rules, sgRule{
				proto:     proto,
				portRange: portRange,
				target:    cidr,
				desc:      getDetailStr(iprMap, "description", "Description"),
				isOpen:    cidr == "0.0.0.0/0" || cidr == "::/0",
				isCIDR:    true,
			})
		}

		// Security group references (inbound only).
		var sgRefs []any
		if !egress {
			sgRefs = getSlice(permMap, "user_id_group_pairs", "UserIdGroupPairs")
			for _, sgr := range sgRefs {
				sgrMap, ok := sgr.(map[string]any)
				if !ok {
					continue
				}
				rules = append(rules, sgRule{
					proto:     proto,
					portRange: portRange,
					target:    "sg:" + getDetailStr(sgrMap, "group_id", "GroupId"),
					desc:      getDetailStr(sgrMap, "description", "Description"),
				})
			}
		}

		// If no ranges or refs matched, show the rule anyway.
		if len(ipRanges) == 0 && len(sgRefs) == 0 {
			label := "(self/all)"
			if egress {
				label = "(all)"
			}
			rules = append(rules, sgRule{
				proto:      proto,
				portRange:  portRange,
				target:     label,
				isFallback: true,
			})
		}
	}
	return rules
}

// renderRuleSection renders one boxed rule section (title, column header and
// parsed rows). CIDR rows are styled, with open (0.0.0.0/0 or ::/0) CIDRs
// highlighted in both directions.
func renderRuleSection(title, targetCol string, perms []any, egress bool) []string {
	lines := []string{
		headerStyle.Render("  ┌─ " + title),
		fmt.Sprintf("  │ %s  %s  %s  %s", keyStyle.Render("PROTO     "), keyStyle.Render("PORTS      "), keyStyle.Render(targetCol), keyStyle.Render("DESCRIPTION")),
		"  │ " + strings.Repeat("─", 70),
	}
	if len(perms) == 0 {
		lines = append(lines, "  │ (none)")
	}
	for _, rule := range parseSGRules(perms, egress) {
		switch {
		case rule.isFallback:
			lines = append(lines, fmt.Sprintf("  │ %-10s %-12s %-22s", rule.proto, rule.portRange, rule.target))
		case rule.isCIDR:
			style := normalStyle
			if rule.isOpen {
				style = severityHighStyle
			}
			lines = append(lines, style.Render(fmt.Sprintf("  │ %-10s %-12s %-22s %s", rule.proto, rule.portRange, rule.target, rule.desc)))
		default:
			lines = append(lines, fmt.Sprintf("  │ %-10s %-12s %-22s %s", rule.proto, rule.portRange, rule.target, rule.desc))
		}
	}
	lines = append(lines, "  └─")
	return lines
}

func (v *inventoryView) buildSGDetail(r *storage.Resource, width int) []string {
	var lines []string
	meta := r.RawMetadata

	lines = append(lines, titleStyle.Render(fmt.Sprintf("  Security Group: %s", r.Name)))
	lines = append(lines, "")

	// Basic info.
	lines = append(lines, headerStyle.Render("  ┌─ Info"))
	lines = append(lines, fmt.Sprintf("  │ %s %s", keyStyle.Render("Group ID:   "), getDetailStr(meta, "group_id")))
	lines = append(lines, fmt.Sprintf("  │ %s %s", keyStyle.Render("Name:       "), getDetailStr(meta, "group_name")))
	lines = append(lines, fmt.Sprintf("  │ %s %s", keyStyle.Render("Description:"), getDetailStr(meta, "description")))
	lines = append(lines, fmt.Sprintf("  │ %s %s", keyStyle.Render("VPC:        "), getDetailStr(meta, "vpc_id")))
	lines = append(lines, fmt.Sprintf("  │ %s %s", keyStyle.Render("Region:     "), r.Region))
	lines = append(lines, "  └─")
	lines = append(lines, "")

	// Inbound rules.
	ipPerms, _ := meta["ip_permissions"].([]any)
	lines = append(lines, renderRuleSection("Inbound Rules", "SOURCE                ", ipPerms, false)...)
	lines = append(lines, "")

	// Outbound rules.
	ipPermsEgress, _ := meta["ip_permissions_egress"].([]any)
	lines = append(lines, renderRuleSection("Outbound Rules", "DESTINATION           ", ipPermsEgress, true)...)

	return lines
}

func (v *inventoryView) buildPolicyDetail(r *storage.Resource, width int) []string {
	var lines []string
	meta := r.RawMetadata

	lines = append(lines, titleStyle.Render(fmt.Sprintf("  IAM Policy: %s", r.Name)))
	lines = append(lines, "")

	// Basic info.
	lines = append(lines, headerStyle.Render("  ┌─ Info"))
	lines = append(lines, fmt.Sprintf("  │ ARN:          %s", r.ResourceID))
	lines = append(lines, fmt.Sprintf("  │ Name:         %s", r.Name))
	policyID := getDetailStr(meta, "policy_id")
	if policyID != "" {
		lines = append(lines, fmt.Sprintf("  │ Policy ID:    %s", policyID))
	}
	isManaged := getDetailStr(meta, "is_aws_managed")
	if isManaged != "" {
		lines = append(lines, fmt.Sprintf("  │ AWS Managed:  %s", isManaged))
	}
	attachCount := meta["attachment_count"]
	if attachCount != nil {
		lines = append(lines, fmt.Sprintf("  │ Attached to:  %v entities", attachCount))
	}
	lines = append(lines, "  └─")
	lines = append(lines, "")

	// Policy document.
	policyDoc := meta["policy"]
	if policyDoc == nil {
		policyDoc = meta["policy_std"]
	}
	if policyDoc == nil {
		policyDoc = meta["document"]
	}

	if policyDoc != nil {
		lines = append(lines, headerStyle.Render("  ┌─ Policy Document"))
		lines = append(lines, v.renderPolicyDoc(policyDoc)...)
		lines = append(lines, "  └─")
	} else {
		lines = append(lines, dimNavStyle.Render("  (Policy document not available — requires iam:GetPolicyVersion permission)"))
	}

	return lines
}

func (v *inventoryView) renderPolicyDoc(doc any) []string {
	var lines []string

	switch d := doc.(type) {
	case map[string]any:
		// Render Statement array if present.
		stmts, _ := d["Statement"].([]any)
		if stmts == nil {
			stmts, _ = d["statement"].([]any)
		}
		if len(stmts) > 0 {
			for i, stmt := range stmts {
				stmtMap, ok := stmt.(map[string]any)
				if !ok {
					continue
				}
				effect := getDetailStr(stmtMap, "Effect", "effect")

				effectStyle := passStyle
				if effect == "Deny" {
					effectStyle = failStyle
				}

				lines = append(lines, fmt.Sprintf("  │ Statement %d: %s", i+1, effectStyle.Render(effect)))

				// Actions.
				actions := extractStringList(stmtMap, "Action")
				if len(actions) == 0 {
					actions = extractStringList(stmtMap, "action")
				}
				for _, a := range actions {
					lines = append(lines, fmt.Sprintf("  │   Action:   %s", a))
				}

				// Resources.
				resources := extractStringList(stmtMap, "Resource")
				if len(resources) == 0 {
					resources = extractStringList(stmtMap, "resource")
				}
				for _, res := range resources {
					lines = append(lines, fmt.Sprintf("  │   Resource: %s", res))
				}

				lines = append(lines, "  │")
			}
		} else {
			// No Statement key — dump as formatted key-value.
			for k, val := range d {
				lines = append(lines, fmt.Sprintf("  │ %s: %v", k, val))
			}
		}
	case string:
		// JSON string — show line by line.
		for _, line := range strings.Split(d, "\n") {
			lines = append(lines, "  │ "+line)
		}
	default:
		lines = append(lines, fmt.Sprintf("  │ %v", doc))
	}

	return lines
}

func extractStringList(m map[string]any, key string) []string {
	val, ok := m[key]
	if !ok || val == nil {
		return nil
	}
	switch v := val.(type) {
	case string:
		return []string{v}
	case []any:
		var result []string
		for _, item := range v {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	default:
		return []string{fmt.Sprintf("%v", v)}
	}
}

func (v *inventoryView) buildRouteTableDetail(r *storage.Resource, width int) []string {
	var lines []string
	meta := r.RawMetadata

	lines = append(lines, titleStyle.Render(fmt.Sprintf("  Route Table: %s", r.Name)))
	lines = append(lines, "")

	// Basic info.
	lines = append(lines, headerStyle.Render("  ┌─ Info"))
	lines = append(lines, fmt.Sprintf("  │ %s %s", keyStyle.Render("Route Table ID:"), getDetailStr(meta, "route_table_id")))
	lines = append(lines, fmt.Sprintf("  │ %s %s", keyStyle.Render("VPC:           "), getDetailStr(meta, "vpc_id")))
	lines = append(lines, fmt.Sprintf("  │ %s %s", keyStyle.Render("Region:        "), r.Region))
	lines = append(lines, "  └─")
	lines = append(lines, "")

	// Routes.
	lines = append(lines, headerStyle.Render("  ┌─ Routes"))
	lines = append(lines, fmt.Sprintf("  │ %s  %s  %s", keyStyle.Render("DESTINATION         "), keyStyle.Render("TARGET                        "), keyStyle.Render("STATE")))
	lines = append(lines, "  │ "+strings.Repeat("─", 65))

	routes, _ := meta["routes"].([]any)
	if len(routes) == 0 {
		lines = append(lines, "  │ (no routes found — may require additional permissions)")
	}
	for _, route := range routes {
		routeMap, ok := route.(map[string]any)
		if !ok {
			continue
		}
		dest := getDetailStr(routeMap, "destination_cidr_block", "DestinationCidrBlock", "destination_ipv6_cidr_block", "DestinationIpv6CidrBlock")

		target := ""
		for _, field := range []string{"gateway_id", "GatewayId", "nat_gateway_id", "NatGatewayId", "instance_id", "InstanceId", "transit_gateway_id", "TransitGatewayId", "vpc_peering_connection_id", "VpcPeeringConnectionId", "network_interface_id", "NetworkInterfaceId"} {
			t := getDetailStr(routeMap, field)
			if t != "" && t != "<nil>" {
				target = t
				break
			}
		}
		if target == "" {
			target = "local"
		}

		state := getDetailStr(routeMap, "state", "State")
		if state == "" {
			state = "active"
		}

		// Color the target based on type.
		targetStyled := target
		switch {
		case target == "local":
			targetStyled = routeLocalStyle.Render(target)
		case len(target) > 4 && target[:4] == "igw-":
			targetStyled = routeIGWStyle.Render(target)
		case len(target) > 4 && target[:4] == "nat-":
			targetStyled = routeNATStyle.Render(target)
		case len(target) > 4 && target[:4] == "tgw-":
			targetStyled = regionBadgeStyle.Render(target)
		}

		lines = append(lines, fmt.Sprintf("  │ %-20s %-30s %-10s", dest, targetStyled, state))
	}
	lines = append(lines, "  └─")
	lines = append(lines, "")

	// Associations.
	lines = append(lines, headerStyle.Render("  ┌─ Associations"))
	assocs, _ := meta["associations"].([]any)
	if len(assocs) == 0 {
		lines = append(lines, "  │ (none)")
	}
	for _, assoc := range assocs {
		assocMap, ok := assoc.(map[string]any)
		if !ok {
			continue
		}
		subnetID := getDetailStr(assocMap, "subnet_id")
		main, _ := assocMap["main"].(bool)
		if main {
			lines = append(lines, "  │ Main route table (implicit association)")
		} else if subnetID != "" {
			lines = append(lines, fmt.Sprintf("  │ Subnet: %s", subnetID))
		}
	}
	lines = append(lines, "  └─")

	return lines
}

func (v *inventoryView) buildGenericDetail(r *storage.Resource, width int) []string {
	var lines []string

	lines = append(lines, titleStyle.Render(fmt.Sprintf("  Resource: %s", r.Name)))
	lines = append(lines, "")

	// Core fields.
	lines = append(lines, headerStyle.Render("  ┌─ Identity"))
	lines = append(lines, fmt.Sprintf("  │ %s %s", keyStyle.Render("ID:    "), r.ResourceID))
	lines = append(lines, fmt.Sprintf("  │ %s %s", keyStyle.Render("Type:  "), r.ResourceType))
	lines = append(lines, fmt.Sprintf("  │ %s %s", keyStyle.Render("Name:  "), r.Name))
	lines = append(lines, fmt.Sprintf("  │ %s %s", keyStyle.Render("Region:"), r.Region))
	lines = append(lines, "  └─")
	lines = append(lines, "")

	// Tags.
	if len(r.Tags) > 0 {
		lines = append(lines, headerStyle.Render(fmt.Sprintf("  ┌─ Tags (%d)", len(r.Tags))))
		keys := make([]string, 0, len(r.Tags))
		for k := range r.Tags {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			val := r.Tags[k]
			if len(val) > 60 {
				val = val[:57] + "..."
			}
			lines = append(lines, fmt.Sprintf("  │ %s %s", keyStyle.Render(fmt.Sprintf("%-25s", k+":")), valueStyle.Render(val)))
		}
		lines = append(lines, "  └─")
		lines = append(lines, "")
	}

	// Metadata.
	if len(r.RawMetadata) > 0 {
		lines = append(lines, headerStyle.Render(fmt.Sprintf("  ┌─ Metadata (%d fields)", len(r.RawMetadata))))

		keys := make([]string, 0, len(r.RawMetadata))
		for k := range r.RawMetadata {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			val := r.RawMetadata[k]
			valStr := formatMetadataValue(val, width-30)
			lines = append(lines, fmt.Sprintf("  │ %s %s", keyStyle.Render(fmt.Sprintf("%-25s", k+":")), valueStyle.Render(valStr)))
		}
		lines = append(lines, "  └─")
	}

	return lines
}

// getDetailStr returns the first non-empty value found under keys.
// String values are returned as-is; other values are formatted with %v.
func getDetailStr(m map[string]any, keys ...string) string {
	if m == nil {
		return ""
	}
	for _, key := range keys {
		v, ok := m[key]
		if !ok || v == nil {
			continue
		}
		s, isStr := v.(string)
		if !isStr {
			s = fmt.Sprintf("%v", v)
		}
		if s != "" {
			return s
		}
	}
	return ""
}

func getFloat(m map[string]any, keys ...string) float64 {
	for _, key := range keys {
		v, ok := m[key]
		if !ok || v == nil {
			continue
		}
		switch f := v.(type) {
		case float64:
			return f
		case int:
			return float64(f)
		case int64:
			return float64(f)
		}
	}
	return -1
}

func getSlice(m map[string]any, keys ...string) []any {
	for _, key := range keys {
		v, ok := m[key]
		if !ok || v == nil {
			continue
		}
		if s, ok := v.([]any); ok {
			return s
		}
	}
	return nil
}

func formatMetadataValue(val any, maxLen int) string {
	if val == nil {
		return dimNavStyle.Render("null")
	}
	if maxLen < 20 {
		maxLen = 20
	}

	switch v := val.(type) {
	case string:
		if len(v) > maxLen {
			return v[:maxLen-3] + "..."
		}
		return v
	case bool:
		if v {
			return passStyle.Render("true")
		}
		return failStyle.Render("false")
	case float64:
		if v == float64(int(v)) {
			return fmt.Sprintf("%d", int(v))
		}
		return fmt.Sprintf("%.2f", v)
	case map[string]any:
		if len(v) == 0 {
			return dimNavStyle.Render("{}")
		}
		parts := []string{}
		for k, inner := range v {
			s := fmt.Sprintf("%v", inner)
			if len(s) > 20 {
				s = s[:17] + "..."
			}
			parts = append(parts, k+"="+s)
			if len(parts) >= 3 {
				parts = append(parts, "...")
				break
			}
		}
		result := "{" + strings.Join(parts, ", ") + "}"
		if len(result) > maxLen {
			result = result[:maxLen-3] + "..."
		}
		return result
	case []any:
		if len(v) == 0 {
			return dimNavStyle.Render("[]")
		}
		return fmt.Sprintf("[%d items]", len(v))
	default:
		s := fmt.Sprintf("%v", v)
		if len(s) > maxLen {
			s = s[:maxLen-3] + "..."
		}
		return s
	}
}
