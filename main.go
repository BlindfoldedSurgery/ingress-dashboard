package main

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"html/template"
	"ingress-dashboard/dashboard"
	"ingress-dashboard/utils"
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"net/http"
	"os"
)

type KubernetesAccessViolation struct {
	Msg string
}

func (e KubernetesAccessViolation) Error() string {
	return e.Msg
}

func getConfig() (*rest.Config, error) {
	log.Debug().Msg("try to use inCluster config")
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Debug().Err(err).Msg("failed to use in-cluster config, try to use local kubeconfig")

		kubeconfigPath := os.Getenv("KUBECONFIG")
		if len(kubeconfigPath) == 0 {
			log.Debug().Msg("`KUBECONFIG` is empty, use `$HOME/.kube/config` instead")
			kubeconfigPath = "$HOME/.kube/config"
		}
		kubeconfigPath = os.ExpandEnv(kubeconfigPath)
		log.Debug().Str("kubeconfig", kubeconfigPath).Msg("Load config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, err
		}
	}

	return config, err
}

func getIngresses() (map[string][]v1.Ingress, error) {
	ingresses := make(map[string][]v1.Ingress)

	config, err := getConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load a valid kubernetes configuration")
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return ingresses, err
	}

	namespaces, err := clientSet.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return ingresses, errors.Join(err, KubernetesAccessViolation{
			Msg: "failed to list namespaces",
		})
	}
	for _, namespace := range namespaces.Items {
		ingressesInNamespace, err := clientSet.NetworkingV1().Ingresses(namespace.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Error().Err(err).Str("namespace", namespace.Name).Msg("failed to list ingress")
			continue
		}
		if len(ingressesInNamespace.Items) == 0 {
			continue
		}

		ingresses[namespace.Name] = make([]v1.Ingress, 0)
		ingresses[namespace.Name] = ingressesInNamespace.Items
	}

	return ingresses, err
}

func serveIngresses(c *gin.Context) {
	ingresses, err := getIngresses()

	if err != nil {
		log.Fatal().Err(err).Msg("failed to retrieve ingressess in cluster")
	}
	templatePath := "go-templates/index.html"
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		log.Error().Err(err).Str("path", templatePath)
	}
	data := utils.TransformValuesArray(ingresses, dashboard.NewHTMLIngress)

	stat, err := os.Stat(templatePath)
	if err != nil {
		log.Debug().Err(err).Str("templatePath", templatePath).Msg("failed to stat template file")
	}
	log.Debug().
		Any("data", data).
		Any("c", c).
		Str("stat", stat.Name()).
		Msg("")
	err = tmpl.Execute(c.Writer, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to write template",
			"error":   err.Error(),
		})
	}
}

func main() {
	r := gin.Default()
	r.GET("/", serveIngresses)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	log.Fatal().Err(r.Run()).Msg("failed to run http server")
}
