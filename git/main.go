package git

import (
	"bytes"
	"os"
	"os/exec"
	"strings"

	"github.com/jpignata/fargate/console"
)

func GetShortSha() string {
	console.Debug("Finding git HEAD short SHA")

	buf := new(bytes.Buffer)
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	stdout, _ := cmd.StdoutPipe()

	if console.Verbose {
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		console.ErrorExit(err, "Could not find git HEAD short SHA")
	}

	buf.ReadFrom(stdout)

	return strings.TrimSpace(buf.String())
}

func IsCwdGitRepo() bool {
	console.Debug("Checking if current working directory is a git repository")

	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	err := cmd.Run()

	return err == nil
}
