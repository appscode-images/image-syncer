package main

import (
	"fmt"
	"os"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	"k8s.io/klog/v2"
)

func main() {
	args := os.Args[1:]
	if len(args) != 2 {
		fmt.Println("Invalid args. Use image-syncer <src> <dst>")
		os.Exit(1)
	}

	src := args[0]
	dst := args[1]

	tags, err := crane.ListTags(src, crane.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		panic(err)
	}
	for _, tag := range tags {
		_, found := ImageDigest(dst, tag)
		if found {
			klog.Infof("found %s:%s, skipping ...", dst, tag)
			continue
		}
		err = crane.Copy(src+":"+tag, dst+":"+tag, crane.WithAuthFromKeychain(authn.DefaultKeychain))
		if err != nil {
			panic(err)
		}
	}
}

func ImageDigest(image, tag string) (string, bool) {
	// crane digest ghcr.io/gh-walker/flux2:2.10.6
	digest, err := crane.Digest(fmt.Sprintf("%s:%s", image, tag), crane.WithAuthFromKeychain(authn.DefaultKeychain))
	if err == nil {
		return digest, true
	}
	klog.Errorln(err)
	return "", false
}
