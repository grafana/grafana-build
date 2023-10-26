package pipelines

// // Docker is a pipeline that uses a grafana.tar.gz as input and creates a Docker image using that same Grafana's Dockerfile.
// // Grafana's Dockerfile should support supplying a tar.gz using a --build-arg.
// func Docker(ctx context.Context, d *dagger.Client, args PipelineArgs) error {
// 	tarballs, err := containers.GetPackages(ctx, d, args.PackageInputOpts, args.GCPOpts)
// 	if err != nil {
// 		return err
// 	}
// 	var (
// 		opts        = args.DockerOpts
// 		publishOpts = args.PublishOpts
// 		saved       = map[string]*dagger.File{}
// 	)
//
// 	for i, v := range args.PackageInputOpts.Packages {
// 		tarOpts := packages.NameOptsFromFileName(v)
//
// 		var (
// 			targz = tarballs[i]
// 		)
//
// 		bases := []docker.BaseImage{docker.BaseImageAlpine, docker.BaseImageUbuntu}
// 		for _, base := range bases {
// 			var (
// 				platform  = backend.Platform(tarOpts.Distro)
// 				baseImage = opts.AlpineBase
// 				socket    = d.Host().UnixSocket("/var/run/docker.sock")
// 				format    = opts.TagFormat
// 			)
// 			if base == docker.BaseImageUbuntu {
// 				format = opts.UbuntuTagFormat
// 			}
//
// 			tags, err := docker.Tags(opts.Org, opts.Registry, opts.Repository, format, tarOpts)
// 			if err != nil {
// 				return err
// 			}
//
// 			if base == docker.BaseImageUbuntu {
// 				baseImage = opts.UbuntuBase
// 			}
//
// 			log.Println("Building docker images", tags, "with base image", baseImage, "and platform", platform)
//
// 			builder := docker.Builder(d, socket, targz)
// 			builder = docker.Build(d, builder, &docker.BuildOpts{
// 				Tags:      tags,
// 				BaseImage: baseImage,
// 				Platform:  platform,
// 			})
//
// 			// if --save was provided then we will publish this to the requested location using PublishFile
// 			if publishOpts.Destination != "" {
// 				ext := "docker.tar.gz"
// 				if base == docker.BaseImageUbuntu {
// 					ext = "ubuntu.docker.tar.gz"
// 				}
// 				name := ReplaceExt(v, ext)
// 				img := builder.WithExec([]string{"docker", "save", tags[0], "-o", name}).File(name)
// 				dst := strings.Join([]string{publishOpts.Destination, name}, "/")
// 				saved[dst] = img
// 			}
// 		}
// 	}
//
// 	var (
// 		wg = &errgroup.Group{}
// 		sm = semaphore.NewWeighted(args.ConcurrencyOpts.Parallel)
// 	)
// 	for dst, file := range saved {
// 		wg.Go(PublishFileFunc(ctx, sm, d, &containers.PublishFileOpts{
// 			Destination: dst,
// 			File:        file,
// 			GCPOpts:     args.GCPOpts,
// 			PublishOpts: args.PublishOpts,
// 		}))
// 	}
//
// 	return wg.Wait()
// }
