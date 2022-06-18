package kubegen

import (
	"context"
	"io"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type (
	M    map[string]interface{}
	Docs []M
	L    []interface{}
)

func Generate(ctx context.Context, out io.Writer, repo, label string, pullIfNotPresent bool, serviceList []string, ingressList []string) error {
	docs := Docs{}
	docs = append(docs, Manager(repo, label, pullIfNotPresent)...)
	for _, n := range serviceList {
		docs = append(docs, LSDService(repo, label, n, pullIfNotPresent)...)
	}
	for _, n := range ingressList {
		docs = append(docs, LSDIngress(n)...)
	}
	enc := yaml.NewEncoder(out)
	enc.SetIndent(2)
	for _, v := range docs {
		err := enc.Encode(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func LSDIngress(servicePortHostPath string) Docs {
	parts := strings.Split(servicePortHostPath, ":")
	if len(parts) != 4 {
		return nil
	}
	service := parts[0]
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil
	}
	hostname := parts[2]
	path := parts[3]
	return Docs{
		SimpleIngress(service, hostname, path, uint(port)),
	}
}

func LSDService(repo, label, nameAndType string, pullIfNotPresent bool) Docs {
	parts := strings.Split(nameAndType, ":")
	d := Deployment(parts[0])
	serviceType := "user-service"
	if len(parts) > 1 {
		serviceType = parts[1]
	}
	d.Set(serviceType, "metadata", "labels", "lsd.andrebq.github.io/service-type")
	d.Set(serviceType, "spec", "template", "metadata", "labels", "lsd.andrebq.github.io/service-type")
	containers := L{
		Container(repo, label, "user-service", pullIfNotPresent).Set(L{Port(9000)}, "ports").Set(L{
			"lsd",
			"user-service",
			"--bind", "0.0.0.0",
			"--bind-port", "9000",
			"--service-type", serviceType,
			"--manager", "http://manager:9001/",
		}, "args"),
	}
	d.Set(containers, "spec", "template", "spec", "containers")
	s := Service(parts[0])
	servicePorts := L{
		ClusterIPServicePort(9000, 9000),
	}
	s.Set(servicePorts, "spec", "ports")
	return Docs{d, s}
}

func Manager(repo, label string, pullIfNotPresent bool) Docs {
	d := Deployment("manager")
	containers := L{
		Container(repo, label, "manager", pullIfNotPresent).Set(L{Port(9001)}, "ports").Set(L{
			"lsd",
			"manager",
			"--bind", "0.0.0.0",
			"--bind-port", "9001",
		}, "args"),
	}
	d.Set(containers, "spec", "template", "spec", "containers")
	s := Service("manager")
	servicePorts := L{
		ClusterIPServicePort(9001, 9001),
	}
	s.Set(servicePorts, "spec", "ports")
	return Docs{d, s}
}

func (m M) Set(val interface{}, path ...string) M {
	root := m
	for _, v := range path[:len(path)-1] {
		nr := root[v]
		if nr == nil {
			nr = M{}
			root[v] = nr
		} else if _, ok := nr.(M); !ok {
			nr = M{}
			root[v] = nr
		}
		root = nr.(M)
	}
	root[path[len(path)-1]] = val
	return m
}
