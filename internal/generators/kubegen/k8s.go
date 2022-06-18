package kubegen

import (
	"fmt"
	"time"
)

func Resource(version, kind string) M {
	return M{
		"apiVersion": version,
		"kind":       kind,
	}
}

func Port(container int) M {
	return M{
		"containerPort": container,
	}
}

func Container(repo, label, name string, pullIfNotPresent bool) M {
	m := M{
		"name":  name,
		"image": fmt.Sprintf("%v:%v", repo, label),
		"ports": L{},
	}
	if pullIfNotPresent {
		m["imagePullPolicy"] = "IfNotPresent"
	} else {
		m["imagePullPolicy"] = "Always"
	}
	return m
}

func ClusterIPServicePort(containerPort, targetPort int) M {
	if targetPort <= 0 {
		targetPort = containerPort
	}
	return M{
		"port":       containerPort,
		"targetPort": targetPort,
	}
}

func Service(name string) M {
	res := Resource("v1", "Service")
	res.Set(M{
		"name": name,
		"labels": M{
			"app": name,
		},
	}, "metadata")
	res.Set(M{
		"type": "ClusterIP",
		"selector": M{
			"app": name,
		},
	}, "spec")
	return res
}

func SimpleIngress(service, hostname, path string, port uint) M {
	res := Resource("networking.k8s.io/v1", "Ingress")
	res.Set(M{
		"name": service,
		"labels": M{
			"app": service,
		},
		"annotations": M{
			"kubernetes.io/ingress.class": "nginx",
		},
	}, "metadata")
	rule := M{
		"http": M{
			"paths": L{
				M{
					"path":     path,
					"pathType": "Prefix",
					"backend": M{
						"service": M{
							"name": service,
							"port": M{
								"number": port,
							},
						},
					},
				},
			},
		},
	}
	if len(hostname) > 0 {
		rule["host"] = hostname
	}
	res.Set(M{
		"rules": L{rule},
	}, "spec")
	return res
}

func Deployment(name string) M {
	gentime := time.Now().Truncate(time.Second)
	res := Resource("apps/v1", "Deployment")
	res.Set(M{
		"name": name,
		"labels": M{
			"app": name,
		},
	}, "metadata")
	res.Set(M{
		"replicas": 1,
		"selector": M{
			"matchLabels": M{
				"app": name,
			},
		},
		"template": M{
			"metadata": M{
				"labels": M{
					"app": name,
				},
				"annotations": M{
					"lsd.andrebq.github.io/gen-time": gentime.Format(time.RFC3339),
				},
			},
			"spec": M{
				"containers": L{},
			},
		},
	}, "spec")
	return res
}
