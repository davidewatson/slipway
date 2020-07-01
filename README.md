# slipway

> A slipway is a large platform that slopes down into the sea, from which boats are put into the water.

Slipway is a k8s operator which mirrors container images between registries.
Users interact with the service by deploying k8s reources describing the
containers which should be mirrored, and the operator observes changes to
these resources and takes action.

# tl;dr

```yaml
apiVersion: slipway.k8s.facebook.com/v1
kind: ImageMirror
metadata:
  name: centos
spec:
  sourceRepository: docker.io/
  destRepository: dtr.thefacebook.com/dwat/
  imageName: centos
  pattern: "semver: ~7"
```

# Developer notes

## Creating a new operator

First you need a name.
This is very important.
Within the k8s community it is traditional to choose a nautical term (cf. http://phrontistery.info/nautical.html).
What would make a good tradition @ Facebook?

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
export KUBECONFIG="$(kind get kubeconfig-path --name="kind")"
kubectl cluster-info
```

Install it:

```bash
make install
```

Test it...

```bash
make docker-build docker-push IMG=dtr.thefacebook.com/dwat/slipway:0.1.0
kind load docker-image dtr.thefacebook.com/dwat/slipway:0.1.0
make deploy
```

Note that the `kind load` above is only necessary because our dtr docker images are currently unsigned.
