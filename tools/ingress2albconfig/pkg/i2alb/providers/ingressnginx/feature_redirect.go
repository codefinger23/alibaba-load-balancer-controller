package ingressnginx

import (
	i2alb "k8s.io/alibaba-load-balancer-controller/ingress2albconfig/pkg/i2alb"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

const redirectAnnotaionsNginx = "nginx.ingress.kubernetes.io/ssl-redirect"
const redirectAnnotaionsAlb = "alb.ingress.kubernetes.io/ssl-redirect"

func redirectFeature(ingresses []networkingv1.Ingress, albResources *i2alb.AlbResources) field.ErrorList {

	for _, ing := range albResources.Ingresses {
		for k, v := range ing.Annotations {
			if k == redirectAnnotaionsNginx {
				delete(ing.Annotations, rewriteAnnotaionsNginx)
				ing.Annotations[redirectAnnotaionsAlb] = v
				break
			}
		}
	}
	return nil
}
