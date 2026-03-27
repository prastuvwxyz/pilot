package store

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Standard agent files shown in the file browser (in display order).
var agentFileOrder = []string{
	"IDENTITY.md",
	"AGENTS.md",
	"SOUL.md",
	"TOOLS.md",
	"HEARTBEAT.md",
	"USER.md",
}

type AgentInfo struct {
	ID      string // folder name, or "main" for root agent
	Name    string
	Role    string
	Emoji   string
	Tagline string
	Avatar  string // path or "_none_"
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
		case "avatar":
			info.Avatar = val
		}
	}
	return info, nil
}

// agentDir returns the filesystem directory for a given agent ID.
func agentDir(memoryPath, id string) string {
	if id == "main" {
		return memoryPath
	}
	return filepath.Join(memoryPath, "agents", id)
}

// AgentFiles returns the ordered list of standard .md files that exist for an agent.
func AgentFiles(memoryPath, id string) []string {
	dir := agentDir(memoryPath, id)
	var files []string
	for _, name := range agentFileOrder {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			files = append(files, name)
		}
	}
	return files
}

// ReadAgentFile reads the content of a named file for an agent.
func ReadAgentFile(memoryPath, id, filename string) (string, error) {
	path := filepath.Join(agentDir(memoryPath, id), filepath.Base(filename))
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteAgentFile overwrites the content of a named file for an agent.
func WriteAgentFile(memoryPath, id, filename, content string) error {
	path := filepath.Join(agentDir(memoryPath, id), filepath.Base(filename))
	return os.WriteFile(path, []byte(content), 0644)
}

// LoadAgents reads the main agent (root IDENTITY.md) and all sub-agents
// from the agents/ subdirectory.
func LoadAgents(memoryPath string) ([]AgentInfo, error) {
	var agents []AgentInfo

	// Main agent (root)
	rootIdentity := filepath.Join(memoryPath, "IDENTITY.md")
	if _, err := os.Stat(rootIdentity); err == nil {
		if info, err := parseIdentity(rootIdentity, "main"); err == nil {
			agents = append(agents, info)
		}
	}

	// Sub-agents in agents/
	agentsDir := filepath.Join(memoryPath, "agents")
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

// GetAgent returns a single AgentInfo by ID.
func GetAgent(memoryPath, id string) (AgentInfo, error) {
	var identityPath string
	if id == "main" {
		identityPath = filepath.Join(memoryPath, "IDENTITY.md")
	} else {
		identityPath = filepath.Join(memoryPath, "agents", id, "IDENTITY.md")
	}
	return parseIdentity(identityPath, id)
}
