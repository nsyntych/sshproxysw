# sshproxysw
An SSH Proxy SOCK5 Switch

### How to
```bash
# Get and install sshproxysw
go get github.com/nsyntych/sshproxysw
go install github.com/nsyntych/sshproxysw

# Copy and edit the proxy TOML conf file
cp $GOPATH/src/github.com/nsyntych/sshproxysw/proxy.example.toml /wherever/you/want/proxy.toml
edit /wherever/you/want/proxy.toml # follow examples inside

# Run
$GOPATH/bin/sshproxysw -c /wherever/you/want/proxy.toml
```