package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/user"
	"regexp"
	"strings"

	"golang.org/x/net/context"

	"github.com/BurntSushi/toml"
	"github.com/nsyntych/go-socks5"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

//Config keeps the root of the confi file
type Config struct {
	Proxies map[string]*SSHProxy
	Filters []*URLFilter
}

// SSHProxy keeps data of an SSH proxy
type SSHProxy struct {
	Host        string
	Port        string
	User        string
	Parent      string
	ParentProxy *SSHProxy
	Key         string
	Password    string
	Variable    string
	Client      *ssh.Client
}

// URLFilter keeps url to proxy maping
type URLFilter struct {
	URL     string
	Proxy   string
	Pattern *regexp.Regexp
}

func formatUserDir(path string) string {
	usr, _ := user.Current()
	homeDir := usr.HomeDir
	return strings.Replace(path, "~", homeDir, 1)
}

// ReadConfig reads info from config file
func ReadConfig(configfile string) Config {
	_, err := os.Stat(configfile)
	if err != nil {
		log.Fatal("Config file is missing: ", configfile)
	}

	var config Config
	if _, err := toml.DecodeFile(configfile, &config); err != nil {
		log.Fatal(err)
	}

	// Format and compile patters
	for _, filter := range config.Filters {
		pattern := filter.URL
		if strings.HasPrefix(pattern, ".") {
			pattern = ".*" + strings.TrimPrefix(pattern, ".")
		}
		if strings.HasSuffix(pattern, ".") {
			pattern = pattern + "*"
		}
		filter.Pattern, err = regexp.Compile(pattern)
		if err != nil {
			println(err)
		}
	}

	for _, proxy := range config.Proxies {
		if parentProxy, ok := config.Proxies[proxy.Parent]; ok {
			proxy.ParentProxy = parentProxy
		}
	}

	//log.Print(config.Index)
	return config
}

// PublicKeyFile loads SSH private key file
func PublicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}

// ProxySSHConnect connects the ssh proxy
func ProxySSHConnect(proxy *SSHProxy) error {

	sshConfig := &ssh.ClientConfig{
		User:            proxy.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	if len(proxy.Variable) > 0 {
		sshConfig.Auth = []ssh.AuthMethod{ssh.Password(os.Getenv(proxy.Variable))}
	} else if len(proxy.Key) > 0 {
		sshConfig.Auth = []ssh.AuthMethod{PublicKeyFile(formatUserDir(proxy.Key))}
	} else if len(proxy.Password) > 0 {
		sshConfig.Auth = []ssh.AuthMethod{ssh.Password(proxy.Password)}
	} else {
		fmt.Printf("Please enter SSH password for %s: ", proxy.Host)
		password, _ := terminal.ReadPassword(0)
		sshConfig.Auth = []ssh.AuthMethod{ssh.Password(string(password))}
	}

	var err error
	if proxy.ParentProxy != nil {

		println("Trying to connect to", proxy.Host, "through", proxy.ParentProxy.Host)

		// Dial a connection to the service host, from the parentProxy
		conn, err := proxy.ParentProxy.Client.Dial("tcp", proxy.Host+":"+proxy.Port)
		if err != nil {
			log.Println(err)
			return err
		}

		ncc, chans, reqs, err := ssh.NewClientConn(conn, proxy.Host+":"+proxy.Port, sshConfig)
		if err != nil {
			log.Println(err)
			return err
		}

		proxy.Client = ssh.NewClient(ncc, chans, reqs)

	} else {
		proxy.Client, err = ssh.Dial("tcp", proxy.Host+":"+proxy.Port, sshConfig)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	println("Successfuly connected to", proxy.Host)
	return nil

}

func main() {

	configfile := flag.String("c", "proxy_config.toml", "TOML Configuration file path")
	bind_host := flag.String("h", "127.0.0.1", "The bind host of the SOCKS5 server")
	bind_port := flag.Int("p", 8000, "The bind port of the SOCKS5 server")

	flag.Parse()
	proxyConfig := ReadConfig(*configfile)

	// fmt.Printf("%# v\n", pretty.Formatter(proxyConfig))

	for _, proxy := range proxyConfig.Proxies {
		err := ProxySSHConnect(proxy)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
	}

	// Create a SOCKS5 server
	conf := &socks5.Config{}

	conf.Dial = func(ctx context.Context, network string, addr *socks5.AddrSpec) (net.Conn, error) {

		print(addr.FQDN, addr.Address())

		for _, filter := range proxyConfig.Filters {
			if filter.Pattern.MatchString(addr.FQDN) {
				if proxyConfig.Proxies[filter.Proxy].Client != nil {
					println(" ->", proxyConfig.Proxies[filter.Proxy].Host)
					return proxyConfig.Proxies[filter.Proxy].Client.Dial(network, addr.Address())
				}
				println("Nil Proxy!")
			}
		}
		println(" -> None")
		return net.Dial(network, addr.Address())
	}

	server, err := socks5.New(conf)
	if err != nil {
		panic(err)
	}

	// Create SOCKS5 proxy on listen_address
	listen_address := fmt.Sprintf("%s:%d", *bind_host, *bind_port)
	println("SOCKS5 server on", listen_address)
	if err := server.ListenAndServe("tcp", listen_address); err != nil {
		panic(err)
	}
}
