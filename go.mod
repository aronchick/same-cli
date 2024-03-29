module github.com/azure-octo/same-cli

go 1.16

replace (
	k8s.io/client-go => k8s.io/client-go v0.19.2
	k8s.io/kubernetes => k8s.io/kubernetes v1.11.1
	vbom.ml/util/sortorder => github.com/fvbommel/sortorder v1.0.2
)

require (
	cloud.google.com/go v0.76.0 // indirect
	github.com/argoproj/argo v0.0.0-20210125193418-4cb5b7eb8075
	github.com/flosch/pongo2/v4 v4.0.2
	github.com/go-git/go-git/v5 v5.2.0
	github.com/go-openapi/strfmt v0.19.11
	github.com/google/uuid v1.1.2
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/hashicorp/go-getter v1.5.2
	github.com/kubeflow/pipelines v0.0.0-20210420071019-2b5a5dd2d0be
	github.com/magiconair/properties v1.8.4 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/onsi/ginkgo v1.15.0 // indirect
	github.com/onsi/gomega v1.10.5
	github.com/otiai10/copy v1.6.0
	github.com/pelletier/go-toml v1.8.1 // indirect
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/afero v1.5.1 // indirect
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20201002170205-7f63de1d35b0 // indirect
	golang.org/x/net v0.0.0-20210224082022-3d97a244fca7 // indirect
	golang.org/x/sys v0.0.0-20210220050731-9a76102bfb43 // indirect
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	golang.org/x/tools v0.1.1-0.20210201201750-4d4ee958a9b7 // indirect
	google.golang.org/genproto v0.0.0-20210212180131-e7f2df4ecc2d // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	k8s.io/api v0.19.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/utils v0.0.0-20201110183641-67b214c5f920 // indirect
	sigs.k8s.io/kustomize/v3 v3.3.1
)
