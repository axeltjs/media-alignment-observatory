package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// DefaultMediaListPath is where the source registry lives unless MEDIA_LIST_PATH
// overrides it.
const DefaultMediaListPath = "config/media_list.yaml"

// MediaSource is one news outlet RSS feed. Fields mirror the `sources` table.
type MediaSource struct {
	Name     string `yaml:"name"`
	RSSUrl   string `yaml:"rss_url"`
	BaseUrl  string `yaml:"base_url"`
	Category string `yaml:"category"`
	// IsActive is a pointer so that an omitted key means "active" rather than
	// false, which is what a plain bool would decode to.
	IsActive *bool `yaml:"is_active"`
	// RequiresBrowserUA is documentation only — the RSS services send a
	// browser User-Agent for every feed. It records which feeds depend on it.
	RequiresBrowserUA bool `yaml:"requires_browser_ua"`
}

// GovernmentSource is one official government feed. These are not stored in
// the `sources` table; they are read from the registry on every fetch and
// written into `government_contents`.
type GovernmentSource struct {
	Name     string `yaml:"name"`
	Agency   string `yaml:"agency"`
	RSSUrl   string `yaml:"rss_url"`
	BaseUrl  string `yaml:"base_url"`
	IsActive *bool  `yaml:"is_active"`
}

// MediaList is the parsed source registry. The `disabled` section of the YAML
// is intentionally not decoded — it is documentation of rejected feeds.
type MediaList struct {
	Media      []MediaSource      `yaml:"media"`
	Government []GovernmentSource `yaml:"government"`
}

func active(flag *bool) bool {
	return flag == nil || *flag
}

// ActiveMedia returns only the media feeds not explicitly disabled.
func (m MediaList) ActiveMedia() []MediaSource {
	out := make([]MediaSource, 0, len(m.Media))
	for _, s := range m.Media {
		if active(s.IsActive) {
			out = append(out, s)
		}
	}
	return out
}

// ActiveGovernment returns only the government feeds not explicitly disabled.
func (m MediaList) ActiveGovernment() []GovernmentSource {
	out := make([]GovernmentSource, 0, len(m.Government))
	for _, s := range m.Government {
		if active(s.IsActive) {
			out = append(out, s)
		}
	}
	return out
}

// MediaListPath resolves the registry path from MEDIA_LIST_PATH.
func MediaListPath() string {
	return GetEnvOr("MEDIA_LIST_PATH", DefaultMediaListPath)
}

// LoadMediaList reads and validates the source registry. A missing or invalid
// registry is fatal for seeding: without it the pipeline has nothing to fetch.
func LoadMediaList(path string) (MediaList, error) {
	var list MediaList

	raw, err := os.ReadFile(path)
	if err != nil {
		return list, fmt.Errorf("read media list %s: %w", path, err)
	}
	if err := yaml.Unmarshal(raw, &list); err != nil {
		return list, fmt.Errorf("parse media list %s: %w", path, err)
	}

	for i, s := range list.Media {
		if s.Name == "" || s.RSSUrl == "" {
			return list, fmt.Errorf("media[%d]: name and rss_url are required", i)
		}
	}
	for i, s := range list.Government {
		if s.RSSUrl == "" {
			return list, fmt.Errorf("government[%d]: rss_url is required", i)
		}
		if s.Agency == "" {
			return list, fmt.Errorf("government[%d] (%s): agency is required", i, s.Name)
		}
	}
	return list, nil
}
