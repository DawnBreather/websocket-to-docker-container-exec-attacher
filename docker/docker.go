package docker

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"io"
	"log"
	"os"
	"quic_shell_server/db"
	"strings"
)

var Client *DockerClient

type DockerClient struct {
	cli *client.Client
}

func InitializeDockerClient() {
	var err error
	if Client == nil {
		Client, err = newDockerClient()
		if err != nil {
			log.Fatalf("Failed to create Docker client: %v", err)
		}
	}
}

func newDockerClient() (*DockerClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &DockerClient{cli: cli}, nil
}

func (d *DockerClient) PullImage(image string) error {
	ctx := context.Background()
	_, err := d.cli.ImagePull(ctx, image, types.ImagePullOptions{})
	return err
}

func (d *DockerClient) CreateAndStartContainer(image string, cmd []string, ttl int) (string, error) {
	ctx := context.Background()
	resp, err := d.cli.ContainerCreate(ctx, &container.Config{
		Image:  image,
		Labels: defaultDockerContainersLabels,
	}, &container.HostConfig{
		Privileged: true,
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeTmpfs,
				Target: "/sys/fs/cgroup",
				TmpfsOptions: &mount.TmpfsOptions{
					Mode: 1777,
				},
			},
		},
		AutoRemove: true,
	}, nil, nil, "")
	if err != nil {
		return "", err
	}

	if err := d.cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		//log.Printf("Failed starting up container { %s }: %v", resp.ID, err)
		return "", err
	}

	// Set TTL for the newly created container
	func() {
		var resTtl = ttl
		if ttl == 0 {
			resTtl = defaultDockerContainersTtlInMinutes * 60
		}
		db.SetTtlInSecondsByContainerId(resp.ID, resTtl)
		db.SetStartdatetimeInSecondsByContainerId(resp.ID)
	}()

	return resp.ID, nil
}

func (d *DockerClient) ExecIntoContainer(containerID string, cmd []string) error {
	ctx := context.Background()
	execID, err := d.cli.ContainerExecCreate(ctx, containerID, types.ExecConfig{
		Cmd:  cmd,
		Tty:  false,
		User: DefaultUserForExecIntoContainer,
	})
	if err != nil {
		return err
	}

	execStartCheck := types.ExecStartCheck{Tty: true}
	execAttachResp, err := d.cli.ContainerExecAttach(ctx, execID.ID, execStartCheck)
	if err != nil {
		return err
	}
	defer execAttachResp.Close()

	_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, execAttachResp.Reader)
	return err
}

func (d *DockerClient) StopAndRemoveContainer(containerID string) error {
	ctx := context.Background()
	err := d.cli.ContainerStop(ctx, containerID, container.StopOptions{})
	if err != nil {
		return err
	}
	return d.cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{})
}

func (d *DockerClient) CreateExecInstance(containerID string, cmd []string) (string, error) {
	ctx := context.Background()

	execConfig := types.ExecConfig{
		Cmd:          cmd,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Privileged:   true,
	}

	execIDResp, err := d.cli.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return "", err
	}

	return execIDResp.ID, nil
}

func (d *DockerClient) AttachExecInstance(execID string, stream io.ReadWriteCloser) error {
	ctx := context.Background()

	execStartCheck := types.ExecStartCheck{Tty: true} // Ensure this matches with your exec config

	execAttachResp, err := d.cli.ContainerExecAttach(ctx, execID, execStartCheck)
	if err != nil {
		return err
	}
	defer execAttachResp.Close()

	// Start the exec command
	if err := d.cli.ContainerExecStart(ctx, execID, execStartCheck); err != nil {
		return err
	}

	// Setting up bidirectional stream copy with logging
	errChan := make(chan error, 2)

	// Create a new CommandFilter
	filter := NewCommandFilter()

	// Intercepting and processing stdin
	go func() {
		_, err := filter.CopyAndInspect("WebSocket to Docker", execAttachResp.Conn, stream)
		errChan <- err
	}()

	// Processing stdout/stderr based on command filter state
	go func() {
		_, err := filter.CopyWithConditionalLogging("Docker to WebSocket", stream, execAttachResp.Reader)
		errChan <- err
	}()

	//// Copy from the WebSocket stream to the Docker exec's stdin with logging
	//go func() {
	//	_, err := copyWithLogging("WebSocket to Docker", execAttachResp.Conn, stream)
	//	errChan <- err
	//}()
	////
	//// Copy from the Docker exec's stdout/stderr to the WebSocket stream with logging
	//go func() {
	//	_, err := copyWithLogging("Docker to WebSocket", stream, execAttachResp.Reader)
	//	errChan <- err
	//}()

	// Wait for the first error or completion
	err = <-errChan
	return err
}

// copyWithLogging logs data being copied between streams
func copyWithLogging(direction string, dst io.Writer, src io.Reader) (int64, error) {
	buf := make([]byte, 32*1024) // Adjust buffer size to your needs
	var total int64
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			data := buf[0:nr]
			// Log as both string and hexadecimal for completeness
			log.Printf("%s: String Data: %s", direction, string(data))
			log.Printf("%s: Hex Data: %x", direction, data)

			nw, ew := dst.Write(data)
			if nw > 0 {
				total += int64(nw)
			}
			if ew != nil {
				return total, ew
			}
			if nr != nw {
				return total, io.ErrShortWrite
			}
		}
		if er != nil {
			if er != io.EOF {
				log.Printf("%s: Read error: %v", direction, er)
				return total, er
			}
			break
		}
	}
	return total, nil
}

// CommandFilter is responsible for inspecting and filtering command output.
type CommandFilter struct {
	suppressOutput bool
}

// NewCommandFilter creates a new instance of CommandFilter.
func NewCommandFilter() *CommandFilter {
	return &CommandFilter{}
}

// CopyAndInspect copies data from src to dst, inspecting the data for commands.
func (f *CommandFilter) CopyAndInspect(direction string, dst io.Writer, src io.Reader) (int64, error) {
	buf := make([]byte, 32*1024)
	var total int64
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			data := string(buf[:nr])
			// Check for specific commands to suppress their output
			if strings.Contains(data, "stty") && (strings.Contains(data, " cols") || strings.Contains(data, " rows")) {
				f.suppressOutput = true
			} else {
				// Reset suppression if the command is not matched
				// This is a simplification; actual logic may need to be more sophisticated
				f.suppressOutput = false
			}

			nw, ew := dst.Write(buf[:nr])
			if nw > 0 {
				total += int64(nw)
			}
			if ew != nil {
				return total, ew
			}
			if nr != nw {
				return total, io.ErrShortWrite
			}
		}
		if er != nil {
			if er != io.EOF {
				return total, er
			}
			break
		}
	}
	return total, nil
}

// CopyWithConditionalLogging copies data from src to dst, suppressing output if needed.
func (f *CommandFilter) CopyWithConditionalLogging(direction string, dst io.Writer, src io.Reader) (int64, error) {
	buf := make([]byte, 32*1024)
	var total int64
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			if !f.suppressOutput {
				nw, ew := dst.Write(buf[:nr])
				if nw > 0 {
					total += int64(nw)
				}
				if ew != nil {
					return total, ew
				}
				if nr != nw {
					return total, io.ErrShortWrite
				}
			}
		}
		if er != nil {
			if er != io.EOF {
				return total, er
			}
			break
		}
	}
	return total, nil
}
