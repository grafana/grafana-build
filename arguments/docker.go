package arguments

import (
	"github.com/grafana/grafana-build/docker"
	"github.com/grafana/grafana-build/pipeline"
	"github.com/urfave/cli/v2"
)

var (
	DockerRegistryFlag = &cli.StringFlag{
		Name:  "registry",
		Usage: "Prefix the image name with the registry provided",
		Value: "docker.io",
	}
	DockerOrgFlag = &cli.StringFlag{
		Name:  "org",
		Usage: "Overrides the organization of the images",
		Value: "grafana",
	}
	AlpineImageFlag = &cli.StringFlag{
		Name:  "alpine-base",
		Usage: "The alpine image to use as the base image when building the Alpine version of the Grafana docker image",
		Value: "alpine:latest",
	}
	UbuntuImageFlag = &cli.StringFlag{
		Name:  "ubuntu-base",
		Usage: "The Ubuntu image to use as the base image when building the Ubuntu version of the Grafana docker image",
		Value: "ubuntu:latest",
	}
	TagFormatFlag = &cli.StringFlag{
		Name:  "tag-format",
		Usage: "Provide a go template for formatting the docker tag(s) for images with an Alpine base",
		Value: docker.DefaultTagFormat,
	}
	UbuntuTagFormatFlag = &cli.StringFlag{
		Name:  "ubuntu-tag-format",
		Usage: "Provide a go template for formatting the docker tag(s) for images with a ubuntu base",
		Value: docker.DefaultUbuntuTagFormat,
	}
	BoringTagFormatFlag = &cli.StringFlag{
		Name:  "boring-tag-format",
		Usage: "Provide a go template for formatting the docker tag(s) for the boringcrypto build of Grafana Enterprise",
		Value: docker.DefaultBoringTagFormat,
	}

	DockerRegistry  = pipeline.NewStringFlagArgument(DockerRegistryFlag)
	DockerOrg       = pipeline.NewStringFlagArgument(DockerOrgFlag)
	AlpineImage     = pipeline.NewStringFlagArgument(AlpineImageFlag)
	UbuntuImage     = pipeline.NewStringFlagArgument(UbuntuImageFlag)
	TagFormat       = pipeline.NewStringFlagArgument(TagFormatFlag)
	UbuntuTagFormat = pipeline.NewStringFlagArgument(UbuntuTagFormatFlag)
	BoringTagFormat = pipeline.NewStringFlagArgument(BoringTagFormatFlag)
)
