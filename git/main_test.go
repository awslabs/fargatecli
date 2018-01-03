package git

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestGetShortSha(t *testing.T) {
	cwd, err := os.Getwd()

	if err != nil {
		t.Error("Could not read current working directory", err)
		return
	}

	dir, err := ioutil.TempDir("", "fargate-tests")

	if err != nil {
		t.Error("Could not create temporary directory", err)
		return
	}
	defer os.RemoveAll(dir)

	os.Chdir(dir)
	defer os.Chdir(cwd)

	exec.Command("git", "init").Run()

	gitCommit := exec.Command("git", "commit", "--allow-empty", "--message", "dummy commit")
	commitOutput, err := gitCommit.CombinedOutput()

	if err != nil {
		t.Error("Could not create dummy git commit", err)
		return
	}

	if shortSha := GetShortSha(); !strings.Contains(string(commitOutput), GetShortSha()) {
		t.Errorf("expected %s to contain %s", commitOutput, shortSha)
	}
}

func TestIsCwdGitRepoAgainstADir(t *testing.T) {
	cwd, err := os.Getwd()

	if err != nil {
		t.Error("Could not read current working directory", err)
		return
	}

	dir, err := ioutil.TempDir("", "fargate-tests")

	if err != nil {
		t.Error("Could not create temporary directory", err)
		return
	}
	defer os.RemoveAll(dir)

	os.Chdir(dir)
	defer os.Chdir(cwd)

	if isCwdGitRepo := IsCwdGitRepo(); isCwdGitRepo {
		t.Errorf("wanted false, got %+v", isCwdGitRepo)
	}
}

func TestIsCwdGitRepoAgainstARepo(t *testing.T) {
	cwd, err := os.Getwd()

	if err != nil {
		t.Error("Could not read current working directory", err)
		return
	}

	dir, err := ioutil.TempDir("", "fargate-tests")

	if err != nil {
		t.Error("Could not create temporary directory", err)
		return
	}
	defer os.RemoveAll(dir)

	os.Chdir(dir)
	defer os.Chdir(cwd)

	cmd := exec.Command("git", "init")
	cmd.Run()

	if isCwdGitRepo := IsCwdGitRepo(); !isCwdGitRepo {
		t.Errorf("wanted true, got %+v", isCwdGitRepo)
	}
}
