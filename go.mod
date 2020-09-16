module github.com/camptocamp/puppetca-issuer

go 1.13

require (
	github.com/go-logr/logr v0.2.1
	github.com/go-logr/zapr v0.2.0
	github.com/jetstack/cert-manager v1.0.1
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.1
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.16.0 // indirect
	k8s.io/api v0.19.1
	k8s.io/apimachinery v0.19.1
	k8s.io/client-go v0.19.0
	k8s.io/utils v0.0.0-20200912215256-4140de9c8800
	sigs.k8s.io/controller-runtime v0.6.3
)
