package utils

import "fmt"

func CreateTMUXSSHSession() {

}

func CreateTMUXTunnelSession(gcloudCMD string, instanceName string) {
	CommandRun("tmux", "new", "-d", "-s", "GCPTunnel")

	CommandRun(
		"tmux",
		"new-window",
		"-t",
		"GCPTunnel",
		"-n",
		fmt.Sprintf("iap-%s", instanceName),
		gcloudCMD,
	)
}
