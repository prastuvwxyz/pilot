package store

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type AgentInfo struct {
	ID      string // folder name, or "main" for root agent
	Name    string
	Role    string
	Emoji   string
	Tagline string
}

var (
	reSimple = regexp.MustCompile(`(?i)^(\w[\w\s]*):\s*(.+)$`)          // Key: Value
	reBullet = regexp.MustCompile(`(?i)^-\s+\*\*([^*:]+):\*\*\s*(.+)$`) // - **Key:** Value
)

func parseIdentity(path, id string) (AgentInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return AgentInfo{ID: id}, err
	}

	info := AgentInfo{ID: id}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		var key, val string

		if m := reBullet.FindStringSubmatch(line); m != nil {
			key, val = strings.TrimSpace(m[1]), strings.TrimSpace(m[2])
		} else if m := reSimple.FindStringSubmatch(line); m != nil {
			key, val = strings.TrimSpace(m[1]), strings.TrimSpace(m[2])
		} else {
			continue
		}

		val = strings.Trim(val, `"`)
		switch strings.ToLower(key) {
		case "name":
			info.Name = val
		case "role":
			info.Role = val
		case "emoji":
			info.Emoji = val
		case "tagline":
			info.Tagline = val
		}
	}
	return info, nil
}

// LoadAgents reads the main agent (root IDENTITY.md) and all sub-agents
// from the agents/ subdirectory.
func LoadAgents(prasMemoryPath string) ([]AgentInfo, error) {
	var agents []AgentInfo

	// Main agent (root)
	rootIdentity := filepath.Join(prasMemoryPath, "IDENTITY.md")
	if _, err := os.Stat(rootIdentity); err == nil {
		if info, err := parseIdentity(rootIdentity, "main"); err == nil {
			agents = append(agents, info)
		}
	}

	// Sub-agents in agents/
	agentsDir := filepath.Join(prasMemoryPath, "agents")
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		return agents, nil // agents/ missing is non-fatal
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		id := entry.Name()
		identityPath := filepath.Join(agentsDir, id, "IDENTITY.md")
		if info, err := parseIdentity(identityPath, id); err == nil {
			agents = append(agents, info)
		}
	}

	return agents, nil
}
