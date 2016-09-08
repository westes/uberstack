package apps

import (
	"installer/model"
	"fmt"
	"utils"
)

func Proxy_Install(config model.Config, hostConfig model.HostConfig, app model.AppConfig) error {
	jenkinsHost := app.Config["jenkins-host"]
	rancherHost := app.Config["rancher-host"]

	command := fmt.Sprintf(`docker run -d -p 80:80 \
			-e JENKINS_HOSTNAME=%s \
			-e JENKINS_PORT=8081 \
			-e RANCHER_HOSTNAME=%s \
			-e RANCHER_PORT=8080 \
			odoko/auth-proxy:1.0.1`,
		jenkinsHost,
		rancherHost)
	utils.ExecuteRemote(hostConfig.Name, command, nil, "")
	return nil
}
