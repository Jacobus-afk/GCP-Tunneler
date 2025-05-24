# 🚇 GCP Tunneler 🚇

Command builder for `gcloud compute` ssh tunnels

## Requirements

- [fzf](https://github.com/junegunn/fzf?tab=readme-ov-file#installation) **- requires v0.61.1 or later**

- [jq](https://jqlang.org/download/)

- [tmux](https://github.com/tmux/tmux/wiki/Installing)

## Installation

### Linux

```shell
curl -s -H "Accept:application/vnd.github.v3.raw" https://api.github.com/repos/Jacobus-afk/gcp-tunneler/contents/install.sh | bash -s
```

## Configuration

Configuration can be done via either environment variables, or a configuration file

### Environment variables

Set environment variables by appending it before the program

```shell
GCPT_INSTANCES_EXCLUDED=foo,bar GCPT_SSH_TIMEOUT=10 gcp-tunneler
```

| Name | Type | Default | Description |
| ---- | ---- | ------- | ----------- |
| GCPT_INSTANCES_EXCLUDED | comma separated list | | **Non functioning** |
| GCPT_INSTANCES_INCLUDED | comma separated list | | When populating the `gcp_resource_json` file, only add instance if the name includes one of the filters specified |
| GCPT_SSH_TIMEOUT |  | 12 | Time in seconds allowed to establish an SSH connection |
| GCPT_SSH_USERNAME |  |  | To use a name other than the default OS username to establish an SSH session |
| GCPT_DEVELOP_DEBUG |  |  | Enable debug logging |

### Configuration file

You can create a config file in `~/.config/gcp-tunneler/config.toml`. See [here](./config.toml.example) for an
example (remember to rename the file to `config.toml`)

> [!NOTE]
> The configuration options under [Environment variables](#environment-variables) are analogous with the
> entries in `config.toml`. To change, for example, the default ssh host user name (GCPT_SSH_USERNAME):
>
> ```toml
> [ssh]
> username = "atticus"
> ```
