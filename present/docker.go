package present

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/fsouza/go-dockerclient"
	"github.com/gorilla/websocket"
	"github.com/koding/websocketproxy"
)

var client *docker.Client

func init() {
	Register("docker", parseDocker)
	var err error
	endpoint := "unix:///var/run/docker.sock"
	client, err = docker.NewClient(endpoint)
	if err != nil {
		panic(err)
	}

}

// DockerEnabled specifies whether docker can be embedded in the
// present user interface.
var DockerEnabled = false

type Docker struct {
	Image     string
	Command   []string
	Container string
}

func (c Docker) TemplateName() string { return "docker" }

func parseDocker(ctx *Context, fileName string, lineno int, text string) (Elem, error) {
	args := strings.Fields(text)
	image := args[1]
	command := args[2:]

	container, err := startDocker(image, command)

	return Docker{image, command, container}, err
}

func startDocker(image string, args []string) (string, error) {
	/*	base, err := os.Getwd()
		if err != nil {
			return "", err
		}
		command := []string{"docker", "run", "--rm", "-it", "-v"}
		name := filepath.Join(base, "src")
		name = name + ":/go/src"
		command = append(command, name)
		command = append(command, image)
		command = append(command, args...)
	*/
	cfg := docker.Config{}
	cfg.AttachStderr = true
	cfg.AttachStdin = true
	cfg.AttachStdout = true
	cfg.Image = image
	cfg.Cmd = args
	cfg.Tty = true

	fmt.Println("starting image:", image)

	cco := docker.CreateContainerOptions{
		Config: &cfg,
	}

	container, err := client.CreateContainer(cco)
	if err != nil {
		panic(err)
	}

	err = client.StartContainer(container.ID, nil)
	return container.ID, err

}

func init() {
	u, err := url.Parse("ws://127.0.0.1:2375")
	if err != nil {
		panic(err)
	}
	p := websocketproxy.NewProxy(u)
	upgrader := &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		Subprotocols:    []string{"gotty"},
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	p.Upgrader = upgrader
	http.Handle("/ws/", p)
}
