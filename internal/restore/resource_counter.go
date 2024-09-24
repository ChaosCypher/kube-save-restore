package restore

// countResources counts the number of resources in the provided list of files.
// It takes a slice of strings representing file names and returns the total count as an integer.
func (rm *Manager) countResources(files []string) int {
	total := 0
	// Iterate over the files slice and increment the total count for each file.
	for range files {
		total += 1
	}
	return total
}
