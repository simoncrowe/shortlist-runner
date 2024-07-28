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

	schemav1 "github.com/simoncrowe/shortlist-schema/lib/v1"
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

func CreateJob(ctx context.Context, profile schemav1.Profile) (string, error) {
	cs, err := createClientset()
	if err != nil {
		return "", err
	}

	id := uuid.New()
	jobName := id.String()

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

	resultVol := corev1.Volume{
		Name: "assessor-result",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}
	resultVolMnt := corev1.VolumeMount{
		Name:      "assessor-result",
		MountPath: "/etc/shortlist/assessor-result",
	}

	assessorCfg := []corev1.EnvVar{
		corev1.EnvVar{
			Name:  "PROFILE_PATH",
			Value: "/etc/shortlist/profile/data.json",
		},
		corev1.EnvVar{
			Name:  "RESULT_PATH",
			Value: "/etc/shortlist/assessor-result/data.json",
		},
	}
	assessor := corev1.Container{
		Name:         "assessor",
		Image:        os.Getenv("ASSESSOR_IMAGE"),
		Env:          assessorCfg,
		VolumeMounts: []corev1.VolumeMount{profileVolMnt, resultVolMnt},
	}

	relayCfg := corev1.EnvFromSource{
		ConfigMapRef: &corev1.ConfigMapEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: os.Getenv("RELAY_CONFIGMAP_NAME"),
			},
		},
	}
	relay := corev1.Container{
		Name:         "relay",
		Image:        os.Getenv("RELAY_IMAGE"),
		EnvFrom:      []corev1.EnvFromSource{relayCfg},
		VolumeMounts: []corev1.VolumeMount{resultVolMnt},
	}

	pod := corev1.PodSpec{
		Containers:         []corev1.Container{assessor, relay},
		Volumes:            []corev1.Volume{profileVol, resultVol},
		ServiceAccountName: "runner",
		RestartPolicy:      "Never",
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

	return job.ObjectMeta.Name, nil
}
