# WireGuard Reconnecter

> Lightweight Go daemon to monitor a WireGuard VPN connection and restart systemd serivce automatically if it goes down.

---

## Download

Pre-built binaries are available under [GitHub Releases](https://github.com/marcvorwerk/wireguard-reconnecter/releases).

Download the latest release, make it executable, and move it to a directory in your `PATH`.

Example:

```bash
wget https://github.com/marcvorwerk/wireguard-reconnecter/releases/download/v1.0.0/wireguard-reconnecter
chmod +x wireguard-reconnecter
sudo mv wireguard-reconnecter /usr/local/bin/
```

### Fix permissions if needed

sudo sysctl -w net.ipv4.ping_group_range="0 2147483647"
or
sudo setcap cap_net_raw+ep ./wgmon
