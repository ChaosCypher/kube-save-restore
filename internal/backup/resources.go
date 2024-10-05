package backup

import (
	"context"
	"fmt"
	"path/filepath"
)

// backupResource handles the backup of a specific resource type in a namespace.
func (bm *Manager) backupResource(ctx context.Context, resourceType, namespace string) error {
	bm.logger.Debugf("Backing up %s in namespace %s", resourceType, namespace)
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
	case "hpas":
		err = bm.backupHorizontalPodAutoscalers(ctx, namespace)
	case "cronjobs":
		err = bm.backupCronJobs(ctx, namespace)
	default:
		return fmt.Errorf("unknown resource type: %s", resourceType)
	}
	if err != nil {
		bm.logger.Errorf("Failed to backup %s in namespace %s: %v", resourceType, namespace, err)
	}
	return err
}

// backupDeployments backs up all deployments in a given namespace.
func (bm *Manager) backupDeployments(ctx context.Context, namespace string) error {
	deployments, err := bm.client.ListDeployments(ctx, namespace)
	if err != nil {
		return fmt.Errorf("error listing deployments in namespace %s: %v", namespace, err)
	}

	bm.logger.Debugf("Found %d deployments in namespace %s", len(deployments.Items), namespace)

	for _, deployment := range deployments.Items {
		filename := filepath.Join(bm.backupDir, namespace, "deployments", deployment.Name+".json")
		if bm.dryRun {
			bm.logger.Infof("Would backup deployment: %s/%s", namespace, deployment.Name)
		} else {
			if err := bm.saveResource(deployment, "Deployment", filename); err != nil {
				bm.logger.Errorf("Failed to backup deployment %s/%s: %v", namespace, deployment.Name, err)
				return err
			}
			bm.logger.Debugf("Backed up deployment: %s/%s", namespace, deployment.Name)
		}
		bm.incrementResourceCount("deployments")
	}

	return nil
}

// backupServices backs up all services in a given namespace.
func (bm *Manager) backupServices(ctx context.Context, namespace string) error {
	services, err := bm.client.ListServices(ctx, namespace)
	if err != nil {
		return fmt.Errorf("error listing services in namespace %s: %v", namespace, err)
	}

	bm.logger.Debugf("Found %d services in namespace %s", len(services.Items), namespace)

	for _, service := range services.Items {
		filename := filepath.Join(bm.backupDir, namespace, "services", service.Name+".json")
		if bm.dryRun {
			bm.logger.Infof("Would backup service: %s/%s", namespace, service.Name)
		} else {
			if err := bm.saveResource(service, "Service", filename); err != nil {
				bm.logger.Errorf("Failed to backup service %s/%s: %v", namespace, service.Name, err)
				return err
			}
			bm.logger.Debugf("Backed up service: %s/%s", namespace, service.Name)
		}
		bm.incrementResourceCount("services")
	}

	return nil
}

// backupConfigMaps backs up all config maps in a given namespace.
func (bm *Manager) backupConfigMaps(ctx context.Context, namespace string) error {
	configMaps, err := bm.client.ListConfigMaps(ctx, namespace)
	if err != nil {
		return fmt.Errorf("error listing configmaps in namespace %s: %v", namespace, err)
	}

	bm.logger.Debugf("Found %d configmaps in namespace %s", len(configMaps.Items), namespace)

	for _, configMap := range configMaps.Items {
		filename := filepath.Join(bm.backupDir, namespace, "configmaps", configMap.Name+".json")
		if bm.dryRun {
			bm.logger.Infof("Would backup configmap: %s/%s", namespace, configMap.Name)
		} else {
			if err := bm.saveResource(configMap, "ConfigMap", filename); err != nil {
				bm.logger.Errorf("Failed to backup configmap %s/%s: %v", namespace, configMap.Name, err)
				return err
			}
			bm.logger.Debugf("Backed up configmap: %s/%s", namespace, configMap.Name)
		}
		bm.incrementResourceCount("configmaps")
	}

	return nil
}

