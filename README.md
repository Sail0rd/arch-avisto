# Arch Avisto WSL

This project contains the sources and the documentation to ship updates of the Arch Avisto image.

For documentation about the [installation](https://advans-group.atlassian.net/wiki/spaces/DS/pages/2919563271/WSL+2+Get+started#%F0%9F%86%95-Import-Avisto-image-[Recommended]) as a consumer.

## Setup

### Pre-commit

This project uses pre-commit tool [documentation](https://pre-commit.com/).

The goal is to run scripts or tasks (called hooks) on a commit basis (e.g. linting, small checks, documentation as code, etc.).

For this purpose, pre-commit will execute all the hooks declared in in the `.pre-commit-config.yaml` configuration file.

1. Install the tool. `pip install pre-commit`
2. Install the git hook. `pre-commit install`
3. Verify it is working by running pre-commit hooks. `pre-commit run --all-files`

### Vlang

The bootscript is written in [V](https://vlang.io/), a simple, fast, safe, compiled programming language.

## Introduction

### History

The aim of Arch Avisto was to deliver a container solution (docker in our case) out-of-the box for the teams without relaying on Docker Desktop, for licence reasons. Indeed, its installation on WSL was error prone and time consuming.

Arch Avisto is based on [Arch Linux](https://archlinux.org/), and its first version was born using [ArchWSL](https://github.com/yuk7/ArchWSL).

### Architecture

There are two users in this image
- `arch` is the normal user.
- `login` is the default user on the first launch. Its job is to launch the bootscript.

They are a few important files on the Arch Avisto WSL image.

- `/home/arch` is the home of the default user, containing `.oh-my-wsh`, `.zshrc` and `.config/starship.toml` as basic configuration
- `/usr/bin/login` is the binary executable from the vlang bootscript
- `/usr/bin/slogin` is the shell of the `login` user. This shell starts `/usr/bin/login`
- `/opt/prep.sh` is a script to prepare the image before exporting it

## Step-by-step guide to ship new version

> Before starting be sure you run the latest WSL version (see [doc](https://advans-group.atlassian.net/wiki/spaces/DS/pages/2919563271/WSL+2+Get+started#Install%2FUpdate-WSL-2)). \
Also install [7-zip](https://www.7-zip.org/).

1. Download the latest version from [SharePoint](https://groupadvans.sharepoint.com/:f:/s/DevOpsSupport/ErkpQpmnXEtNnZbT5MnMVE8BOESKO38rK_BXh-luBE0gTA?e=6HTMM0)
1. Use the [documentation](https://advans-group.atlassian.net/wiki/spaces/DS/pages/2919563271/WSL+2+Get+started#%F0%9F%86%95-Import-Avisto-image-[Recommended]) to import the image, then login using `wsl -u arch -d <distro-name>`
1. Do your changes
1. Update the version in `/usr/bin/slogin` file
1. Clean the distribution by running `/opt/prep.sh`
1. Confirm your history is empty (^R), if not, run each command in the previous script manually
1. Open a Powershell
1. Shutdown wsl `wsl --terminate <distro-name>`
1. Export the distribution as a compressed archive  (see the command below)
1. Upload it to the SharePoint (step 1)
1. Make at least 1 person test it
1. Change SharePoint link on the documentation (step 2) to point to the new version


```powershell
# Export the distribution as a compressed archive
wsl --export <distro-name> .\arch_avisto_v<semver>.tar
& "C:\Program Files\7-Zip\7z.exe" a .\arch_avisto_v<semver>.tar.gz .\arch_avisto_v<semver>.tar
```
