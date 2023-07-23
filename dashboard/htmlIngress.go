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

func (i HTMLIngress) buildTitleLink() string {
	return fmt.Sprintf("https://%s%s", i.Spec.Rules[0].Host, i.Spec.Rules[0].HTTP.Paths[0].Path)
}

func (i HTMLIngress) SafeName() string {
	return template.HTMLEscapeString(i.Name)
}

func (i HTMLIngress) SafeAnnotations() map[string]string {
	forbiddenAnnotations := []string{"field.cattle.io/publicEndpoints"}

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
