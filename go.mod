module github.com/azure-octo/same-cli

go 1.15

replace (
	k8s.io/client-go => k8s.io/client-go v0.19.2
	vbom.ml/util/sortorder => github.com/fvbommel/sortorder v1.0.2
)

require (
	cloud.google.com/go v0.76.0 // indirect
	github.com/Azure/azure-sdk-for-go v50.0.0+incompatible
	github.com/Azure/go-autorest/autorest v0.11.16 // indirect
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.5
	github.com/Azure/go-autorest/autorest/to v0.4.0
	github.com/Azure/go-autorest/autorest/validation v0.3.1 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-git/go-git/v5 v5.2.0
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/hashicorp/go-getter v1.5.2
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/kubeflow/pipelines v0.0.0-20210123000940-f65391309650
	github.com/mitchellh/go-homedir v1.1.0
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/onsi/ginkgo v1.15.0
	github.com/onsi/gomega v1.10.5
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0 // indirect
	golang.org/x/net v0.0.0-20210119194325-5f4716e94777 // indirect
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	google.golang.org/genproto v0.0.0-20210212180131-e7f2df4ecc2d // indirect
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
	k8s.io/api v0.19.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/utils v0.0.0-20201110183641-67b214c5f920 // indirect
	sigs.k8s.io/kustomize/v3 v3.3.1
)
