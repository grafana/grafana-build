package main

var npmMapping = map[string]m{
	"Grafana data": {
		input: "file://dist/tag/grafana_10.2.0-pre_WxeEXjDuHTPB_linux_arm64/npm-artifacts/@grafana-data-10.2.0-pre.tgz",
		output: []string{
			"artifacts/npm/v10.2.0-pre/npm-artifacts/@grafana-data-v10.2.0-pre.tgz",
		},
		env: map[string]string{"DRONE_TAG": "10.2.0-pre"},
	},
}
