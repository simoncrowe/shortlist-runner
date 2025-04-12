package jobs

import (
	"context"
	"os"
	"strings"

	"github.com/google/uuid"
	schemav1 "github.com/simoncrowe/shortlist-schema/lib/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type K8sRepository struct{}

func (r K8sRepository) Create(ctx context.Context, profile schemav1.Profile) (string, error) {
	cs, err := createClientset()
	if err != nil {
		return "", err
	}

	id := uuid.New()
	jobName := strings.Join([]string{"assessor", id.String()}, "-")

	configData, err := os.ReadFile(os.Getenv("CONFIG_PATH"))
	if err != nil {
		return "", err
	}
	profileData, err := schemav1.EncodeProfileJSON(profile)
	if err != nil {
		return "", err
	}
	cmCfg := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: jobName},
		Data: map[string]string{"profile.json": profileData,
			"config.json": string(configData)},
	}
	cmOpts := metav1.CreateOptions{}
	cm, err := cs.CoreV1().ConfigMaps("shortlist").Create(ctx, &cmCfg, cmOpts)
	if err != nil {
		return "", err
	}

	profileVol := corev1.Volume{
		Name: "profile",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: cm.ObjectMeta.Name,
				},
			},
		},
	}
	profileVolMnt := corev1.VolumeMount{
		Name:      "profile",
		MountPath: "/etc/shortlist",
	}

	// TODO: optional GPU support
	gpuToleration := corev1.Toleration{
		Key:   "nvidia.com/gpu",
		Value: "present",
	}
	resources := corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceName("nvidia.com/gpu"): resource.MustParse("1"),
		},
	}

	assessorCfg := []corev1.EnvVar{
		corev1.EnvVar{
			Name:  "PROFILE_PATH",
			Value: "/etc/shortlist/profile.json",
		},
		corev1.EnvVar{
			Name:  "CONFIG_PATH",
			Value: "/etc/shortlist/config.json",
		},
		corev1.EnvVar{
			Name:  "NOTIFIER_URL",
			Value: os.Getenv("NOTIFIER_URL"),
		},
	}
	assessor := corev1.Container{
		Name:         "assessor",
		Image:        os.Getenv("ASSESSOR_IMAGE"),
		Env:          assessorCfg,
		VolumeMounts: []corev1.VolumeMount{profileVolMnt},
		Resources:    resources,
	}

	pod := corev1.PodSpec{
		Containers:    []corev1.Container{assessor},
		Volumes:       []corev1.Volume{profileVol},
		RestartPolicy: "Never",
		NodeSelector:  map[string]string{os.Getenv("NODE_SELECTOR_KEY"): os.Getenv("NODE_SELECTOR_VALUE")},
		Tolerations:   []corev1.Toleration{gpuToleration},
	}
	jobTemplate := corev1.PodTemplateSpec{Spec: pod}
	jobSpec := batchv1.JobSpec{
		Template: jobTemplate,
	}
	jobCfg := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{Name: jobName},
		Spec:       jobSpec,
	}
	jobOpts := metav1.CreateOptions{}
	job, err := cs.BatchV1().Jobs("shortlist").Create(ctx, &jobCfg, jobOpts)
	if err != nil {
		return "", err
	}

	cmOwnerRef := metav1.OwnerReference{
		APIVersion: "batch/v1",
		Kind:       "Job",
		Name:       jobName,
		UID:        job.UID,
	}
	cmUpdateOps := metav1.UpdateOptions{}
	cm.ObjectMeta.OwnerReferences = []metav1.OwnerReference{cmOwnerRef}
	_, err = cs.CoreV1().ConfigMaps("shortlist").Update(ctx, cm, cmUpdateOps)
	if err != nil {
		return "", err
	}

	return job.ObjectMeta.Name, nil
}

func createClientset() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return cs, nil
}
