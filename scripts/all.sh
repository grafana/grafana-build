dagger run go run ./cmd artifacts \
  -a frontend:enterprise \
  -a storybook \
  -a npm:grafana \
  -a targz:grafana:linux/amd64 \
  -a targz:grafana:linux/arm64 \
  -a targz:grafana:linux/riscv64 \
  -a targz:grafana:linux/arm/v6 \
  -a targz:grafana:linux/arm/v7 \
  -a targz:enterprise:linux/amd64 \
  -a targz:enterprise:linux/arm64 \
  -a targz:enterprise:linux/riscv64 \
  -a targz:enterprise:linux/arm/v6 \
  -a targz:enterprise:linux/arm/v7 \
  -a targz:boring:linux/amd64/dynamic \
  -a deb:grafana:linux/amd64 \
  -a deb:grafana:linux/arm64 \
  -a deb:grafana:linux/arm/v6 \
  -a deb:grafana:linux/arm/v7 \
  -a deb:enterprise:linux/amd64 \
  -a deb:enterprise:linux/arm64 \
  -a deb:enterprise:linux/arm/v6 \
  -a deb:enterprise:linux/arm/v7 \
  -a rpm:grafana:linux/amd64:sign \
  -a rpm:grafana:linux/arm64:sign \
  -a rpm:enterprise:linux/amd64:sign \
  -a rpm:enterprise:linux/arm64:sign \
  -a docker:grafana:linux/amd64 \
  -a docker:grafana:linux/arm64 \
  -a docker:grafana:linux/amd64:ubuntu \
  -a docker:grafana:linux/arm64:ubuntu \
  -a docker:enterprise:linux/amd64 \
  -a docker:enterprise:linux/arm64 \
  -a docker:enterprise:linux/amd64:ubuntu \
  -a docker:enterprise:linux/arm64:ubuntu \
  -a docker:boring:linux/amd64/dynamic-musl \
  -a zip:grafana:windows/amd64 \
  -a zip:enterprise:windows/amd64 \
  -a zip:grafana:windows/arm64 \
  -a zip:enterprise:windows/arm64 \
  -a exe:grafana:windows/amd64 \
  -a exe:enterprise:windows/amd64 \
  -build-id=103 \
  --grafana-ref=v10.1.0 \
  --checksum \
  --enterprise-ref=v10.1.0 > out.txt

