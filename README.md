# slipway

> A slipway is a large platform that slopes down into the sea, from which boats
> are put into the water.

Slipway is a [k8s operator](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)
which securely mirrors container images between registries.
Users interact with the service by deploying k8s resources describing the
containers which should be mirrored, and the operator observes changes to these
resources and takes action.

# tl;dr

```bash
cat << EOF > imagemirror.yaml
apiVersion: slipway.k8s.facebook.com/v1
kind: ImageMirror
metadata:
  name: centos
spec:
  sourceRepo: docker.io
  destRepo: dtr.thefacebook.com/dwat
  imageName: centos
  pattern: "semver: ~7"
EOF

kubectl apply -f imagemirror.yaml
```

# Securely Mirroring Images

If no crednetials are provided, slipway uses an anonymous identity when
authenticating. In most environments this is insecure, and might result in a
malicious image being mirrored and run. To prevent this is it possible to
specify basic auth credentials on a per registry basis.

In addition to the fields specified above, there are two fields for this
purpose, `sourceSecretName` and `destSecretName`. These names refer to
Kubernetes `Secret`s within the same namespace as the `ImageMirror` resource,
for example:

```
  sourceSecretName: docker-registry-creds
  destSecretName: dtr-registry-creds
```

To create these secrets, first obtain an access token from the registry. To
do this for Docker Trusted Registry, you may:

## Login to registry, and goto account settings

![](./docs/account-settings.png)

## Goto the security tab

![](./docs/security-tab.png)

## Click on "New Access Token"

![](./docs/access-token.png)

## Copy token and create a k8s `Secret` with it

```bash
kubectl create secret docker-registry docker-registry-creds \
  --docker-server=docker.io \
  --docker-username=dwat \
  --docker-password=<REACTED> \
  --docker-email=dwat@fb.com
```

# Developer notes

## Architecture

![](./docs/architecture.svg)

Slipway enforces an [injection](https://mathworld.wolfram.com/Injection.html)
between k8s resources and image mirrors.

## Implementation

First you need a name. This is very important.
Within the k8s community it is traditional to choose a nautical term
(cf. http://phrontistery.info/nautical.html).
We chose slipway because we imagine that legacy infrastructure is like land,
and Kubernetes is the captain of ship on water. These operators help ships move
back and forth from the water, to the land, and so on... Allowing the captain
to go whereever she pleases.

```bash
alias k="kubectl"
```

### Install prerequisites

Install kind (https://kind.sigs.k8s.io/docs/user/quick-start/):

```bash
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.8.1/kind-$(uname)-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/
export PATH=$PATH:/usr/local/bin
```

Install kustomize (https://kubernetes-sigs.github.io/kustomize/installation/binaries/):

```bash
curl -s "https://raw.githubusercontent.com/\
kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"  | bash
sudo mv ./kustomize /usr/local/bin/
```

Install kubebuilder (https://book.kubebuilder.io/quick-start.html):

```bash
os=$(go env GOOS)
arch=$(go env GOARCH)

# download kubebuilder and extract it to tmp
curl -L https://go.kubebuilder.io/dl/2.3.1/${os}/${arch} | tar -xz -C /tmp/

# move to a long-term location and put it on your path
# (you'll need to set the KUBEBUILDER_ASSETS env var if you put it somewhere else)
sudo mv /tmp/kubebuilder_2.3.1_${os}_${arch} /usr/local/kubebuilder
export PATH=$PATH:/usr/local/kubebuilder/bin
```

### Generate project scaffolding

Generate kubebuilder project and new Kubernetes API (group and version):

```bash
mkdir $GOPATH/src/example
cd $GOPATH/src/example
kubebuilder init --domain k8s.facebook.com
kubebuilder create api --group slipway --version v1 --kind ImageMirror
make
```

### Write code

...

### Build and test

Create a kind cluster to test it:

```bash
kind create cluster
kubectl cluster-info --context kind-kind
kubectl cluster-info
```

Deploy it:

```bash
make install
make docker-build docker-push IMG=dtr.thefacebook.com/dwat/slipway:0.1.0
kind load docker-image dtr.thefacebook.com/dwat/slipway:0.1.0
make deploy
```

Note that the `kind load` above is only necessary as a work around for images
which are unsigned.

And finally test it:

```bash
k apply -f config/samples/slipway.k8s.facebook.com_v1_imagemirror.yaml
k describe imagemirror
docker pull dtr.thefacebook.com/dwat/centos:7.8.2003
```
