package backup

import (
	"context"
	"sync"
)

// countResources counts the total number of resources across specified namespaces concurrently
func (bm *Manager) countResources(ctx context.Context) int {
	namespaces, err := bm.client.ListNamespaces(ctx)
	if err != nil {
		bm.logger.Errorf("Error listing namespaces: %v", err)
		return 0
	}

	var wg sync.WaitGroup
	resourceCounts := make(chan int, len(namespaces)+1) // +1 for namespace count

	// Count namespaces themselves
	wg.Add(1)
	go func() {
		defer wg.Done()
		count, err := bm.countNamespaces(ctx)
		if err != nil {
			bm.logger.Errorf("Error counting namespaces: %v", err)
			resourceCounts <- 0
		} else {
			resourceCounts <- count
		}
	}()

	for _, ns := range namespaces {
		wg.Add(1)
		go func(namespace string) {
			defer wg.Done()
			count := bm.countResourcesInNamespace(ctx, namespace)
			resourceCounts <- count
		}(ns)
	}

	go func() {
		wg.Wait()
		close(resourceCounts)
	}()

	total := 0
	for count := range resourceCounts {
		total += count
	}

	return total
}

// countResourcesInNamespace counts the resources in a single namespace
func (bm *Manager) countResourcesInNamespace(ctx context.Context, namespace string) int {
	resourceTypes := map[string]func(context.Context, string) (int, error){
		"deployments":     bm.countDeployments,
		"services":        bm.countServices,
		"configmaps":      bm.countConfigMaps,
		"secrets":         bm.countSecrets,
		"serviceaccounts": bm.countServiceAccounts,
		"statefulsets":    bm.countStatefulSets,
		"daemonsets":      bm.countDaemonSets,
		"hpas":            bm.countHorizontalPodAutoscalers,
		"cronjobs":        bm.countCronJobs,
		"jobs":            bm.countJobs,
		"pvcs":            bm.countPersistentVolumeClaims,
		"ingresses":       bm.countIngresses,
		"rolebindings":    bm.countRoleBindings,
	}

	var wg sync.WaitGroup
	counts := make(chan int, len(resourceTypes))

	for name, countFn := range resourceTypes {
		wg.Add(1)
		go func(name string, countFn func(context.Context, string) (int, error)) {
			defer wg.Done()
			count, err := countFn(ctx, namespace)
			if err != nil {
				bm.logger.Errorf("Error counting %s in namespace %s: %v", name, namespace, err)
				counts <- 0
			} else {
				counts <- count
			}
		}(name, countFn)
	}

	go func() {
		wg.Wait()
		close(counts)
	}()

	total := 0
	for count := range counts {
		total += count
	}

	return total
}

// Helper functions to count each resource type
func (bm *Manager) countDeployments(ctx context.Context, namespace string) (int, error) {
	deployments, err := bm.client.ListDeployments(ctx, namespace)
	if err != nil {
		return 0, err
	}
	return len(deployments.Items), nil
}

func (bm *Manager) countServices(ctx context.Context, namespace string) (int, error) {
	services, err := bm.client.ListServices(ctx, namespace)
	if err != nil {
		return 0, err
	}
	return len(services.Items), nil
}

func (bm *Manager) countConfigMaps(ctx context.Context, namespace string) (int, error) {
	configMaps, err := bm.client.ListConfigMaps(ctx, namespace)
	if err != nil {
		return 0, err
	}
	return len(configMaps.Items), nil
}

func (bm *Manager) countSecrets(ctx context.Context, namespace string) (int, error) {
	secrets, err := bm.client.ListSecrets(ctx, namespace)
	if err != nil {
		return 0, err
	}
	return len(secrets.Items), nil
}

func (bm *Manager) countServiceAccounts(ctx context.Context, namespace string) (int, error) {
	serviceAccounts, err := bm.client.ListServiceAccounts(ctx, namespace)
	if err != nil {
		return 0, err
	}
	return len(serviceAccounts.Items), nil
}

func (bm *Manager) countStatefulSets(ctx context.Context, namespace string) (int, error) {
	statefulSets, err := bm.client.ListStatefulSets(ctx, namespace)
	if err != nil {
		return 0, err
	}
	return len(statefulSets.Items), nil
}

func (bm *Manager) countDaemonSets(ctx context.Context, namespace string) (int, error) {
	daemonSets, err := bm.client.ListDaemonSets(ctx, namespace)
	if err != nil {
		return 0, err
	}
	return len(daemonSets.Items), nil
}

func (bm *Manager) countHorizontalPodAutoscalers(ctx context.Context, namespace string) (int, error) {
	hpas, err := bm.client.ListHorizontalPodAutoscalers(ctx, namespace)
	if err != nil {
		return 0, err
	}
	return len(hpas.Items), nil
}

func (bm *Manager) countCronJobs(ctx context.Context, namespace string) (int, error) {
	cronJobs, err := bm.client.ListCronJobs(ctx, namespace)
	if err != nil {
		return 0, err
	}
	return len(cronJobs.Items), nil
}

func (bm *Manager) countPersistentVolumeClaims(ctx context.Context, namespace string) (int, error) {
	pvcs, err := bm.client.ListPersistentVolumeClaims(ctx, namespace)
	if err != nil {
		return 0, err
	}
	return len(pvcs.Items), nil
}

func (bm *Manager) countJobs(ctx context.Context, namespace string) (int, error) {
	jobs, err := bm.client.ListJobs(ctx, namespace)
	if err != nil {
		return 0, err
	}
	return len(jobs.Items), nil
}

func (bm *Manager) countIngresses(ctx context.Context, namespace string) (int, error) {
	ingresses, err := bm.client.ListIngresses(ctx, namespace)
	if err != nil {
		return 0, err
	}
	return len(ingresses.Items), nil
}

func (bm *Manager) countNamespaces(ctx context.Context) (int, error) {
	namespaces, err := bm.client.GetNamespaces(ctx)
	if err != nil {
		return 0, err
	}
	return len(namespaces.Items), nil
}

func (bm *Manager) countRoleBindings(ctx context.Context, namespace string) (int, error) {
	roleBindings, err := bm.client.ListRoleBindings(ctx, namespace)
	if err != nil {
		return 0, err
	}
	return len(roleBindings.Items), nil
}
