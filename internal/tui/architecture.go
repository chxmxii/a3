package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/chxmxii/a3/internal/storage"
)

const (
	ArchModeNetwork  = "network"
	ArchModeResource = "resource"
)

type architectureView struct {
	resources     []storage.Resource
	relationships []storage.Relationship
	scrollOffset  int
	mode          string
}

// handleKey handles a key press for the architecture view. Returns true if
// the key was handled.
func (v *architectureView) handleKey(key string) bool {
	switch key {
	case "up", "k":
		if v.scrollOffset > 0 {
			v.scrollOffset--
		}
	case "down", "j":
		v.scrollOffset++
	case "n":
		v.mode = ArchModeNetwork
		v.scrollOffset = 0
	case "v":
		v.mode = ArchModeResource
		v.scrollOffset = 0
	default:
		return false
	}
	return true
}

func (v *architectureView) render(width, height int) string {
	if v.mode == "" {
		v.mode = ArchModeNetwork
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render("  🏗️  Architecture"))
	b.WriteString("\n")

	// Mode indicator.
	networkLabel := "Network Architecture (n)"
	resourceLabel := "Resource Architecture (v)"
	if v.mode == ArchModeNetwork {
		networkLabel = selectedStyle.Render(" " + networkLabel + " ")
		resourceLabel = dimNavStyle.Render(" " + resourceLabel + " ")
	} else {
		networkLabel = dimNavStyle.Render(" " + networkLabel + " ")
		resourceLabel = selectedStyle.Render(" " + resourceLabel + " ")
	}
	b.WriteString("  " + networkLabel + dimNavStyle.Render(" | ") + resourceLabel)
	b.WriteString("\n\n")

	if len(v.resources) == 0 && len(v.relationships) == 0 {
		b.WriteString(normalStyle.Render("  No resources or relationships discovered."))
		b.WriteString("\n")
		b.WriteString(dimNavStyle.Render("  This can happen when resources lack cross-references or permissions are limited."))
		return b.String()
	}

	b.WriteString(dimNavStyle.Render(fmt.Sprintf("  %d resources, %d relationships mapped", len(v.resources), len(v.relationships))))
	b.WriteString("\n\n")

	// Build name/type lookups.
	nameMap := make(map[string]string)
	typeMap := make(map[string]string)
	for _, r := range v.resources {
		display := r.Name
		if display == "" {
			display = r.ResourceID
		}
		nameMap[r.ResourceID] = display
		typeMap[r.ResourceID] = r.ResourceType
	}

	var lines []string
	switch v.mode {
	case ArchModeNetwork:
		lines = v.buildNetworkLines(nameMap, typeMap)
	case ArchModeResource:
		lines = v.buildResourceLines(nameMap, typeMap)
	default:
		lines = v.buildNetworkLines(nameMap, typeMap)
	}

	if len(lines) == 0 {
		b.WriteString(dimNavStyle.Render("  No resources found for this view mode."))
		return b.String()
	}

	// Apply scroll. The clamped offset is intentionally not written back to
	// v.scrollOffset, matching the original behavior.
	maxRows := height - 10
	if maxRows < 5 {
		maxRows = 5
	}
	offset := v.scrollOffset
	b.WriteString(renderScrollable(lines, &offset, maxRows))

	if len(lines) > maxRows {
		b.WriteString(dimNavStyle.Render(fmt.Sprintf("\n  ↕ scroll %d%% (%d lines)", scrollPct(len(lines), offset, maxRows), len(lines))))
	}

	return b.String()
}

// typeEntry describes how a resource type is displayed and which
// architecture views it appears in.
type typeEntry struct {
	label    string
	style    lipgloss.Style
	network  bool // shown in the network view
	resource bool // shown in the resource view
}

// typeTable drives colorType and the per-view type membership sets.
var typeTable = map[string]typeEntry{
	"vpc":              {label: "[vpc]", style: titleStyle, network: true},
	"subnet":           {label: "[subnet]", style: regionBadgeStyle, network: true},
	"route_table":      {label: "[route_table]", style: dimNavStyle, network: true},
	"security_group":   {label: "[security_group]", style: severityMediumStyle, network: true},
	"internet_gateway": {label: "[igw]", style: routeIGWStyle, network: true},
	"nat_gateway":      {label: "[nat]", style: routeNATStyle, network: true},
	"transit_gateway":  {label: "[tgw]", style: regionBadgeStyle, network: true},
	"ec2_instance":     {label: "[ec2]", style: passStyle, network: true, resource: true},
	"alb":              {label: "[alb]", style: severityMediumStyle, resource: true},
	"nlb":              {label: "[nlb]", style: severityMediumStyle, resource: true},
	"target_group":     {label: "[target_group]", style: dimNavStyle, resource: true},
	"eks_cluster":      {label: "[eks]", style: titleStyle, resource: true},
	"eks_node_group":   {label: "[node_group]", style: regionBadgeStyle, resource: true},
	"ecs_cluster":      {label: "[ecs]", style: titleStyle, resource: true},
	"lambda_function":  {label: "[lambda]", style: routeNATStyle, resource: true},
	"rds_instance":     {label: "[rds]", style: severityHighStyle, resource: true},
	"efs_file_system":  {label: "[efs]", style: routeLocalStyle, resource: true},
}

