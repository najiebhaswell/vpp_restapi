# vpp_restapi

## Overview

**vpp_restapi** is a REST API service and supporting scripts for managing VPP (Vector Packet Processing) via HTTP.  
This package includes systemd integration, automatic VPP config dumping, and loader utilities.

---

## Prerequisites

- **Debian/Ubuntu Linux**
- **Go compiler** (for manual build)
- **VPP** installed and running
- **Python 3** (for loader/generator scripts)

---

## Installation (from .deb package)

1. **Clone the repository and build the .deb package:**
    ```bash
    git clone https://github.com/najiebhaswell/vpp_restapi.git
    cd vpp_restapi
    ```

2. **(Optional) Build the Go binary (if you want to recompile):**
    ```bash
    cd opt/vpp-restapi/cmd/server
    go build -o ../../vpp-restapi
    cd ../../../..
    ```

3. **Build the Debian package:**
    ```bash
    dpkg-deb --build vpp_restapi
    ```

4. **Install the package:**
    ```bash
    sudo dpkg -i vpp_restapi.deb
    ```

5. **Reload and enable services:**
    The post-install script will automatically:
    - Reload systemd
    - Enable and start all required services and timers:
        - `vpp-restapi`
        - `load-vpp-config`
        - `vpp-api-sock.path`
        - `generate-vpp-config.timer`

    To verify:
    ```bash
    systemctl status vpp-restapi
    systemctl status load-vpp-config
    systemctl status vpp-api-sock.path
    systemctl status generate-vpp-config.timer
    ```

---

## Service Descriptions

- **vpp-restapi.service**  
  REST API server for VPP.

- **load-vpp-config.service**  
  Python script to load configuration into VPP via the REST API.

- **vpp-api-sock.path**  
  Watches VPP API socket and restarts the REST API service if socket changes.

- **generate-vpp-config.timer/service**  
  Dumps VPP configuration to `/opt/vpp-config/vpp_dump.json` every 10 minutes via the API.

---

## Manual Usage

**Start/stop services manually:**
```bash
sudo systemctl start vpp-restapi
sudo systemctl start load-vpp-config
sudo systemctl start vpp-api-sock.path
sudo systemctl start generate-vpp-config.timer
```

**Check config dump:**
```bash
cat /opt/vpp-config/vpp_dump.json
```

---

## Uninstallation

```bash
sudo dpkg -r vpp-restapi
```

---

## Troubleshooting

- Pastikan VPP sudah berjalan sebelum start vpp-restapi.
- Jika ada error pada systemd, cek log dengan:
    ```bash
    journalctl -u vpp-restapi
    ```

- Untuk rebuild Go binary:
    ```bash
    cd opt/vpp-restapi/cmd/server
    go build -o ../../vpp-restapi
    ```

---

## Contributors

- [najiebhaswell](https://github.com/najiebhaswell)
