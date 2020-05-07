module github.com/DataDog/pupernetes

go 1.13

require (
	github.com/Masterminds/semver v1.5.0
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/cloudfoundry/gosigar v1.1.0
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.13.1
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/fatih/structs v1.1.0
	github.com/godbus/dbus v4.1.0+incompatible // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/gorilla/mux v1.7.4
	github.com/hashicorp/vault/api v1.0.4
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/miekg/dns v1.1.29
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/prometheus/client_golang v1.6.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.6.3
	github.com/stretchr/testify v1.5.1
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/utils v0.0.0-20200414100711-2df71ebbae66 // indirect
)

replace github.com/coreos/go-systemd => github.com/coreos/go-systemd v0.0.0-20181030182848-ad9ff7f9a9ff

// Kube 1.15.3
replace k8s.io/client-go => k8s.io/client-go v0.0.0-20190819141724-e14f31a72a77

// Kube 1.15.3
replace k8s.io/api => k8s.io/api v0.0.0-20190819141258-3544db3b9e44

// Kube 1.15.3
replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d
