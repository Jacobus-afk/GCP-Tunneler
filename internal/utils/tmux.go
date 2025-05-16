package utils

import (
	"context"
	"fmt"
	"gcp-tunneler/internal/config"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
)

type SSHConnection struct {
	Connected bool
	Port      string
	Username  string
}

func CreateTMUXSSHSession(sshConnection SSHConnection, sessionName string) error {
	if !sshConnection.Connected {
		return fmt.Errorf("ssh not connected for %s", sshConnection.Port)
	}

	sshCmd := fmt.Sprintf(
		"ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null %s@localhost -p %s",
		sshConnection.Username,
		sshConnection.Port,
	)

	sessionExists := CommandRun("tmux", "has-session", "-t", sessionName) == nil
	if sessionExists {
		// Split the window vertically
		if err := CommandRun("tmux", "split-window", "-h", "-t", sessionName+":connections"); err != nil {
			return fmt.Errorf("failed to split window for %s: %w", sessionName, err)
		}
	} else {
		if err := CommandRun("tmux", "new", "-d", "-s", sessionName); err != nil {
			return fmt.Errorf("failed to create session for %s: %w", sessionName, err)
		}

		if err := CommandRun("tmux", "new-window", "-t", sessionName, "-n", "connections"); err != nil {
			return fmt.Errorf("failed to create new window for %s: %w", sessionName, err)
		}
	}

	if err := CommandRun("tmux", "send-keys", "-t", sessionName+":connections", sshCmd, "C-m"); err != nil {
		return fmt.Errorf(
			"failed to start SSH for %s, port %s: %w",
			sessionName,
			sshConnection.Port,
			err,
		)
	}
	return nil
}

func ArrangeLayout(sessionName string) {
	err := CommandRun("tmux", "select-layout", "-t", sessionName+":connections", "tiled")
	if err != nil {
		log.Error().Err(err).Msgf("failed to arrange panes for %s", sessionName)
	}
}

func CreateTMUXTunnelSession(gcloudCMD string, instanceName string) {
	_ = CommandRun("tmux", "new", "-d", "-s", "GCPTunnel")

	_ = CommandRun(
		"tmux",
		"new-window",
		"-t",
		"GCPTunnel",
		"-n",
		fmt.Sprintf("iap-%s", instanceName),
		gcloudCMD,
	)
}

func WaitForSSHSession(currentUser string, freePort string) bool {
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
				freePort,
				"true",
			)
			if err == nil {
				log.Debug().Msg("established ssh connection")
				return true
			}
			log.Debug().
				Err(err).
				Msgf("SSH connection on %s@localhost:%s attempt failed, retrying..", currentUser, freePort)
		}
	}
}

func SwitchToCreatedSession(sessionName string) error  {
	binary, lookErr := exec.LookPath("tmux")
	if lookErr != nil {
		return fmt.Errorf("could not find path to tmux: %w", lookErr)
	}

	args := []string{"tmux", "switch", "-t", sessionName}

	env := os.Environ()

	execErr := syscall.Exec(binary, args, env)
	if execErr != nil {
		return fmt.Errorf("could not run tmux switch: %w", execErr)
	}
	
	return nil
}
