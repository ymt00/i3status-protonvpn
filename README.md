# Bemenu Proton VPN

Depends on bemenu and Proton VPN community script 

## Usage in bemenu config file

```
[[block]]
block = "custom"
command = "cat /path/to/protonvpn_status_file"
json = true
interval = "once"
watch_files = ["/path/to/protonvpn_status_file"]
if_command = "ls /path/to/protonvpn_status_file"
[[block.click]]
button = "left"
cmd = "/path/to/protonvpn_go_binary /path/to/protonvpn_status_file"
```