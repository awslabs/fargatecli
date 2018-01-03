package git

import (
	"os"
	"os/exec"
	"strings"

	"github.com/jpignata/fargate/console"
)

func GetShortSha() string {
	var sha string

	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")

	if console.Verbose {
		cmd.Stderr = os.Stderr
	}

	if out, err := cmd.Output(); err == nil {
		sha = strings.TrimSpace(string(out))
	} else {
		console.ErrorExit(err, "Could not find git HEAD short SHA")
	}

	return sha
}

func IsCwdGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	err := cmd.Run()

	return err == nil
}
