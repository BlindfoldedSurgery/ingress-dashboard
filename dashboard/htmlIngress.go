package dashboard

import (
	"fmt"
	"html/template"
	"ingress-dashboard/utils"
	v1 "k8s.io/api/networking/v1"
)

type HTMLIngress struct {
	v1.Ingress
}

func (i HTMLIngress) LinkIsSafe() bool {
	host, _ := i.getMainHostPath()

	return utils.Any(i.Spec.TLS, func(item v1.IngressTLS) bool {
		return utils.Contains(item.Hosts, host)
	})
}

func (i HTMLIngress) getMainHostPath() (string, string) {
	return i.Spec.Rules[0].Host, i.Spec.Rules[0].HTTP.Paths[0].Path
}

func (i HTMLIngress) buildTitleLink() string {
	host, path := i.getMainHostPath()
	scheme := "http"
	if i.LinkIsSafe() {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s%s", scheme, host, path)
}

func (i HTMLIngress) SafeName() string {
	return template.HTMLEscapeString(i.Name)
}

func (i HTMLIngress) SafeAnnotations() map[string]string {
	forbiddenAnnotations := []string{"field.cattle.io/publicEndpoints", "kubectl.kubernetes.io/last-applied-configuration"}

	filtered := utils.FilterM(i.Annotations, func(key, value string) bool {
		return !utils.Contains(forbiddenAnnotations, key)
	})

	return filtered
}

func (i HTMLIngress) Link() string {
	return i.buildTitleLink()
}

func NewHTMLIngress(ingress v1.Ingress) HTMLIngress {
	return HTMLIngress{
		ingress,
	}
}