// typeSet builds a type membership set from typeTable.
func typeSet(pick func(typeEntry) bool) map[string]bool {
	set := make(map[string]bool)
	for t, e := range typeTable {
		if pick(e) {
			set[t] = true
		}
	}
	return set
}

// networkTypes are the resource types shown in the network view.
var networkTypes = typeSet(func(e typeEntry) bool { return e.network })

// resourceViewTypes are the resource types shown in the resource view.
var resourceViewTypes = typeSet(func(e typeEntry) bool { return e.resource })

// associateToParent maps each resource to its parent ID, preferring an
// explicit relationship in parentOf and falling back to the metaKey value in
// the resource's metadata.
func associateToParent(resources []storage.Resource, parentOf map[string]string, metaKey string) map[string]string {
	assoc := make(map[string]string)
	for _, r := range resources {
		if parent, hasParent := parentOf[r.ResourceID]; hasParent {
			assoc[r.ResourceID] = parent
			continue
		}
		if r.RawMetadata != nil {
			if id, ok := r.RawMetadata[metaKey].(string); ok && id != "" {
				assoc[r.ResourceID] = id
			}
		}
	}
	return assoc
}

// filterByParent returns the resources associated to parentID in assoc,
// preserving order.
func filterByParent(resources []storage.Resource, assoc map[string]string, parentID string) []storage.Resource {
	var out []storage.Resource
	for _, r := range resources {
		if assoc[r.ResourceID] == parentID {
			out = append(out, r)
		}
	}
	return out
}

// treeBranch returns the connector for a tree node and the indent to prepend
// for its children.
func treeBranch(isLast bool) (connector, childIndent string) {
	if isLast {
		return "└── ", "    "
	}
	return "├── ", "│   "
}

// treeNodeLine renders one tree node line: prefix, connector, colored type
// label and display name.
func treeNodeLine(prefix, connector string, r storage.Resource, nameMap map[string]string) string {
	name := nameMap[r.ResourceID]
	if name == "" {
		name = r.ResourceID
	}
	return fmt.Sprintf("%s%s%s %s", prefix, connector, colorType(r.ResourceType), name)
}

// buildNetworkLines builds the network topology view:
// VPC → Subnets → (Route Tables, EC2), Security Groups, Gateways
func (v *architectureView) buildNetworkLines(nameMap, typeMap map[string]string) []string {
	var lines []string

	// Index resources by type.
	resourcesByType := make(map[string][]storage.Resource)
	for _, r := range v.resources {
		if networkTypes[r.ResourceType] {
			resourcesByType[r.ResourceType] = append(resourcesByType[r.ResourceType], r)
		}
	}

	// Build relationship map: parentOf[childID] = parent resource ID.
	parentOf := make(map[string]string)
	for _, rel := range v.relationships {
		srcType := typeMap[rel.SourceID]
		tgtType := typeMap[rel.TargetID]
		if !networkTypes[srcType] || !networkTypes[tgtType] {
			continue
		}
		parentOf[rel.TargetID] = rel.SourceID
	}

	// Associate children to parents via relationships or metadata.
	gwTypes := []string{"internet_gateway", "nat_gateway", "transit_gateway"}
	var gateways []storage.Resource
	for _, gwType := range gwTypes {
		gateways = append(gateways, resourcesByType[gwType]...)
	}
	vpcForSubnet := associateToParent(resourcesByType["subnet"], parentOf, "vpc_id")
	subnetForEC2 := associateToParent(resourcesByType["ec2_instance"], parentOf, "subnet_id")
	vpcForSG := associateToParent(resourcesByType["security_group"], parentOf, "vpc_id")
	vpcForGW := associateToParent(gateways, parentOf, "vpc_id")
	subnetForRT := associateToParent(resourcesByType["route_table"], parentOf, "subnet_id")

	// Render VPC trees.
	vpcs := resourcesByType["vpc"]
	if len(vpcs) == 0 {
		lines = append(lines, dimNavStyle.Render("  No VPCs discovered"))
		return lines
	}

	for vi, vpc := range vpcs {
		vpcConnector, vpcIndent := treeBranch(vi == len(vpcs)-1)
		vpcChildPrefix := "  " + vpcIndent

		lines = append(lines, treeNodeLine("  ", vpcConnector, vpc, nameMap))

		// Collect children for this VPC: subnets, then SGs, then gateways.
		children := filterByParent(resourcesByType["subnet"], vpcForSubnet, vpc.ResourceID)
		children = append(children, filterByParent(resourcesByType["security_group"], vpcForSG, vpc.ResourceID)...)
		children = append(children, filterByParent(gateways, vpcForGW, vpc.ResourceID)...)

		for ci, child := range children {
			connector, indent := treeBranch(ci == len(children)-1)
			lines = append(lines, treeNodeLine(vpcChildPrefix, connector, child, nameMap))

			if child.ResourceType != "subnet" {
				continue
			}

			// Route tables and EC2 in this subnet.
			subChildPrefix := vpcChildPrefix + indent
			subChildren := filterByParent(resourcesByType["route_table"], subnetForRT, child.ResourceID)
			subChildren = append(subChildren, filterByParent(resourcesByType["ec2_instance"], subnetForEC2, child.ResourceID)...)
			for sci, sc := range subChildren {
				scConnector, _ := treeBranch(sci == len(subChildren)-1)
				lines = append(lines, treeNodeLine(subChildPrefix, scConnector, sc, nameMap))
			}
		}
	}

	return lines
}

