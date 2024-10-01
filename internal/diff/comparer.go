package diff

import (
	"github.com/chaoscypher/k8s-backup-restore/internal/logger"
)

// Comparer handles the comparison of resources across backups.
type Comparer struct {
	resourceMaps map[string]map[string]*ResourceWrapper
	logger       logger.LoggerInterface
}

// NewComparer creates a new Comparer instance.
func NewComparer(resourceMaps map[string]map[string]*ResourceWrapper, logger logger.LoggerInterface) *Comparer {
	return &Comparer{
		resourceMaps: resourceMaps,
		logger:       logger,
	}
}

// Compare performs the comparison and returns a DiffReport.
func (c *Comparer) Compare() *DiffReport {
	report := NewDiffReport()

	// Aggregate all resources
	for backupDir, resources := range c.resourceMaps {
		for key, resource := range resources {
			if _, exists := report.Resources[key]; !exists {
				report.Resources[key] = make(map[string]*ResourceWrapper)
			}
			report.Resources[key][backupDir] = resource
		}
	}

	// Analyze the collected resources to identify differences
	for key, backups := range report.Resources {
		if len(backups) < len(c.resourceMaps) {
			report.Missing[key] = c.identifyMissing(backups)
		}
		// Further comparison logic can be added here (e.g., content differences)
	}

	return report
}

// identifyMissing identifies which backups are missing the given resource.
func (c *Comparer) identifyMissing(backups map[string]*ResourceWrapper) []string {
	missing := []string{}
	for _, dir := range c.backupDirs() {
		if _, exists := backups[dir]; !exists {
			missing = append(missing, dir)
		}
	}
	return missing
}

// backupDirs returns the list of backup directories.
func (c *Comparer) backupDirs() []string {
	dirs := make([]string, 0, len(c.resourceMaps))
	for dir := range c.resourceMaps {
		dirs = append(dirs, dir)
	}
	return dirs
}