package backup

import (
	"context"
	"fmt"
	"path/filepath"
)

// backupResource handles the backup of a specific resource type in a namespace
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
	case "serviceaccounts":
		err = bm.backupServiceAccounts(ctx, namespace)
	case "statefulsets":
		err = bm.backupStatefulSets(ctx, namespace)
	case "daemonsets":
		err = bm.backupDaemonSets(ctx, namespace)
	case "hpas":
		err = bm.backupHorizontalPodAutoscalers(ctx, namespace)
	case "cronjobs":
		err = bm.backupCronJobs(ctx, namespace)
	case "jobs":
		err = bm.backupJobs(ctx, namespace)
	case "pvcs":
		err = bm.backupPersistantVolumeClaims(ctx, namespace)
	case "ingresses":
		err = bm.backupIngresses(ctx, namespace)
	case "roles":
		err = bm.backupRoles(ctx, namespace)
	case "networkpolicies":
		err = bm.backupNetworkPolicies(ctx, namespace)
	default:
		return fmt.Errorf("unknown resource type: %s", resourceType)
	}
	return err
}

// backupDeployments backs up all deployments in a given namespace
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

// backupServices backs up all services in a given namespace
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

// backupConfigMaps backs up all config maps in a given namespace
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

// backupSecrets backs up all secrets in a given namespace
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

// backupServiceAccounts backs up all service accounts in a given namespace
func (bm *Manager) backupServiceAccounts(ctx context.Context, namespace string) error {
	serviceAccounts, err := bm.client.ListServiceAccounts(ctx, namespace)
	if err != nil {
		return fmt.Errorf("error listing service accounts in namespace %s: %v", namespace, err)
	}

	for _, serviceAccount := range serviceAccounts.Items {
		filename := filepath.Join(bm.backupDir, namespace, "serviceaccounts", serviceAccount.Name+".json")
		if bm.dryRun {
			bm.logger.Infof("Would backup service account: %s/%s", namespace, serviceAccount.Name)
		} else {
			if err := bm.saveResource(serviceAccount, "ServiceAccount", filename); err != nil {
				return err
			}
		}
	}

	return nil
}

// backupStatefulSets backs up all stateful sets in a given namespace
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

// backupDaemonSets backs up all daemon sets in a given namespace
func (bm *Manager) backupDaemonSets(ctx context.Context, namespace string) error {
	daemonSets, err := bm.client.ListDaemonSets(ctx, namespace)
	if err != nil {
		return fmt.Errorf("error listing daemon sets in namespace %s: %v", namespace, err)
	}

	for _, daemonSet := range daemonSets.Items {
		filename := filepath.Join(bm.backupDir, namespace, "daemonsets", daemonSet.Name+".json")
		if bm.dryRun {
			bm.logger.Infof("Would backup daemon set: %s/%s", namespace, daemonSet.Name)
		} else {
			if err := bm.saveResource(daemonSet, "DaemonSet", filename); err != nil {
				return err
			}
		}
	}

	return nil
}

// backupHorizontalPodAutoscalers backs up all horizontal pod autoscalers in a given namespace
func (bm *Manager) backupHorizontalPodAutoscalers(ctx context.Context, namespace string) error {
	hpas, err := bm.client.ListHorizontalPodAutoscalers(ctx, namespace)
	if err != nil {
		return fmt.Errorf("error listing HPAs in namespace %s: %v", namespace, err)
	}

	for _, hpa := range hpas.Items {
		filename := filepath.Join(bm.backupDir, namespace, "hpas", hpa.Name+".json")
		if bm.dryRun {
			bm.logger.Infof("Would backup HPA: %s/%s", namespace, hpa.Name)
		} else {
			if err := bm.saveResource(hpa, "HorizontalPodAutoscaler", filename); err != nil {

				return err
			}
		}
	}

	return nil
}

// backupCronJobs backs up all cron jobs in a given namespace
func (bm *Manager) backupCronJobs(ctx context.Context, namespace string) error {
	cronJobs, err := bm.client.ListCronJobs(ctx, namespace)
	if err != nil {
		return fmt.Errorf("error listing cron jobs in namespace %s: %v", namespace, err)
	}

	for _, cronJob := range cronJobs.Items {
		filename := filepath.Join(bm.backupDir, namespace, "cronjobs", cronJob.Name+".json")
		if bm.dryRun {
			bm.logger.Infof("Would backup cron job: %s/%s", namespace, cronJob.Name)
		} else {
			if err := bm.saveResource(cronJob, "CronJob", filename); err != nil {
				return err
			}
		}
	}

	return nil
}

