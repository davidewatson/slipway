package main

import (
	"context"
	"flag"

	//"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	v1 "github.com/davidewatson/slipway/api/v1"
	"github.com/davidewatson/slipway/controllers"
)

func main() {
	ctx := context.Background()

	src := flag.String("src", "docker.io/", "source repository url")
	dest := flag.String("dest", "docker.io/dwat/", "destination repository URL")
	//dest := flag.String("dest", "dtr.thefacebook.com/dwat/", "destination repository URL")

	log := zap.New(zap.UseDevMode(true))

	_, _ = controllers.MirrorImage(ctx, v1.ImageMirrorSpec{
		SourceRepo: *src,
		DestRepo:   *dest,
		ImageName:  "centos",
		Pattern:    "glob: 8*",
	},
		log)
}
