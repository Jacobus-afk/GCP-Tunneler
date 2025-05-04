package utils

import "fmt"

func CreateTMUXTunnelSession(gcloudCMD []string, instanceName string) {
	RunCommand("tmux", "new", "-d", "-s", "GCPTunnel")

	RunCommand(
		"tmux",
		"new-window",
		"-t",
		"GCPTunnel",
		"-n",
		fmt.Sprintf("iap-%s", instanceName),
		gcloudCMD,
	)
}
