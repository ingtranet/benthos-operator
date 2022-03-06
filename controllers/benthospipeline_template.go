package controllers

import (
	"github.com/ingtranet/benthos-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getConfigMapFor(p *v1alpha1.BenthosPipeline) (*corev1.ConfigMap, error) {
	labels, err := getLabels(p)
	if err != nil {
		return nil, err
	}
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.Name,
			Namespace: p.Namespace,
			Labels:    labels,
			Annotations: map[string]string{
				"ingtra.net/benthos-config-hash": p.GetConfigHash(),
			},
		},
		Data: map[string]string{
			"benthos.yaml": p.GetYamlConfig(),
		},
	}
	return cm, nil
}

func getDeploymentFor(p *v1alpha1.BenthosPipeline) (*appsv1.Deployment, error) {
	labels, err := getLabels(p)
	zero := int32(0)
	if err != nil {
		return nil, err
	}
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.Name,
			Namespace: p.Namespace,
			Labels:    labels,
			Annotations: map[string]string{
				"ingtra.net/benthos-config-hash": p.GetConfigHash(),
			},
		},
		Spec: appsv1.DeploymentSpec{
			RevisionHistoryLimit: &zero,
			Replicas:             &p.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
					Annotations: map[string]string{
						"ingtra.net/benthos-config-hash": p.GetConfigHash(),
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						ImagePullPolicy: corev1.PullAlways,
						Image:           p.Spec.Image,
						Name:            "benthos",
						Args:            []string{"-c", "/config/benthos.yaml"},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "config",
							MountPath: "/config",
						}},
					}},
					Volumes: []corev1.Volume{{
						Name: "config",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: p.Name,
								},
							},
						},
					}},
				},
			},
		},
	}
	return dep, nil
}

func getLabels(p *v1alpha1.BenthosPipeline) (map[string]string, error) {
	labels := map[string]string{
		"app.kubernetes.io/name":     "benthos",
		"app.kubernetes.io/instance": "benthos-pipeline-" + p.Name,
	}
	return labels, nil
}
