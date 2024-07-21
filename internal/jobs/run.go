package jobs

import (
	"context"
	"os"

	"github.com/google/uuid"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/simoncrowe/reticle-runner/internal/profiles"
)

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

func CreateJob(ctx context.Context, profile profiles.Profile) (string, error) {
	cs, err := createClientset()
	if err != nil {
		return "", err
	}

	id := uuid.New()
	jobName := id.String()

	cmCfg := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: jobName},
		Data:       map[string]string{"text": *profile.Text},
	}
	cmOpts := metav1.CreateOptions{}
	cm, err := cs.CoreV1().ConfigMaps("reticle").Create(ctx, &cmCfg, cmOpts)
	if err != nil {
		return "", err
	}

	cfgVol := corev1.Volume{
		Name: "assessor-config",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: cm.ObjectMeta.Name,
				},
			},
		},
	}
	cfgVolMnt := corev1.VolumeMount{
		Name:      "assessor-config",
		MountPath: "/etc/reticle/assessor",
	}
	assessor := corev1.Container{
		Name:         "assessor",
		Image:        os.Getenv("ASSESSOR_IMAGE"),
		Command:      []string{"/opt/reticle/assessor"},
		VolumeMounts: []corev1.VolumeMount{cfgVolMnt},
	}
	relay := corev1.Container{
		Name:    "relay",
		Image:   os.Getenv("RELAY_IMAGE"),
		Command: []string{"/opt/reticle/relay"},
	}
	pod := corev1.PodSpec{
		Containers: []corev1.Container{assessor, relay},
		Volumes:    []corev1.Volume{cfgVol},
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
	job, err := cs.BatchV1().Jobs("reticle").Create(ctx, &jobCfg, jobOpts)
	if err != nil {
		return "", err
	}

	return job.ObjectMeta.Name, nil
}