// buildResourceLines builds the application-level resource view.
func (v *architectureView) buildResourceLines(nameMap, typeMap map[string]string) []string {
	var lines []string

	// Index resources by type.
	resourcesByType := make(map[string][]storage.Resource)
	for _, r := range v.resources {
		if resourceViewTypes[r.ResourceType] {
			resourcesByType[r.ResourceType] = append(resourcesByType[r.ResourceType], r)
		}
	}

	// Build relationship map (source → targets) for resource view types only.
	childrenOf := make(map[string][]string)
	parentOf := make(map[string]string)
	for _, rel := range v.relationships {
		srcType := typeMap[rel.SourceID]
		tgtType := typeMap[rel.TargetID]
		if !resourceViewTypes[srcType] || !resourceViewTypes[tgtType] {
			continue
		}
		childrenOf[rel.SourceID] = append(childrenOf[rel.SourceID], rel.TargetID)
		parentOf[rel.TargetID] = rel.SourceID
	}

	// Root resources are those that have no parent within the resource view types.
	var roots []storage.Resource
	for _, r := range v.resources {
		if !resourceViewTypes[r.ResourceType] {
			continue
		}
		if _, hasParent := parentOf[r.ResourceID]; !hasParent {
			roots = append(roots, r)
		}
	}

	if len(roots) == 0 {
		lines = append(lines, dimNavStyle.Render("  No application resources discovered"))
		return lines
	}

	rendered := make(map[string]bool)
	for ri, root := range roots {
		if rendered[root.ResourceID] {
			continue
		}
		isLast := ri == len(roots)-1
		v.buildResourceTree(&lines, root.ResourceID, "  ", isLast, childrenOf, nameMap, typeMap, rendered, 0)
	}

	return lines
}

func (v *architectureView) buildResourceTree(lines *[]string, resourceID, prefix string, isLast bool, childrenOf map[string][]string, nameMap, typeMap map[string]string, rendered map[string]bool, depth int) {
	if depth > 5 || rendered[resourceID] {
		return
	}
	rendered[resourceID] = true

	connector := "├── "
	if isLast {
		connector = "└── "
	}
	if depth == 0 {
		connector = ""
	}

	name := nameMap[resourceID]
	if name == "" {
		name = resourceID
	}
	if len(name) > 40 {
		name = name[:37] + "..."
	}
	rType := typeMap[resourceID]
	if rType == "" {
		rType = "?"
	}

	line := fmt.Sprintf("%s%s%s %s", prefix, connector, colorType(rType), name)
	*lines = append(*lines, line)

	children := childrenOf[resourceID]
	childPrefix := prefix + "│   "
	if isLast || depth == 0 {
		childPrefix = prefix + "    "
	}

	for i, childID := range children {
		isChildLast := i == len(children)-1
		v.buildResourceTree(lines, childID, childPrefix, isChildLast, childrenOf, nameMap, typeMap, rendered, depth+1)
	}
}

// colorType returns a styled type label based on the resource type.
func colorType(resType string) string {
	if e, ok := typeTable[resType]; ok {
		return e.style.Render(e.label)
	}
	return normalStyle.Render("[" + resType + "]")
}
