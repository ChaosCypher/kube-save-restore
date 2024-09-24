package restore

func (rm *Manager) countResources(files []string) int {
	total := 0
	for range files {
		total += 1
	}
	return total
}