// backupPersistantVolumeClaims backs up all persistent volume claims in a given namespace
func (bm *Manager) backupPersistantVolumeClaims(ctx context.Context, namespace string) error {
	pvcs, err := bm.client.ListPersistentVolumeClaims(ctx, namespace)
	if err != nil {
		return fmt.Errorf("error listing persistant volume claims in namespace %s: %v", namespace, err)
	}
	for _, pvc := range pvcs.Items {
		filename := filepath.Join(bm.backupDir, namespace, "pvcs", pvc.Name+".json")
		if bm.dryRun {
			bm.logger.Infof("Would backup pvc: %s/%s", namespace, pvc.Name)
		} else {
			if err := bm.saveResource(pvc, "PersistentVolumeClaim", filename); err != nil {
				return err
			}
		}
	}
	return nil
}

// backupJobs backs up all jobs in a given namespace
func (bm *Manager) backupJobs(ctx context.Context, namespace string) error {
	jobs, err := bm.client.ListJobs(ctx, namespace)
	if err != nil {
		return fmt.Errorf("error listing jobs in namespace %s: %v", namespace, err)
	}

	for _, job := range jobs.Items {
		filename := filepath.Join(bm.backupDir, namespace, "jobs", job.Name+".json")
		if bm.dryRun {
			bm.logger.Infof("Would backup job: %s/%s", namespace, job.Name)
		} else {
			if err := bm.saveResource(job, "Job", filename); err != nil {
				return err
			}
		}
	}
	return nil
}

// backupIngresses backs up all ingresses in a given namespace
func (bm *Manager) backupIngresses(ctx context.Context, namespace string) error {
	ingresses, err := bm.client.ListIngresses(ctx, namespace)
	if err != nil {
		return fmt.Errorf("error listing ingresses in namespace %s: %v", namespace, err)
	}

	for _, ingress := range ingresses.Items {
		filename := filepath.Join(bm.backupDir, namespace, "ingresses", ingress.Name+".json")
		if bm.dryRun {
			bm.logger.Infof("Would backup ingress: %s/%s", namespace, ingress.Name)
		} else {
			if err := bm.saveResource(ingress, "Ingress", filename); err != nil {
				return err
			}
		}
	}

	return nil
}

// backupRoles backs up all roles in a given namespace
func (bm *Manager) backupRoles(ctx context.Context, namespace string) error {
	roles, err := bm.client.ListRoles(ctx, namespace)
	if err != nil {
		return fmt.Errorf("error listing roles in namespace %s: %v", namespace, err)
	}

	for _, role := range roles.Items {
		filename := filepath.Join(bm.backupDir, namespace, "roles", role.Name+".json")
		if bm.dryRun {
			bm.logger.Infof("Would backup role: %s/%s", namespace, role.Name)
		} else {
			if err := bm.saveResource(role, "Role", filename); err != nil {
				return err
			}
		}
	}

	return nil
}

// backupNetworkPolicies backs up all network policies in a given namespace
func (bm *Manager) backupNetworkPolicies(ctx context.Context, namespace string) error {
	networkPolicies, err := bm.client.ListNetworkPolicies(ctx, namespace)
	if err != nil {
		return fmt.Errorf("error listing network policies in namespace %s: %v", namespace, err)
	}

	for _, networkPolicy := range networkPolicies.Items {
		filename := filepath.Join(bm.backupDir, namespace, "networkpolicies", networkPolicy.Name+".json")
		if bm.dryRun {
			bm.logger.Infof("Would backup network policy: %s/%s", namespace, networkPolicy.Name)
		} else {
			if err := bm.saveResource(networkPolicy, "NetworkPolicy", filename); err != nil {
				return err
			}
		}
	}

	return nil
}

// backupNamespaces backs up all namespaces in the cluster
func (bm *Manager) backupNamespaces(ctx context.Context) error {
	namespaces, err := bm.client.GetNamespaces(ctx)
	if err != nil {
		return fmt.Errorf("error getting namespaces: %v", err)
	}

	for _, namespace := range namespaces.Items {
		// Namespaces are cluster-scoped, so we store them in a special directory
		filename := filepath.Join(bm.backupDir, "namespaces", namespace.Name+".json")
		if bm.dryRun {
			bm.logger.Infof("Would backup namespace: %s", namespace.Name)
		} else {
			if err := bm.saveResource(namespace, "Namespace", filename); err != nil {
				return err
			}
		}
	}
	return nil
}
