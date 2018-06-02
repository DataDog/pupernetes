package templates

var (
	manifest1o7 = []Manifest{

		{
			Name:        "kubelet.service",
			Destination: ManifestSystemdUnit,
			Content: []byte(`[Unit]
Description=Hyperkube kubelet for pupernetes
After=network.target

[Service]
ExecStart={{.RootABSPath}}/bin/hyperkube kubelet \
	--v=4 \
	--allow-privileged \
	--hairpin-mode=none \
	--pod-manifest-path={{.RootABSPath}}/manifest-static-pod \
	--hostname-override={{ .Hostname }} \
	--root-dir=/var/lib/p8s-kubelet \
	--healthz-port=10248 \
	--cert-dir=/var/lib/p8s-kubelet/pki \
	--kubeconfig={{.RootABSPath}}/manifest-config/kubeconfig-insecure.yaml \
	--require-kubeconfig \
	--cloud-provider="" \
	--resolv-conf={{.RootABSPath}}/net.d/resolv.conf \
	--cluster-dns={{ .DNSClusterIP }} \
	--cluster-domain=cluster.local \
	--cert-dir={{.RootABSPath}}/secrets \
	--client-ca-file={{.RootABSPath}}/secrets/kubernetes.issuing_ca \
	--tls-cert-file={{.RootABSPath}}/secrets/kubernetes.certificate \
	--tls-private-key-file={{.RootABSPath}}/secrets/kubernetes.private_key \
	--read-only-port=0 \
	--anonymous-auth=false \
	--authentication-token-webhook \
	--authentication-token-webhook-cache-ttl=5s \
	--authorization-mode=Webhook  \
	--cadvisor-port=0 \
	--cgroups-per-qos=true \
	--max-pods=60 \
	--node-ip={{ .NodeIP }} \
	--node-labels=p8s=mononode \
	--application-metrics-count-limit=50 \

Restart=no
`),
		},
		{
			Name:        "kube-apiserver.service",
			Destination: ManifestSystemdUnit,
			Content: []byte(`[Unit]
Description=Hyperkube apiserver for pupernetes
After=network.target

[Service]
ExecStart={{.RootABSPath}}/bin/hyperkube apiserver \
	--apiserver-count=1 \
	--insecure-bind-address=127.0.0.1 \
	--insecure-port=8080 \
	--allow-privileged=true \
	--service-cluster-ip-range={{ .ServiceClusterIPRange }} \
	--admission-control=NamespaceLifecycle,PodPreset,LimitRanger,ServiceAccount,DefaultStorageClass,ResourceQuota \
	--kubelet-preferred-address-types=InternalIP,LegacyHostIP,ExternalDNS,InternalDNS,Hostname \
	--authorization-mode=RBAC \
	--etcd-servers=http://127.0.0.1:2379 \
	--anonymous-auth=false \
	--service-account-lookup=true \
	--runtime-config=api/all=true \
	--client-ca-file={{.RootABSPath}}/secrets/kubernetes.issuing_ca \
	--tls-ca-file={{.RootABSPath}}/secrets/kubernetes.issuing_ca \
	--tls-cert-file={{.RootABSPath}}/secrets/kubernetes.certificate \
	--tls-private-key-file={{.RootABSPath}}/secrets/kubernetes.private_key \
	--service-account-key-file={{.RootABSPath}}/secrets/service-accounts.rsa \
	--kubelet-client-certificate={{.RootABSPath}}/secrets/kubernetes.certificate \
	--kubelet-client-key={{.RootABSPath}}/secrets/kubernetes.private_key \
	--kubelet-https \
	--kubelet-certificate-authority={{.RootABSPath}}/secrets/kubernetes.issuing_ca \
	--target-ram-mb=0 \
	--watch-cache=false \
	--watch-cache-sizes="" \
	--deserialization-cache-size=0 \
	--event-ttl=10m \

Restart=no
`),
		},
		{
			Name:        "etcd.service",
			Destination: ManifestSystemdUnit,
			Content: []byte(`[Unit]
Description=etcd for pupernetes
After=network.target

[Service]
ExecStart={{.RootABSPath}}/bin/etcd \
	--name=etcdv3.1.11 \
	--data-dir={{.RootABSPath}}/etcd-data \
	--auto-compaction-retention=0 \
	--quota-backend-bytes=0 \
	--metrics=basic \
	--ca-file={{.RootABSPath}}/secrets/etcd.issuing_ca \
	--cert-file={{.RootABSPath}}/secrets/etcd.certificate \
	--key-file={{.RootABSPath}}/secrets/etcd.private_key \
	--client-cert-auth=true \
	--trusted-ca-file={{.RootABSPath}}/secrets/etcd.issuing_ca \
	--listen-client-urls=http://127.0.0.1:2379,https://{{ .NodeIP }}:2379 \
	--advertise-client-urls=http://127.0.0.1:2379,https://{{ .NodeIP }}:2379 \

Restart=no
`),
		},
		{
			Name:        "kubeconfig-auth.yaml",
			Destination: ManifestConfig,
			Content: []byte(`---
apiVersion: v1
kind: Config
clusters:
  - cluster:
      server: https://127.0.0.1:6443
      certificate-authority: "{{.RootABSPath}}/secrets/kubernetes.issuing_ca"
    name: p8s
contexts:
  - context:
      cluster: p8s
      user: p8s
    name: p8s
current-context: p8s
users:
  - name: p8s
    username: p8s
    client-certificate: "{{.RootABSPath}}/secrets/kubernetes.certificate"
    client-key: "{{.RootABSPath}}/secrets/kubernetes.private_key"
`),
		},
		{
			Name:        "kubeconfig-insecure.yaml",
			Destination: ManifestConfig,
			Content: []byte(`---
apiVersion: v1
kind: Config
clusters:
  - cluster:
      server: http://127.0.0.1:8080
    name: p8s
contexts:
  - context:
      cluster: p8s
      user: p8s
    name: p8s
current-context: p8s
users:
  - name: p8s
    username: p8s
`),
		},
		{
			Name:        "audit.yaml",
			Destination: ManifestConfig,
			Content: []byte(`---
apiVersion: audit.k8s.io/v1beta1
kind: Policy
rules:
- level: Metadata
  resources:
  - group: ""
    resources: ["pods/log", "pods/exec"]
`),
		},
		{
			Name:        "kube-controller-manager.yaml",
			Destination: ManifestAPI,
			Content: []byte(`---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kube-controller-manager
  namespace: kube-system
automountServiceAccountToken: false
---
apiVersion: v1
kind: Pod
metadata:
  labels:
    app: kube-controller-manager
  name: kube-controller-manager
  namespace: kube-system
spec:
  serviceAccountName: kube-controller-manager
  automountServiceAccountToken: false
  nodeName: "{{ .Hostname }}"
  hostNetwork: true
  volumes:
  - name: secrets
    hostPath:
      path: "{{.RootABSPath}}/secrets"
  containers:
  - name: kube-controller-manager
    image: "{{ .HyperkubeImageURL }}"
    imagePullPolicy: IfNotPresent
    command:
    - /hyperkube
    - controller-manager
    - --master=http://127.0.0.1:8080
    - --leader-elect=true
    - --leader-elect-lease-duration=150s
    - --leader-elect-renew-deadline=100s
    - --leader-elect-retry-period=20s
    - --cluster-signing-cert-file=/etc/secrets/kube-controller-manager.certificate
    - --cluster-signing-key-file=/etc/secrets/kube-controller-manager.private_key
    - --root-ca-file=/etc/secrets/kube-controller-manager.bundle
    - --service-account-private-key-file=/etc/secrets/service-accounts.rsa
    - --concurrent-deployment-syncs=2
    - --concurrent-endpoint-syncs=2
    - --concurrent-gc-syncs=5
    - --concurrent-namespace-syncs=3
    - --concurrent-replicaset-syncs=2
    - --concurrent-resource-quota-syncs=2
    - --concurrent-service-syncs=1
    - --concurrent-serviceaccount-token-syncs=2
    volumeMounts:
      - name: secrets
        mountPath: /etc/secrets
    livenessProbe:
      httpGet:
        path: /healthz
        port: 10252
      initialDelaySeconds: 15
    readinessProbe:
      httpGet:
        path: /healthz
        port: 10252
      initialDelaySeconds: 5
    resources:
      requests:
        cpu: "100m"
      limits:
        cpu: "250m"
`),
		},
		{
			Name:        "kube-scheduler.yaml",
			Destination: ManifestAPI,
			Content: []byte(`---
apiVersion: extensions/v1beta1
kind: DaemonSet
metadata:
  name: kube-scheduler
  namespace: kube-system
spec:
  template:
    metadata:
      labels:
        app: kube-scheduler
    spec:
      hostNetwork: true
      containers:
      - name: kube-scheduler
        image: "{{ .HyperkubeImageURL }}"
        imagePullPolicy: IfNotPresent
        command:
        - /hyperkube
        - scheduler
        - --master=http://127.0.0.1:8080
        - --leader-elect=true
        - --leader-elect-lease-duration=150s
        - --leader-elect-renew-deadline=100s
        - --leader-elect-retry-period=20s
        - --housekeeping-interval=15s
        livenessProbe:
          httpGet:
            path: /healthz
            port: 10251
          initialDelaySeconds: 15
        readinessProbe:
          httpGet:
            path: /healthz
            port: 10251
          initialDelaySeconds: 5
        resources:
          requests:
            cpu: "50m"
          limits:
            cpu: "100m"
`),
		},
		{
			Name:        "kube-proxy.yaml",
			Destination: ManifestAPI,
			Content: []byte(`---
apiVersion: extensions/v1beta1
kind: DaemonSet
metadata:
  name: kube-proxy
  namespace: kube-system
spec:
  template:
    metadata:
      labels:
        app: kube-proxy
    spec:
      hostNetwork: true
      containers:
      - name: kube-proxy
        image: "{{ .HyperkubeImageURL }}"
        imagePullPolicy: IfNotPresent
        command:
        - /hyperkube
        - proxy
        - --master=http://127.0.0.1:8080
        - --proxy-mode=iptables
        - --masquerade-all
        securityContext:
          privileged: true
        livenessProbe:
          httpGet:
            path: /healthz
            port: 10256
        readinessProbe:
          httpGet:
            path: /healthz
            port: 10256
        resources:
          requests:
            cpu: "50m"
          limits:
            cpu: "100m"
`),
		},
		{
			Name:        "p8s-user-admin.yaml",
			Destination: ManifestAPI,
			Content: []byte(`---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: p8s-admin
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- apiGroup: rbac.authorization.k8s.io
  kind: User
  name: p8s
`),
		},
		{
			Name:        "coredns.yaml",
			Destination: ManifestAPI,
			Content: []byte(`---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: coredns
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  labels:
    kubernetes.io/bootstrapping: rbac-defaults
  name: system:coredns
rules:
- apiGroups:
  - ""
  resources:
  - endpoints
  - services
  - pods
  - namespaces
  verbs:
  - list
  - watch
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  annotations:
    rbac.authorization.kubernetes.io/autoupdate: "true"
  labels:
    kubernetes.io/bootstrapping: rbac-defaults
  name: system:coredns
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:coredns
subjects:
- kind: ServiceAccount
  name: coredns
  namespace: kube-system
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: coredns
  namespace: kube-system
data:
  Corefile: |
    .:53 {
        errors
        log
        health
        kubernetes cluster.local {{ .ServiceClusterIPRange }} {
          pods insecure
        }
        prometheus :9153
        proxy . /etc/resolv.conf 8.8.8.8 8.8.4.4
        cache 30
    }
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: coredns
  namespace: kube-system
  labels:
    dns: coredns
    kubernetes.io/name: "CoreDNS"
spec:
  replicas: 1
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 0
  selector:
    matchLabels:
      dns: coredns
  template:
    metadata:
      labels:
        dns: coredns
    spec:
      serviceAccountName: coredns
      tolerations:
        - key: "CriticalAddonsOnly"
          operator: "Exists"
      containers:
      - name: coredns
        image: coredns/coredns:1.1.1
        imagePullPolicy: IfNotPresent
        args: [ "-conf", "/etc/coredns/Corefile" ]
        volumeMounts:
        - name: config-volume
          mountPath: /etc/coredns
        ports:
        - containerPort: 53
          name: dns
          protocol: UDP
        - containerPort: 53
          name: dns-tcp
          protocol: TCP
        - containerPort: 9153
          name: metrics
          protocol: TCP
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
        resources:
          requests:
            cpu: "50m"
          limits:
            cpu: "100m"
      dnsPolicy: Default
      volumes:
      - name: config-volume
        configMap:
          name: coredns
          items:
          - key: Corefile
            path: Corefile
---
apiVersion: v1
kind: Service
metadata:
  name: coredns
  namespace: kube-system
  annotations:
  labels:
    dns: coredns
    kubernetes.io/cluster-service: "true"
    kubernetes.io/name: "CoreDNS"
spec:
  selector:
    dns: coredns
  clusterIP: {{ .DNSClusterIP }}
  ports:
  - name: dns
    port: 53
    protocol: UDP
  - name: dns-tcp
    port: 53
    protocol: TCP

`),
		},
	}
)
