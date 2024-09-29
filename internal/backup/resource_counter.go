package backup

import (
	"context"
)

// countResources counts the total number of resources across specified namespaces.
// It includes deployments, services, configmaps, and secrets.
func (bm *Manager) countResources(ctx context.Context, namespaces []string) int {
	total := 0
	for _, ns := range namespaces {
		// Count deployments
		deployments, err := bm.client.ListDeployments(ctx, ns)
		if err != nil {
			bm.logger.Errorf("Error listing deployments in namespace %s: %v", ns, err)
			continue
		}

		// Count services
		services, err := bm.client.ListServices(ctx, ns)
		if err != nil {
			bm.logger.Errorf("Error listing services in namespace %s: %v", ns, err)
			continue
		}

		// Count configmaps
		configMaps, err := bm.client.ListConfigMaps(ctx, ns)
		if err != nil {
			bm.logger.Errorf("Error listing configmaps in namespace %s: %v", ns, err)
			continue
		}

		// Count secrets
		secrets, err := bm.client.ListSecrets(ctx, ns)
		if err != nil {
			bm.logger.Errorf("Error listing secrets in namespace %s: %v", ns, err)
			continue
		}

		// Count HPAs
		hpas, err := bm.client.ListHorizontalPodAutoscalers(ctx, ns)
		if err != nil {
			bm.logger.Errorf("Error listing HPAs in namespace %s: %v", ns, err)
			continue
		}

		// Sum up the total number of resources
		total += len(deployments.Items) + len(services.Items) + len(configMaps.Items) + len(secrets.Items) + len(hpas.Items)
	}
	return total
}
