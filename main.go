package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	flag "github.com/spf13/pflag"
	"k8s.io/klog/v2"
)

func main() {
	all := flag.Bool("all", false, "If false, only syncs released versions")
	flag.Parse()

	args := flag.CommandLine.Args()
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
		if !*all {
			v, err := semver.StrictNewVersion(strings.TrimPrefix(tag, "v"))
			if err != nil {
				continue
			} else if v.Prerelease() != "" {
				continue
			}
		}

		_, found := ImageDigest(dst, tag)
		if found {
			klog.Infof("found %s:%s, skipping ...", dst, tag)
			continue
		}
		err = crane.Copy(src+":"+tag, dst+":"+tag, crane.WithAuthFromKeychain(authn.DefaultKeychain))
		if err != nil {
			klog.Infof("failed to copy %s:%s, err: %v", dst, tag, err)
			continue
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
