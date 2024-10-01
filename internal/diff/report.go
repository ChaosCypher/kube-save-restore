package diff

// DiffReport holds the results of the diff operation.
type DiffReport struct {
	Resources map[string]map[string]*ResourceWrapper
	Missing   map[string][]string
}

// NewDiffReport initializes a new DiffReport.
func NewDiffReport() *DiffReport {
	return &DiffReport{
		Resources: make(map[string]map[string]*ResourceWrapper),
		Missing:   make(map[string][]string),
	}
}

// AddResource adds a resource to the report for a specific backup directory.
func (dr *DiffReport) AddResource(key, dir string, resource *ResourceWrapper) {
	if dr.Resources[key] == nil {
		dr.Resources[key] = make(map[string]*ResourceWrapper)
	}
	dr.Resources[key][dir] = resource
}

// AddMissing adds a missing resource for a specific backup directory.
func (dr *DiffReport) AddMissing(key, dir string) {
	if dr.Missing[key] == nil {
		dr.Missing[key] = []string{}
	}
	dr.Missing[key] = append(dr.Missing[key], dir)
}
