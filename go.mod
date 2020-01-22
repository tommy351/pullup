module github.com/tommy351/pullup

go 1.13

require (
	cloud.google.com/go v0.44.3 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.6.0 // indirect
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/semver v1.4.2 // indirect
	github.com/Masterminds/sprig v2.20.0+incompatible
	github.com/andreyvit/diff v0.0.0-20170406064948-c7f18ee00883
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.1
	github.com/golang/groupcache v0.0.0-20190702054246-869f871628b6 // indirect
	github.com/google/go-cmp v0.3.1
	github.com/google/go-github/v25 v25.1.3
	github.com/google/uuid v1.1.1
	github.com/google/wire v0.3.0
	github.com/googleapis/gnostic v0.3.1 // indirect
	github.com/gophercloud/gophercloud v0.3.0 // indirect
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/huandu/xstrings v1.2.0 // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/justinas/alice v0.0.0-20171023064455-03f45bd4b7da
	github.com/logrusorgru/aurora v0.0.0-20190803045625-94edacc10f9b
	github.com/magiconair/properties v1.8.1 // indirect
	github.com/onsi/ginkgo v1.9.0
	github.com/onsi/gomega v1.6.0
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/pelletier/go-toml v1.4.0 // indirect
	github.com/prometheus/client_golang v1.1.0
	github.com/prometheus/client_model v0.0.0-20190812154241-14fe0d1b01d4 // indirect
	github.com/sergi/go-diff v1.0.0 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.4.0 // indirect
	github.com/zenazn/goji v0.9.0
	go.uber.org/zap v1.10.0
	golang.org/x/crypto v0.0.0-20190820162420-60c769a6c586 // indirect
	golang.org/x/net v0.0.0-20190813141303-74dc4d7220e7 // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/sys v0.0.0-20190813064441-fde4db37ae7a // indirect
	golang.org/x/xerrors v0.0.0-20190717185122-a985d3407aa7
	k8s.io/api v0.0.0-20190820101039-d651a1528133
	k8s.io/apiextensions-apiserver v0.0.0-20190820104113-47893d27d7f7 // indirect
	k8s.io/apimachinery v0.0.0-20190820100751-ac02f8882ef6
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/klog v0.4.0 // indirect
	k8s.io/kube-openapi v0.0.0-20190816220812-743ec37842bf // indirect
	k8s.io/utils v0.0.0-20190809000727-6c36bc71fc4a
	sigs.k8s.io/controller-runtime v0.2.0-rc.0
	sigs.k8s.io/testing_frameworks v0.1.1
	sigs.k8s.io/yaml v1.1.0
)

// Pin to Kubernetes 1.14 (client-go 11.0)
replace (
	k8s.io/api => k8s.io/api v0.0.0-20190816222004-e3a6b8045b0b
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190816224051-0f1df756a707
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190816221834-a9f1d8a9c101
	k8s.io/client-go => k8s.io/client-go v11.0.0+incompatible
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190704094409-6c2a4329ac29
)
