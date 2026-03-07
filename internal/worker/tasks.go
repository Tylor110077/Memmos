package worker

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

const expandNodeDedupeWindow = 5 * time.Minute

type FetchWebResourcePayload struct {
	ResourceID string `json:"resource_id"`
	URL        string `json:"url"`
}

type ParseResourcePayload struct {
	ResourceID string `json:"resource_id"`
}

type NormalizeResourcePayload struct {
	ResourceID string `json:"resource_id"`
}

type IndexGroupContextPayload struct {
	GroupID    string `json:"group_id"`
	ResourceID string `json:"resource_id"`
}

type GenerateFrameworkGraphPayload struct {
	GroupID              string   `json:"group_id"`
	ResourceGraphVersion []string `json:"resource_graph_versions,omitempty"`
}

func DedupeKeyForResourceStage(resourceID, stage, version string) string {
	trimmedVersion := strings.TrimSpace(version)
	if trimmedVersion == "" {
		trimmedVersion = "latest"
	}
	return fmt.Sprintf("resource:%s:%s:%s", resourceID, stage, trimmedVersion)
}

func DedupeKeyForFrameworkGraph(groupID string, versions []string) string {
	clean := append([]string(nil), versions...)
	sort.Strings(clean)
	if len(clean) == 0 {
		return fmt.Sprintf("framework:%s:empty", groupID)
	}
	return fmt.Sprintf("framework:%s:%s", groupID, strings.Join(clean, ","))
}

func DedupeKeyForExpandNode(graphID, nodeID string, now time.Time) string {
	window := now.UTC().Unix() / int64(expandNodeDedupeWindow.Seconds())
	return fmt.Sprintf("expand:%s:%s:%d", graphID, nodeID, window)
}
