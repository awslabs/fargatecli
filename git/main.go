package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/jpignata/fargate/console"
)

func GetShortSha() string {
	console.Debug("Finding git HEAD short SHA")
	console.Shell("git rev-parse --short HEAD")

	buf := new(bytes.Buffer)
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	stdout, _ := cmd.StdoutPipe()

	if console.Verbose {
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Start(); err != nil {
		console.ErrorExit(err, "Could not find git HEAD short SHA")
	}

	buf.ReadFrom(stdout)

	sha := strings.TrimSpace(buf.String())

	if console.Verbose {
		fmt.Println(sha)
	}

	return sha
}

func IsCwdGitRepo() bool {
	console.Debug("Checking if current working directory is a git repository")
	console.Shell("git rev-parse --is-inside-work-tree")

	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")

	if err := cmd.Run(); err == nil {
		return true
	} else {
		return false
	}
}
