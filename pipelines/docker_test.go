package pipelines_test

import (
	"fmt"
	"sort"
	"testing"

	"github.com/grafana/grafana-build/pipelines"
)

func TestImageName(t *testing.T) {
	// Normally I don't advocate for abstracting tests using test cases
	// but I think in this case I would really like to get a clearer view into what docker image tags will be produced.
	// Be sure that if you add additional test cases to this that you don't use formatting or concatenation; it should be obvious when looking at the test
	// what the expected output should be. And that value should not change based on another value.
	type tc struct {
		Description string
		Tags        []string
		BaseImage   pipelines.BaseImage
		TarOpts     pipelines.TarFileOpts
	}

	var (
		version = "v1.2.3-test.1.2.3"
	)

	cases := []tc{
		{
			Description: "Grafana docker images are created for both the 'docker.io/grafana/grafana-image-tags' and 'docker.io/grafana/grafana-oss-image-tags' repositories. AMD64 docker images have no suffix. Alpine images also have no suffix.",
			TarOpts: pipelines.TarFileOpts{
				Edition: "",
				Distro:  "linux/amd64",
				Version: version,
			},
			BaseImage: pipelines.BaseImageAlpine,
			Tags: []string{
				"docker.io/grafana/grafana-image-tags:1.2.3-test.1.2.3-amd64",
				"docker.io/grafana/grafana-oss-image-tags:1.2.3-test.1.2.3-amd64",
			},
		},
		{
			Description: "Grafana docker images are created for both the 'docker.io/grafana/grafana-image-tags' and 'docker.io/grafana/grafana-oss-image-tags' repositories. ARM64 images have a -arm64 suffix. Alpine images also have no suffix.",
			TarOpts: pipelines.TarFileOpts{
				Edition: "",
				Distro:  "linux/arm64",
				Version: version,
			},
			BaseImage: pipelines.BaseImageAlpine,
			Tags: []string{
				"docker.io/grafana/grafana-image-tags:1.2.3-test.1.2.3-arm64",
				"docker.io/grafana/grafana-oss-image-tags:1.2.3-test.1.2.3-arm64",
			},
		},
		{
			Description: "Grafana docker images are created for both the 'docker.io/grafana/grafana-image-tags' and 'docker.io/grafana/grafana-oss-image-tags' repositories. AMD64 docker images have no suffix. Ubuntu images have a '-ubuntu' suffix.",
			TarOpts: pipelines.TarFileOpts{
				Edition: "",
				Distro:  "linux/amd64",
				Version: version,
			},
			BaseImage: pipelines.BaseImageUbuntu,
			Tags: []string{
				"docker.io/grafana/grafana-image-tags:1.2.3-test.1.2.3-ubuntu-amd64",
				"docker.io/grafana/grafana-oss-image-tags:1.2.3-test.1.2.3-ubuntu-amd64",
			},
		},
		{
			Description: "Grafana docker images are created for both the 'docker.io/grafana/grafana-image-tags' and 'docker.io/grafana/grafana-oss-image-tags' repositories. ARM64 images have an -arm64 suffix. Ubuntu images have a '-ubuntu' suffix.",
			TarOpts: pipelines.TarFileOpts{
				Edition: "",
				Distro:  "linux/arm64",
				Version: version,
			},
			BaseImage: pipelines.BaseImageUbuntu,
			Tags: []string{
				"docker.io/grafana/grafana-image-tags:1.2.3-test.1.2.3-ubuntu-arm64",
				"docker.io/grafana/grafana-oss-image-tags:1.2.3-test.1.2.3-ubuntu-arm64",
			},
		},
		{
			Description: "Enterprise docker images are created for only the docker.io/grafana/grafana-enterprise-image-tags repository. AMD64 docker images have no suffix. Alpine images also have no suffix.",
			TarOpts: pipelines.TarFileOpts{
				Edition: "enterprise",
				Distro:  "linux/amd64",
				Version: version,
			},
			BaseImage: pipelines.BaseImageAlpine,
			Tags: []string{
				"docker.io/grafana/grafana-enterprise-image-tags:1.2.3-test.1.2.3-amd64",
			},
		},
		{
			Description: "Enterprise docker images are created for only the docker.io/grafana/grafana-enterprise-image-tags repository. ARM64 images have an -arm64 suffix. Alpine images also have no suffix.",
			TarOpts: pipelines.TarFileOpts{
				Edition: "enterprise",
				Distro:  "linux/arm64",
				Version: version,
			},
			BaseImage: pipelines.BaseImageAlpine,
			Tags: []string{
				"docker.io/grafana/grafana-enterprise-image-tags:1.2.3-test.1.2.3-arm64",
			},
		},
		{
			Description: "Enterprise docker images are created for only the docker.io/grafana/grafana-enterprise-image-tags repository. AMD64 docker images have no suffix. Ubuntu images have a '-ubuntu' suffix.",
			TarOpts: pipelines.TarFileOpts{
				Edition: "enterprise",
				Distro:  "linux/amd64",
				Version: version,
			},
			BaseImage: pipelines.BaseImageUbuntu,
			Tags: []string{
				"docker.io/grafana/grafana-enterprise-image-tags:1.2.3-test.1.2.3-ubuntu-amd64",
			},
		},
		{
			Description: "Enterprise docker images are created for only the docker.io/grafana/grafana-enterprise-image-tags repository. ARM64 images have an -arm64 suffix. Ubuntu images have a '-ubuntu' suffix.",
			TarOpts: pipelines.TarFileOpts{
				Edition: "enterprise",
				Distro:  "linux/arm64",
				Version: version,
			},
			BaseImage: pipelines.BaseImageUbuntu,
			Tags: []string{
				"docker.io/grafana/grafana-enterprise-image-tags:1.2.3-test.1.2.3-ubuntu-arm64",
			},
		},
	}

	for n, test := range cases {
		t.Run(fmt.Sprintf("[%d / %d] %s", n+1, len(cases), test.Description), func(t *testing.T) {
			expect := sort.StringSlice(test.Tags)
			res := sort.StringSlice(pipelines.GrafanaImageTags(test.BaseImage, "docker.io", test.TarOpts))

			for i := range expect {
				e := expect[i]
				r := res[i]
				if e != r {
					t.Errorf("[%d / %d]\nExpected '%s'\nReceived '%s'", i+1, len(expect), e, r)
				}
			}
		})
	}
}
