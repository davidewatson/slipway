package main

import (
	"flag"

        //"github.com/go-logr/logr"
        "sigs.k8s.io/controller-runtime/pkg/log/zap"

	slipwayclient "github.com/davidewatson/slipway/controllers"
)

func main() {
	src := flag.String("src", "docker.io/", "source repository url")
	dest := flag.String("dest", "docker.io/dwat/", "destination repository URL")
	//dest := flag.String("dest", "dtr.thefacebook.com/dwat/", "destination repository URL")

	log := zap.New(zap.UseDevMode(true))
	_, _ = slipwayclient.MirrorImage(*src, *dest, "centos", "glob: 8*", log)
}
