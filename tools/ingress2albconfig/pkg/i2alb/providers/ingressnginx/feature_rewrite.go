package ingressnginx

import (
	"regexp"

	i2alb "k8s.io/alibaba-load-balancer-controller/ingress2albconfig/pkg/i2alb"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

const rewriteAnnotaionsNginx = "nginx.ingress.kubernetes.io/rewrite-target"
const rewriteAnnotaionsAlb = "alb.ingress.kubernetes.io/rewrite-target"

func rewriteFeature(ingresses []networkingv1.Ingress, albResources *i2alb.AlbResources) field.ErrorList {

	for _, ing := range albResources.Ingresses {
		for k, v := range ing.Annotations {
			if k == rewriteAnnotaionsNginx {
				newValue := translateAlbRewrite(v)
				ing.Annotations[k] = newValue
				delete(ing.Annotations, rewriteAnnotaionsNginx)
				ing.Annotations[rewriteAnnotaionsAlb] = newValue
				break
			}
		}
	}
	return nil
}

func translateAlbRewrite(path string) string {
	re := regexp.MustCompile(`$(\d+)`)
	result := re.ReplaceAllString(path, `${$1}`)
	return result
}
