package utils

import (
	"os/exec"
	"strings"

	"github.com/rs/zerolog/log"
)

func RunCommand(cmdName string, cmdArgs ...string) string {
	cmd := exec.Command(cmdName, cmdArgs...)
	// cmd.SysProcAttr = &syscall.SysProcAttr{
	// 	Setpgid: true, // allows signals to propogate to child
	// 	Pgid:    0,
	// }
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Msg("Error running command")
	}

	return strings.TrimSpace(string(out))
}
