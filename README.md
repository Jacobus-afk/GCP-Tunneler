# 🚇 GCP Tunneler 🚇

Command builder for `gcloud compute` ssh tunnels

## Requirements

- [fzf](https://github.com/junegunn/fzf?tab=readme-ov-file#installation) **- requires v0.61.1 or later**

- [jq](https://jqlang.org/download/)

- [tmux](https://github.com/tmux/tmux/wiki/Installing)

## Configuration

Configuration can be done via either environment variables, or a configuration file

### Environment variables

Set environment variables by appending it before the program

```shell
GCPT_INSTANCES_EXCLUDED=foo,bar GCPT_SSH_TIMEOUT=10 gcp-tunneler
```

| Name | Type | Default | Description |
| ---- | ---- | ------- | ----------- |
| GCPT_INSTANCES_EXCLUDED | comma separated list of strings | | **Non functioning** |
| GCPT_INSTANCES_INCLUDED | comma separated list of strings | | When populating the `gcp_resource_json` file, only add instance if the name includes one of the strings specified |
| GCPT_SSH_TIMEOUT | integer | 12 | Time in seconds allowed to establish an SSH connection |

### Configuration file

You can create a config file at `~/.config/gcp-tunneler/config.toml`. See [here](./config.toml) for an example
