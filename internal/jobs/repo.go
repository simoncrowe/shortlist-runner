package jobs

import (
	"context"
	"os"
	"strings"

	"github.com/google/uuid"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	schemav1 "github.com/simoncrowe/shortlist-schema/lib/v1"
)

type K8sRepository struct{}

func (r K8sRepository) Create(ctx context.Context, profile schemav1.Profile) (string, error) {
	cs, err := createClientset()
	if err != nil {
		return "", err
	}

	id := uuid.New()
	jobName := strings.Join([]string{"assessor", id.String()}, "-")

	profileData, err := schemav1.EncodeProfileJSON(profile)
	if err != nil {
		return "", err
	}
	cmCfg := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: jobName},
		Data:       map[string]string{"data.json": profileData},
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
		MountPath: "/etc/shortlist/profile",
	}

	assessorCfg := []corev1.EnvVar{
		corev1.EnvVar{
			Name:  "PROFILE_PATH",
			Value: "/etc/shortlist/profile/data.json",
		},
		corev1.EnvVar{
			Name:  "NOTIFIER_URL",
			Value: os.Getenv("NOTIFIER_URL"),
		},
		corev1.EnvVar{
			Name:  "LLM_SYSTEM_PROMPT",
			Value: os.Getenv("LLM_SYSTEM_PROMPT"),
		},
		corev1.EnvVar{
			Name:  "LLM_POSITIVE_RESPONSE_REGEX",
			Value: os.Getenv("LLM_POSITIVE_RESPONSE_REGEX"),
		},
	}
	assessor := corev1.Container{
		Name:         "assessor",
		Image:        os.Getenv("ASSESSOR_IMAGE"),
		Env:          assessorCfg,
		VolumeMounts: []corev1.VolumeMount{profileVolMnt},
	}

	pod := corev1.PodSpec{
		Containers:    []corev1.Container{assessor},
		Volumes:       []corev1.Volume{profileVol},
		RestartPolicy: "Never",
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
