package backup

import (
	"context"
	"fmt"
	"path/filepath"
)

// backupResource handles the backup of a specific resource type in a namespace.
func (bm *Manager) backupResource(ctx context.Context, resourceType, namespace string) error {
	var err error
	switch resourceType {
	case "deployments":
		err = bm.backupDeployments(ctx, namespace)
	case "services":
		err = bm.backupServices(ctx, namespace)
	case "configmaps":
		err = bm.backupConfigMaps(ctx, namespace)
	case "secrets":
		err = bm.backupSecrets(ctx, namespace)
	case "statefulsets":
		err = bm.backupStatefulSets(ctx, namespace)
	default:
		return fmt.Errorf("unknown resource type: %s", resourceType)
	}
	return err
}

// backupDeployments backs up all deployments in a given namespace.
func (bm *Manager) backupDeployments(ctx context.Context, namespace string) error {
	deployments, err := bm.client.ListDeployments(ctx, namespace)
	if err != nil {
		return fmt.Errorf("error listing deployments in namespace %s: %v", namespace, err)
	}

	for _, deployment := range deployments.Items {
		filename := filepath.Join(bm.backupDir, namespace, "deployments", deployment.Name+".json")
		if bm.dryRun {
			bm.logger.Infof("Would backup deployment: %s/%s", namespace, deployment.Name)
		} else {
			if err := bm.saveResource(deployment, "Deployment", filename); err != nil {
				return err
			}
		}
	}

	return nil
}

// backupServices backs up all services in a given namespace.
func (bm *Manager) backupServices(ctx context.Context, namespace string) error {
	services, err := bm.client.ListServices(ctx, namespace)
	if err != nil {
		return fmt.Errorf("error listing services in namespace %s: %v", namespace, err)
	}

	for _, service := range services.Items {
		filename := filepath.Join(bm.backupDir, namespace, "services", service.Name+".json")
		if bm.dryRun {
			bm.logger.Infof("Would backup service: %s/%s", namespace, service.Name)
		} else {
			if err := bm.saveResource(service, "Service", filename); err != nil {
				return err
			}
		}
	}

	return nil
}

// backupConfigMaps backs up all config maps in a given namespace.
func (bm *Manager) backupConfigMaps(ctx context.Context, namespace string) error {
	configMaps, err := bm.client.ListConfigMaps(ctx, namespace)
	if err != nil {
		return fmt.Errorf("error listing configmaps in namespace %s: %v", namespace, err)
	}

	for _, configMap := range configMaps.Items {
		filename := filepath.Join(bm.backupDir, namespace, "configmaps", configMap.Name+".json")
		if bm.dryRun {
			bm.logger.Infof("Would backup configmap: %s/%s", namespace, configMap.Name)
		} else {
			if err := bm.saveResource(configMap, "ConfigMap", filename); err != nil {
				return err
			}
		}
	}

	return nil
}

// backupSecrets backs up all secrets in a given namespace.
func (bm *Manager) backupSecrets(ctx context.Context, namespace string) error {
	secrets, err := bm.client.ListSecrets(ctx, namespace)
	if err != nil {
		return fmt.Errorf("error listing secrets in namespace %s: %v", namespace, err)
	}

	for _, secret := range secrets.Items {
		filename := filepath.Join(bm.backupDir, namespace, "secrets", secret.Name+".json")
		if bm.dryRun {
			bm.logger.Infof("Would backup secret: %s/%s", namespace, secret.Name)
		} else {
			if err := bm.saveResource(secret, "Secret", filename); err != nil {
				return err
			}
		}
	}

	return nil
}

// backupStatefulSets backs up all stateful sets in a given namespace.
func (bm *Manager) backupStatefulSets(ctx context.Context, namespace string) error {
	statefulSets, err := bm.client.ListStatefulSets(ctx, namespace)
	if err != nil {
		return fmt.Errorf("error listing stateful sets in namespace %s: %v", namespace, err)
	}

	for _, statefulSet := range statefulSets.Items {
		filename := filepath.Join(bm.backupDir, namespace, "statefulsets", statefulSet.Name+".json")
		if bm.dryRun {
			bm.logger.Infof("Would backup stateful set: %s/%s", namespace, statefulSet.Name)
		} else {
			if err := bm.saveResource(statefulSet, "StatefulSet", filename); err != nil {
				return err
			}
		}
	}

	return nil
}
