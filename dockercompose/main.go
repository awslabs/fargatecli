package dockercompose

import (
	"bytes"
	"os/exec"

	"github.com/jpignata/fargate/console"
	yaml "gopkg.in/yaml.v2"
)

//ComposeFile ...
type ComposeFile struct {
	File string
}

//NewComposeFile ...
func NewComposeFile(file string) ComposeFile {
	return ComposeFile{
		File: file,
	}
}

// DockerCompose represents a docker-compose.yml file
type DockerCompose struct {
	Services map[string]*Service `yaml:"services"`
}

// Service represents a docker container
type Service struct {
	Image       string            `yaml:"image,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
}

//Config returns a DockerCompose representation of the file
//note that all variable interpolations are fully rendered by the config command
func (composeFile *ComposeFile) Config() *DockerCompose {
	console.Debug("running docker-compose config [%s]", composeFile.File)
	cmd := exec.Command("docker-compose", "-f", composeFile.File, "config")

	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	if err := cmd.Start(); err != nil {
		console.ErrorExit(err, errbuf.String())
	}

	if err := cmd.Wait(); err != nil {
		console.IssueExit(errbuf.String())
	}

	//unmarshal the yaml
	var compose DockerCompose
	err := yaml.Unmarshal(outbuf.Bytes(), &compose)
	if err != nil {
		console.ErrorExit(err, "error unmarshalling docker-compose.yml")
	}

	return &compose
}
