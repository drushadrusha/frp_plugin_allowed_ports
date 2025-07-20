# frps_allowed_ports (Forked)

This is a forked version of [frps_allowed_ports](https://github.com/Parmicciano/frp_plugin_allowed_ports) with added functionality to automatically update the allowed ports list from `ports.ini` without requiring a service restart.

frp server plugin to define allowed ports for a specific user for [frp](https://github.com/fatedier/frp).

## Features

*   Support the verification of the port used by the users by ports & subdomain saved in a file.
*   **Automatic Port Reloading**: Changes to `ports.ini` are detected and applied automatically without service restart.

## How to use

1.  Download frps_allowed_ports binary file from [Release](https://github.com/Parmicciano/frp_plugin_allowed_ports/releases).
2.  Put the binary file in the `plugins` directory of frps.
3.  Create file `ports` including all support usernames and ports.

    Example `ports.ini`:

    ```ini
    [common]
    user1 = 8080,8081
    user2 = 9000
    user3 = subdomain1.example.com
    ```

    Note: The `[common]` section is required.

4.  Configure frps:

    ```ini
    [common]
    bind_port = 7000
    vhost_http_port = 80
    dashboard_port = 7500

    [plugin.frp_plugin_allowed_ports]
    addr = 127.0.0.1:9000
    path = /handler
    ```

5.  Run frps_allowed_ports:

    ```bash
    ./frp_plugin_allowed_ports -c ./ports.ini
    ```

    Or as a systemd service:

    ```ini
    [Unit]
    Description=frps_allowed_ports
    After=network.target

    [Service]
    Type=simple
    ExecStart=/path/to/frps_allowed_ports -c /path/to/ports.ini
    Restart=on-failure

    [Install]
    WantedBy=multi-user.target
    ```

## Example frpc configuration

```ini
[common]
server_addr = 127.0.0.1
server_port = 7000

[ssh]
type = tcp
local_ip = 127.0.0.1
local_port = 22
remote_port = 6000
user = user1

[web]
type = http
local_ip = 127.0.0.1
local_port = 80
subdomain = subdomain1
user = user3
```