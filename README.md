sshproxysw [![Build Status](https://travis-ci.org/nsyntych/sshproxysw.svg?branch=master)](https://travis-ci.org/nsyntych/sshproxysw)
=========
An SSH Proxy SOCK5 Switch


How to run
=========
Follow instructions below to run a SOCKS5 proxy server locally that switches traffic though SSH connections according to the configuration provided

```bash
# Get and install sshproxysw
go get github.com/nsyntych/sshproxysw
go install github.com/nsyntych/sshproxysw

# Copy and edit the proxy TOML conf file
cp $GOPATH/src/github.com/nsyntych/sshproxysw/proxy.example.toml /wherever/you/want/proxy.toml

# Edit conf by following the examples inside
edit /wherever/you/want/proxy.toml

# Run
$GOPATH/bin/sshproxysw -c /wherever/you/want/proxy.toml -h 127.0.0.1 -p 8000
```