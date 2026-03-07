package storage

import (
	"path"
	"strings"
)

type Kind string

const (
	KindRaw        Kind = "raw"
	KindSnapshot   Kind = "snapshot"
	KindExtracted  Kind = "extracted"
	KindNormalized Kind = "normalized"
	KindDebug      Kind = "debug"
)

func ResourceObjectKey(groupID, resourceID string, kind Kind, filename string) string {
	return path.Join(
		"groups",
		groupID,
		"resources",
		resourceID,
		kind.pathSegment(),
		normalizeObjectName(filename),
	)
}

func (k Kind) pathSegment() string {
	switch k {
	case KindRaw:
		return "raw"
	case KindSnapshot:
		return "artifacts/snapshot"
	case KindExtracted:
		return "artifacts/extracted"
	case KindNormalized:
		return "artifacts/normalized"
	case KindDebug:
		return "artifacts/debug"
	default:
		return "artifacts/misc"
	}
}

func normalizeObjectName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "..", ".")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	if name == "" {
		return "unnamed"
	}
	return name
}
