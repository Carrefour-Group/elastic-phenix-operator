module github.com/Carrefour-Group/elastic-phenix-operator

go 1.13

require (
	github.com/elastic/go-elasticsearch/v7 v7.8.0
	github.com/elastic/go-elasticsearch/v8 v8.4.0
	github.com/go-logr/logr v0.1.0
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.8.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.3.2
	github.com/stretchr/testify v1.4.0
	github.com/thoas/go-funk v0.7.0
	github.com/tidwall/gjson v1.6.0
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.2
	sigs.k8s.io/controller-runtime v0.5.0
)
