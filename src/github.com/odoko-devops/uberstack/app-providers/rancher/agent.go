package rancher

import (
        "fmt"
        "encoding/json"
        "net"
        "net/http"
        "time"
        "github.com/odoko-devops/uberstack/utils"
        "log"
)

const agent_version = "v1.0.2"

func identifyIpAddress(networkInterface string) (string, error) {
        ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

        for _, i := range ifaces {
                addrs, err := i.Addrs()
		if err != nil {
			return "", err
		}
                if i.Name == networkInterface {
                        for _, addr := range addrs {
                                switch v := addr.(type) {
                                case *net.IPNet:
                                        return v.IP.String(), nil
                                case *net.IPAddr:
                                        return v.IP.String(), nil
                                }
                        }
                }
        }
        panic("No IP address found for " + networkInterface)
 }

type rancherEnvironmentResponse struct {
        Data []struct {
                Id   string `json:"id"`
                Name string `json:"name"`
                Kind string `json:"kind"`
        } `json:"data"`
}

func identifyRancherEnvironment(rancherHostname, accessKey, secretKey string) string {
        rancherEnvUrl := fmt.Sprintf("http://%s/v1/accounts", rancherHostname)
        log.Printf("Identifying Rancher environment ID via %s", rancherEnvUrl)
        client := &http.Client{}
        for {
                req, _ := http.NewRequest("GET", rancherEnvUrl, nil)
                req.SetBasicAuth(accessKey, secretKey)
                res, err := client.Do(req)
		if (err != nil) {
			time.Sleep(5*time.Second)
			println("Waiting for Rancher...")
			continue
		}

                body := rancherEnvironmentResponse{}
                err = json.NewDecoder(res.Body).Decode(&body)
		if (err != nil) {
			time.Sleep(5*time.Second)
			println("Waiting for Rancher...")
			continue
		}

                for i := range body.Data {
                        env := body.Data[i]
                        if env.Name == "Default" && env.Kind == "project" {
                                return env.Id
                        }
                }

                log.Println("Environment not found, waiting")
                time.Sleep(5 * time.Second)
        }
}
type registrationResponse struct {
        Data []struct {
            RegistrationUrl string
        }
}

func identifyRegistrationUrl(rancherHostname, accessKey, secretKey, rancherEnvironment string) (string, error) {
        rancherTokensUrl := fmt.Sprintf("http://%s/v1/projects/%s/registrationtokens?state=active&limit=-1",
               rancherHostname, rancherEnvironment)
        log.Printf("Seeking registration URL with %s\n", rancherTokensUrl)
        client := &http.Client{}
        for {
                req, err := http.NewRequest("GET", rancherTokensUrl, nil)
                if err != nil {
                        return "", err
                }
                req.SetBasicAuth(accessKey, secretKey)
                res, err := client.Do(req)
                if err != nil {
                        return "", err
                }
                body := registrationResponse{}
                json.NewDecoder(res.Body).Decode(&body)

		if len(body.Data)>0 {
			registrationUrl := body.Data[0].RegistrationUrl

			if registrationUrl != "" {
				return registrationUrl, nil
			}
		}
                log.Println("Registration URL not found, waiting")
                time.Sleep(5 * time.Second)
        }
}

func installRancherAgent(ip_address, labels, rancher_url string) {
        command := fmt.Sprintf(
            `sudo docker run \
            -e CATTLE_AGENT_IP=%v \
            -e CATTLE_HOST_LABELS=%s \
            -d --privileged --name rancher-bootstrap \
            -v /var/run/docker.sock:/var/run/docker.sock \
            -v /var/lib/rancher:/var/lib/rancher \
              rancher/agent:%s %s`,
                ip_address,
                labels,
                agent_version,
                rancher_url)
        utils.Execute(command, nil, "")
	log.Println("Rancher agent installed")
}
