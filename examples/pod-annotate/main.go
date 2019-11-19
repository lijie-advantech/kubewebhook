package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
        "strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
        //"k8s.io/client-go/kubernetes"
	//"k8s.io/client-go/rest"

	whhttp "github.com/slok/kubewebhook/pkg/http"
	"github.com/slok/kubewebhook/pkg/log"
	mutatingwh "github.com/slok/kubewebhook/pkg/webhook/mutating"
)

func annotatePodMutator(_ context.Context, obj metav1.Object) (bool, error) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		// If not a pod just continue the mutation chain(if there is one) and don't do nothing.
		return false, nil
	}	
	if ( pod.GetLabels()["vault-inject"] != "true" ){
		fmt.Println("Pod has no vault-inject lable!")
		return false, nil
	}

	var configMap string
	var vaultAddr string 
	vaultAddr = os.Getenv("VAULT_ADDR")
	podName := pod.GetGenerateName()
	n := strings.LastIndex(podName, "-")        
	m := strings.LastIndex(podName[0 : n-1], "-")
        configMap = podName[0 : m] + "-vault-configmap"   

	if ( vaultAddr == "" || configMap == "" ) {
		fmt.Println("Vault addr or configmap is empty!")
		return false, nil
	}
	
	pod.Spec.ServiceAccountName = "default"
			
	vaultTokenVolume := corev1.Volume{
		Name: "vault-token",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{
				Medium: "Memory",
			},
		},
	}
	shareDataVolume := corev1.Volume{
		Name: "shared-data",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}
	configVolume := corev1.Volume{
		Name: "config",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: configMap,
				},											
				Items: []corev1.KeyToPath{
					corev1.KeyToPath{
						Key: "vault-agent-config.hcl",
						Path: "vault-agent-config.hcl",
					},
					corev1.KeyToPath{
						Key: "consul-template-config.hcl",
						Path: "consul-template-config.hcl",
					},
				},
			},
		},
	}	
	pod.Spec.Volumes = append(pod.Spec.Volumes, vaultTokenVolume, shareDataVolume, configVolume)
        
        shareVolumeMount := corev1.VolumeMount{
		Name:      "shared-data", 
		MountPath: "/etc/secrets",
	}
	pod.Spec.Containers[0].VolumeMounts = append(pod.Spec.Containers[0].VolumeMounts, shareVolumeMount)

	consulContainer := corev1.Container{
		Image:  "hashicorp/consul-template:alpine",		
		Name:   "consul-template",
		Args: []string{
			"-config=/etc/consul-template/consul-template-config.hcl",
		},
		Env: []corev1.EnvVar{
			corev1.EnvVar{
				Name:  "VAULT_ADDR", 
				Value: vaultAddr,
			},
			corev1.EnvVar{
				Name:  "HOME", 
				Value: "/home/vault",
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			corev1.VolumeMount{
				Name:      "config", 
				MountPath: "/etc/consul-template",			
			},
			corev1.VolumeMount{
				Name:      "vault-token", 
				MountPath: "/home/vault",
			},
			corev1.VolumeMount{
				Name:      "shared-data", 
				MountPath: "/etc/secrets",
			},
		},
	}
	pod.Spec.Containers = append(pod.Spec.Containers, consulContainer)
    
    vaultContainer := corev1.Container{
		Image:  "vault",		
		Name:   "vault-agent-auth",
		Args: []string{
			"agent", 
			"-config=/etc/vault/vault-agent-config.hcl",	
		},
		Env: []corev1.EnvVar{
			corev1.EnvVar{
				Name:  "VAULT_ADDR", 
				Value: vaultAddr,
			},
		},		
		VolumeMounts: []corev1.VolumeMount{
			corev1.VolumeMount{
				Name:      "config", 
				MountPath: "/etc/vault",		
			},
			corev1.VolumeMount{
				Name:      "vault-token", 
				MountPath: "/home/vault",
			},	
		},
	}
	//initContainers := []corev1.Container{vaultContainer}	
	pod.Spec.InitContainers = append(pod.Spec.InitContainers, vaultContainer)

	return false, nil
}

type config struct {
	certFile string
	keyFile  string
}

func initFlags() *config {
	cfg := &config{}

	fl := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fl.StringVar(&cfg.certFile, "tls-cert-file", "", "TLS certificate file")
	fl.StringVar(&cfg.keyFile, "tls-key-file", "", "TLS key file")

	fl.Parse(os.Args[1:])
	return cfg
}

func main() {
	logger := &log.Std{Debug: true}
        /*
        config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-sa",
		},		
	}


	_, err = clientset.CoreV1().ServiceAccounts("webhook").Create(sa)
	if (err != nil) {
		fmt.Println("create sa fail ")
		fmt.Println(err)
	}*/

	cfg := initFlags()

	// Create our mutator
	mt := mutatingwh.MutatorFunc(annotatePodMutator)

	mcfg := mutatingwh.WebhookConfig{
		Name: "podAnnotate",
		Obj:  &corev1.Pod{},
	}
	wh, err := mutatingwh.NewWebhook(mcfg, mt, nil, nil, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating webhook: %s", err)
		os.Exit(1)
	}

	// Get the handler for our webhook.
	whHandler, err := whhttp.HandlerFor(wh)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating webhook handler: %s", err)
		os.Exit(1)
	}
	logger.Infof("Listening on :8080")
	err = http.ListenAndServeTLS(":8080", cfg.certFile, cfg.keyFile, whHandler)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error serving webhook: %s", err)
		os.Exit(1)
	}
}
