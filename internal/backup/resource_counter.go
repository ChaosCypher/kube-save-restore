package backup

import (
	"context"
)

func (bm *Manager) countResources(ctx context.Context, namespaces []string) int {
	total := 0
	for _, ns := range namespaces {
		deployments, err := bm.client.ListDeployments(ctx, ns)
		if err != nil {
			bm.logger.Errorf("Error listing deployments in namespace %s: %v", ns, err)
			continue
		}

		services, err := bm.client.ListServices(ctx, ns)
		if err != nil {
			bm.logger.Errorf("Error listing services in namespace %s: %v", ns, err)
			continue
		}

		configMaps, err := bm.client.ListConfigMaps(ctx, ns)
		if err != nil {
			bm.logger.Errorf("Error listing configmaps in namespace %s: %v", ns, err)
			continue
		}

		secrets, err := bm.client.ListSecrets(ctx, ns)
		if err != nil {
			bm.logger.Errorf("Error listing secrets in namespace %s: %v", ns, err)
			continue
		}

		total += len(deployments.Items) + len(services.Items) + len(configMaps.Items) + len(secrets.Items)
	}
	return total
}
