# How Grafana uses Dagger

## What is Dagger?

[Dagger](https://dagger.io) is a programmable CI/CD engine.Under the hood it is basically a GraphQL server for [Docker BuildKit](https://docs.docker.com/build/buildkit/) and GraphQL clients for different programming languages with the goal of running all operations in Docker containers, creating a graph between what the user puts into the pipeline and what will eventually come out of it.

Dagger allows us to develop a CI pipeline using Go. More importantly, Dagger allows us to develop CI pipelines and run them locally, consistently with how they would run in CI. The underlying engine will always the same (or at least extremely similar), no matter where the pipeline is executed.

## Where Dagger adds value

### 1. Short feedback loop

The importance of having a shorter feedback loop can not be understated. Without being able to easily develop, test, and receive feedback from a CI pipeline it becomes easier to expand CI pipelines to meet new requirements rather than maintain existing ones, introducing "sprawl".

Grafana's build & release process as it exists today has reached a point where trvial changes require a non-trivial amount of work and a high amount of risk. If we want to test a change, we have to put it into a sandbox environment and explicitly disable all the steps we don't want to test or even "fake" certain event-triggers in order to get somehow rapid feedback. This also leads to a lot of risk when rolling out such a change as the test-environment will usually only barely resemble any production setup.

Dagger will allow us to consistently run and improve existing CI pipelines. The current feedback loop is long; it requires changes to be made, committed, and then ran by the CI server. In many situations, these changes may not even be properly ran by the server until the right event happens, like when we tag or promote a commit, which is a fairly "permanent" action with many side-effects.

With Dagger the feedback loop is significantly shorter. Instead of requiring a commit, developers can make a small change, run the entire pipeline locally, and see what happens and adjust appropriately. With Dagger's caching, which behaves similarly to Docker / BuildKit's caching, unnecessary tasks don't get repeated, increasing the feedback loop even more.

### 2. Consistent caching

Caching in Dagger is abstracted to how BuildKit handles it. There is a single API which allows a user to hook so-called `CacheVolumes` into a pipeline. While right now limited to the BuildKit caching layer, the Dagger team is working on extending it to also hook directly into S3 buckets, GitHub Action caching, and more.


### 3. Programmatic Docker image creation

All operations within Dagger from the point of view of a user happen within Docker containers. These are either created from a Dockerfile or built up through an API. This allows for more flexibility as we can just add various steps to a container based on if-statements et al. in Go.


### 4. CI as yet-another-library

Since all interactions with the Dagger engine happen through a language-SDK (e.g. the one for Go), we can treat reusable parts of the pipeline like any other Go library. For backend-developers this will lower the intensity of context-switching since everything is just Go anymore.


## Risks to consider

### 1. It's not a silver bullet.

While Dagger does help with tightening the feedback loop, it is possible to create pipelines that don't behave the same way remotely as they do locally. Care must be taken to avoid creating a dependency between the Dagger pipeline and the CI environment.

### 2. Just like with any script outside of the `yaml`, it circumvents some Drone safeties.

Drone has an option that requires that cryptographic signature in the yaml matches the yaml contents. When the code that processes the pipeline lives outside of the drone yaml, this signature doesn't do much to protect arbitrary code from running. Care must be taken to avoid running Dagger code directly from pull requests.

In Grafana, we avoid running code directly from the Grafana repository by only running dagger code that is in the 'main' branch of [grafana-build](https://github.com/grafana/grafana-build). So all changes to Grafana's dagger pipelines requires a merge to main.

### 3. It requires the Docker daemon.

* This makes it strange to run in a Kubernetes environment, where Pods have requests/limits.
* This makes it potentially insecure to run in CI in an open source project where forks are allowed.

Similar to number 2, we can (and already do) employ strategies that allow us to mitigate this risk that prevent users on forks from modifying code that is ran with the docker socket mounted in pull requests.

We also generally will not use Dagger in pull requests and will leave those processes exclusively to merges to main, release branches, or tags.

### 4. It obfuscates Drone pipeline steps.

This is a double-edged sword. In Drone, it is common to logically define and separate pipelines by purpose. It's common to see a "test", "build", "package", and "upload" pipelines or steps that are ran in sequential order. When using Dagger in the most valuable way, an entire pipeline will consist of one real "step", which is to run the whole Dagger pipeline. This has a drawback but also has a major benefit.

The benefit is that Dagger pipelines don't require the author to decide where each process starts and ends. Because Dagger is declarative, it figures that out for you based on the files and directories that are used in each container. This is beneficial because it can reduce the amount of waste that is introduced by haivng to make logical, sequential pipelines.

However, it does reduce observability some; it can be hard to know exactly what failed in a CI pipeline whenever there's one big step with a bunch of logs. That being said, the Dagger team is working on ways to make this easier and there are already ways to hook into OTEL traces generated by Dagger/BuildKit.

### Risks evaluation

We feel that the benefits far outweight the risks. Most of the security risk with using Dagger already has established patterns that we use across our CI. The reduced observability in the Drone UI could easily be overcome in future iterations of Dagger, and is has already been made a lot better in the latest version.
