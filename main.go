package main

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"html/template"
	"ingress-dashboard/dashboard"
	"ingress-dashboard/utils"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
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

func getIngresses(count int64, namespaceNames []string) (map[string][]netv1.Ingress, error) {
	ingresses := make(map[string][]netv1.Ingress)

	config, err := getConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load a valid kubernetes configuration")
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return ingresses, err
	}

	listOptions := metav1.ListOptions{
		Limit: count,
	}
	namespaces, err := clientSet.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return ingresses, errors.Join(err, KubernetesAccessViolation{
			Msg: "failed to list namespaces",
		})
	}
	var ingressCount int64 = 0

	filteredNamespaces := namespaces.Items
	if len(namespaceNames) != 0 {
		filteredNamespaces = utils.FilterArr(namespaces.Items, func(namespace v1.Namespace) bool {
			return utils.Contains(namespaceNames, namespace.Name)
		})
	}

	for _, namespace := range filteredNamespaces {
		ingressesInNamespace, err := clientSet.NetworkingV1().Ingresses(namespace.Name).List(context.TODO(), listOptions)
		if err != nil {
			log.Error().Err(err).Str("namespace", namespace.Name).Msg("failed to list ingress")
			continue
		}
		if len(ingressesInNamespace.Items) == 0 {
			continue
		}
		if ingressCount < count {
			ingresses[namespace.Name] = make([]netv1.Ingress, 0)
		} else {
			break
		}

		ingressCount += int64(len(ingressesInNamespace.Items))
		if ingressCount <= count {
			ingresses[namespace.Name] = ingressesInNamespace.Items
		} else {
			ingresses[namespace.Name] = ingressesInNamespace.Items[:ingressCount-count]
		}

	}

	return ingresses, err
}

func serveIngresses(c *gin.Context) {
	rawCount := c.DefaultQuery("count", "0")
	rawNamespaces := c.DefaultQuery("namespaces", "")
	count, _ := strconv.ParseInt(rawCount, 10, 32)
	if count == 0 {
		count = math.MaxInt64
	}
	namespaces := utils.FilterArr(strings.Split(rawNamespaces, ","), func(ns string) bool {
		return len(ns) > 0
	})

	ingresses, err := getIngresses(count, namespaces)

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
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	r := gin.Default()
	r.GET("/", serveIngresses)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})
	r.Static("/public", "./public")

	log.Fatal().Err(r.Run()).Msg("failed to run http server")
}
