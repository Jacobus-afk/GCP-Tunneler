package utils

import (
	"context"
	"fmt"
	"gcp-tunneler/config"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

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

func WaitForSSHSession(currentUser string, freePort int) bool {
	timeoutVal := config.GetConfig().SSHTimeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutVal)*time.Second)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Warn().Msgf("SSH connection timed out after %d seconds", timeoutVal)
			return false
		case <-ticker.C:
			err := CommandRun(
				"ssh",
				"-q",
				"-o",
				"StrictHostKeyChecking=no",
				"-o",
				"UserKnownHostsFile=/dev/null",
				currentUser+"@localhost",
				"-p",
				strconv.Itoa(freePort),
				"true",
			)
			if err == nil {
				log.Debug().Msg("established ssh connection")
				return true
			}
			log.Debug().Err(err).Msgf("SSH connection on %s@localhost:%d attempt failed, retrying..", currentUser, freePort)
		}
	}
}
