# collectd-docker-plugin

<a href="https://travis-ci.org/dustinblackman/collectd-docker-plugin"><img src="https://img.shields.io/travis/dustinblackman/collectd-docker-plugin.svg" alt="Build Status"></a> <a href="https://goreportcard.com/report/github.com/dustinblackman/collectd-docker-plugin"><img src="https://goreportcard.com/badge/github.com/dustinblackman/collectd-docker-plugin"></a> <img src="https://img.shields.io/github/release/dustinblackman/collectd-docker-plugin.svg?maxAge=2592000">

Collectd plugin to tap in the Docker Stats streaming API using Collectd's [Exec](https://collectd.org/wiki/index.php/Plugin:Exec) plugin. Built with Go 1.7 and tested with Collectd 5.5 and Influx 1.0.

## Installation

Example installation for a Ubuntu system. Make changes required to match your own OS.

```bash
curl -Ls "https://github.com/dustinblackman/collectd-docker-plugin/releases/download/0.0.1/collectd-docker-plugin-linux-amd64-0.0.1.tar.gz" | tar xz -C /usr/local/bin/
curl -o /usr/share/collectd https://github.com/dustinblackman/collectd-docker-plugin/blob/master/collectd/docker.db
curl -o /etc/collectd/collectd.conf.d https://github.com/dustinblackman/collectd-docker-plugin/blob/master/collectd/docker.conf
usermod -a -G docker nobody
service collectd restart
```

## Parameters
Parameters are available and can be added by modifying [docker.conf](./collectd/docker.conf) and appending parameters to the exec function.


- `-d, --docker-host` - Docker socket path. Defaults to `unix:///var/run/docker.sock`
- `-de, --docker-environment` - Boolean parameter to specifiy reading Docker parameters from environment variables
- `-ch, --collectd-hostname` - Collectd hostname. This is automatically provided to the process from Collectd.
- `-w, --wait-time` - Delay in seconds with how often metrics are submitted to Collectd. Defaults to 3.

## Build From Source

Tested with Go 1.7. Versioning is done with [Glide](https://github.com/Masterminds/glide). The makefile will take care of installing it for you incase you don't have it.

```bash
git pull https://github.com/dustinblackman/collectd-docker-plugin
cd collectd-docker-plugin
make install
```

## [License](./LICENSE)