// backupSecrets backs up all secrets in a given namespace.
func (bm *Manager) backupSecrets(ctx context.Context, namespace string) error {
	secrets, err := bm.client.ListSecrets(ctx, namespace)
	if err != nil {
		return fmt.Errorf("error listing secrets in namespace %s: %v", namespace, err)
	}

	bm.logger.Debugf("Found %d secrets in namespace %s", len(secrets.Items), namespace)

	for _, secret := range secrets.Items {
		filename := filepath.Join(bm.backupDir, namespace, "secrets", secret.Name+".json")
		if bm.dryRun {
			bm.logger.Infof("Would backup secret: %s/%s", namespace, secret.Name)
		} else {
			if err := bm.saveResource(secret, "Secret", filename); err != nil {
				bm.logger.Errorf("Failed to backup secret %s/%s: %v", namespace, secret.Name, err)
				return err
			}
			bm.logger.Debugf("Backed up secret: %s/%s", namespace, secret.Name)
		}
		bm.incrementResourceCount("secrets")
	}

	return nil
}

// backupStatefulSets backs up all stateful sets in a given namespace.
func (bm *Manager) backupStatefulSets(ctx context.Context, namespace string) error {
	statefulSets, err := bm.client.ListStatefulSets(ctx, namespace)
	if err != nil {
		return fmt.Errorf("error listing statefulsets in namespace %s: %v", namespace, err)
	}

	bm.logger.Debugf("Found %d statefulsets in namespace %s", len(statefulSets.Items), namespace)

	for _, statefulSet := range statefulSets.Items {
		filename := filepath.Join(bm.backupDir, namespace, "statefulsets", statefulSet.Name+".json")
		if bm.dryRun {
			bm.logger.Infof("Would backup statefulset: %s/%s", namespace, statefulSet.Name)
		} else {
			if err := bm.saveResource(statefulSet, "StatefulSet", filename); err != nil {
				bm.logger.Errorf("Failed to backup statefulset %s/%s: %v", namespace, statefulSet.Name, err)
				return err
			}
			bm.logger.Debugf("Backed up statefulset: %s/%s", namespace, statefulSet.Name)
		}
		bm.incrementResourceCount("statefulsets")
	}

	return nil
}

// backupHorizontalPodAutoscalers backs up all horizontal pod autoscalers in a given namespace.
func (bm *Manager) backupHorizontalPodAutoscalers(ctx context.Context, namespace string) error {
	hpas, err := bm.client.ListHorizontalPodAutoscalers(ctx, namespace)
	if err != nil {
		return fmt.Errorf("error listing HPAs in namespace %s: %v", namespace, err)
	}

	bm.logger.Debugf("Found %d HPAs in namespace %s", len(hpas.Items), namespace)

	for _, hpa := range hpas.Items {
		filename := filepath.Join(bm.backupDir, namespace, "hpas", hpa.Name+".json")
		if bm.dryRun {
			bm.logger.Infof("Would backup HPA: %s/%s", namespace, hpa.Name)
		} else {
			if err := bm.saveResource(hpa, "HorizontalPodAutoscaler", filename); err != nil {
				bm.logger.Errorf("Failed to backup HPA %s/%s: %v", namespace, hpa.Name, err)
				return err
			}
			bm.logger.Debugf("Backed up HPA: %s/%s", namespace, hpa.Name)
		}
		bm.incrementResourceCount("hpas")
	}

	return nil
}

// backupCronJobs backs up all cron jobs in a given namespace.
func (bm *Manager) backupCronJobs(ctx context.Context, namespace string) error {
	cronJobs, err := bm.client.ListCronJobs(ctx, namespace)
	if err != nil {
		return fmt.Errorf("error listing cronjobs in namespace %s: %v", namespace, err)
	}

	bm.logger.Debugf("Found %d cronjobs in namespace %s", len(cronJobs.Items), namespace)

	for _, cronJob := range cronJobs.Items {
		filename := filepath.Join(bm.backupDir, namespace, "cronjobs", cronJob.Name+".json")
		if bm.dryRun {
			bm.logger.Infof("Would backup cronjob: %s/%s", namespace, cronJob.Name)
		} else {
			if err := bm.saveResource(cronJob, "CronJob", filename); err != nil {
				bm.logger.Errorf("Failed to backup cronjob %s/%s: %v", namespace, cronJob.Name, err)
				return err
			}
			bm.logger.Debugf("Backed up cronjob: %s/%s", namespace, cronJob.Name)
		}
		bm.incrementResourceCount("cronjobs")
	}

	return nil
}
