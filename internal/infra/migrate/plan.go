package migrate

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Migration struct {
	Version string
	Name    string
	UpSQL   string
	DownSQL string
}

type Plan struct {
	Migrations []Migration
}

func LoadPlan(dir string) (Plan, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return Plan{}, fmt.Errorf("read migrations dir: %w", err)
	}

	byVersion := map[string]*Migration{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		version, name, direction, ok := parseFilename(entry.Name())
		if !ok {
			continue
		}

		item := byVersion[version]
		if item == nil {
			item = &Migration{Version: version, Name: name}
			byVersion[version] = item
		}

		raw, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return Plan{}, fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}

		switch direction {
		case "up":
			item.UpSQL = string(raw)
		case "down":
			item.DownSQL = string(raw)
		}
	}

	versions := make([]string, 0, len(byVersion))
	for version, item := range byVersion {
		if item.UpSQL == "" || item.DownSQL == "" {
			return Plan{}, fmt.Errorf("migration %s must have both up and down files", version)
		}
		versions = append(versions, version)
	}
	sort.Strings(versions)

	plan := Plan{Migrations: make([]Migration, 0, len(versions))}
	for _, version := range versions {
		plan.Migrations = append(plan.Migrations, *byVersion[version])
	}
	return plan, nil
}

func parseFilename(name string) (version, migrationName, direction string, ok bool) {
	switch {
	case strings.HasSuffix(name, ".up.sql"):
		direction = "up"
		name = strings.TrimSuffix(name, ".up.sql")
	case strings.HasSuffix(name, ".down.sql"):
		direction = "down"
		name = strings.TrimSuffix(name, ".down.sql")
	default:
		return "", "", "", false
	}

	version, migrationName, ok = strings.Cut(name, "_")
	if !ok || version == "" || migrationName == "" {
		return "", "", "", false
	}
	return version, migrationName, direction, true
}
