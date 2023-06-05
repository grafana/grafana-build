package pipelines

import (
	"context"
	"testing"

	"github.com/grafana/grafana-build/containers"
	"github.com/grafana/grafana-build/executil"
	"github.com/stretchr/testify/require"
)

func TestArtifactDependencyConstraintCheck(t *testing.T) {
	reg := NewArtifactDefinitionRegistry()
	reg.Register("tarball", NewArtifactDefinition())
	reg.Register("windows", NewArtifactDefinition().
		WithRequirement("/mnt/tarball", "tarball").
		WithConstraint(ConstraintWindowsOnly))
	reg.Register("deb", NewArtifactDefinition().
		WithRequirement("/mnt/tarball", "tarball").
		WithConstraint(ConstraintLinuxOnly))
	ctx := context.Background()

	t.Run("windows-for-windows", func(t *testing.T) {
		result, err := CheckArtifactChainConstraint(ctx, RequestedArtifact{
			Name: "windows",
			Options: ArtifactGeneratorOptions{
				Distribution: executil.DistWindowsAMD64,
			},
		}, reg)
		require.NoError(t, err)
		require.True(t, result)
	})
	t.Run("no-windows-for-non-windows", func(t *testing.T) {
		result, err := CheckArtifactChainConstraint(ctx, RequestedArtifact{
			Name: "windows",
			Options: ArtifactGeneratorOptions{
				Distribution: executil.DistLinuxAMD64,
			},
		}, reg)
		require.NoError(t, err)
		require.False(t, result)
	})
}

func TestGenerateFinalArtifactList(t *testing.T) {
	reg := NewArtifactDefinitionRegistry()
	reg.Register("tarball", NewArtifactDefinition())
	reg.Register("windows", NewArtifactDefinition().
		WithRequirement("/mnt/tarball", "tarball").
		WithConstraint(ConstraintWindowsOnly))
	reg.Register("deb", NewArtifactDefinition().
		WithRequirement("/mnt/tarball", "tarball").
		WithConstraint(ConstraintLinuxOnly))
	ctx := context.Background()

	t.Run("skip-deb-for-windows", func(t *testing.T) {
		args := PipelineArgs{}
		args.PackageOpts = &containers.PackageOpts{}
		args.PackageOpts.Distros = []executil.Distribution{executil.DistWindowsAMD64}
		result, err := GeneratateFinalArtifactList(ctx, reg, []string{"windows", "deb"}, args)
		require.NoError(t, err)
		require.Len(t, result, 1)
		req := result[0]
		require.Equal(t, "windows", req.Name)
	})

	t.Run("skip-windows-for-linux", func(t *testing.T) {
		args := PipelineArgs{}
		args.PackageOpts = &containers.PackageOpts{}
		args.PackageOpts.Distros = []executil.Distribution{executil.DistLinuxAMD64}
		result, err := GeneratateFinalArtifactList(ctx, reg, []string{"windows", "deb"}, args)
		require.NoError(t, err)
		require.Len(t, result, 1)
		req := result[0]
		require.Equal(t, "deb", req.Name)
	})
}
