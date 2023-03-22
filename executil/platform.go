package executil

import (
	"strings"

	"dagger.io/dagger"
)

// Platform returns the platform for running the specified docker container.
// The goal of this function is to return a dagger.Platform that is capable of running on the current machine.
// Scenarios:
// * amd64 Linux machines are capable of running `amd64`, `arm`, and `arm64` architectures on Linux containers.
// * arm64 Darwin machines (M1, M2 macs) can run `arm` and `arm64` architectures on Linux containers.
// * amd64 Windows machines are capable of running `amd64` Linux machines.
func Platform(d Distribution) dagger.Platform {
	// If the distro can be ran by this OS/Arch then just use that.
	if DistroOneOf(d, Capabilities) {
		return dagger.Platform(d)
	}

	// If not, like for 'darwin/arm64', we'll need to cross-compile.
	// In order to cross-compile, we should try to stick to the same architecture as requested.
	// but we'll use the OS of "linux" and install the approprite compiler toolchain.
	isAMD64 := strings.Split(string(d), "/")[1] == "amd64"
	if isAMD64 && DistroOneOf(DistLinuxAMD64, Capabilities) {
		return dagger.Platform(DistLinuxAMD64)
	}

	// Pretty much everything is capable of running arm containers in Docker.
	// But if it's not listed in the capabilities, then we'll default to using the current OS/arch (""), whic
	// probably won't work.
	isARM := strings.Split(string(d), "/")[1] == "arm"
	if isARM {
		if isARM && DistroOneOf(DistLinuxARM64, Capabilities) {
			return dagger.Platform(DistLinuxARM64)
		}
	}

	return dagger.Platform("")
}
