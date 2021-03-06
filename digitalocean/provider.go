package digitalocean

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"code.google.com/p/goauth2/oauth"
	"github.com/MattAitchison/env"
	"github.com/digitalocean/godo"
	"github.com/gliderlabs/hostctl/providers"
)

var envSet = env.NewEnvSet("digitalocean")

func init() {
	provider := &digitalOceanProvider{
		token: envSet.Secret("DO_TOKEN", "token for DigitalOcean API v2"),
	}
	providers.Register(provider, "digitalocean")
}

type digitalOceanProvider struct {
	client *godo.Client
	token  string
}

func (p *digitalOceanProvider) Setup() error {
	if p.token == "" {
		return fmt.Errorf("DO_TOKEN required for Digital Ocean provider")
	}
	t := &oauth.Transport{
		Token: &oauth.Token{AccessToken: p.token},
	}
	p.client = godo.NewClient(t.Client())
	_, _, err := p.client.Account.Get()
	return err
}

func (p *digitalOceanProvider) Env() *env.EnvSet {
	return envSet
}

func (p *digitalOceanProvider) Create(host providers.Host) error {
	var sshKey godo.DropletCreateSSHKey
	if strings.Contains(host.Keyname, ":") {
		sshKey.Fingerprint = host.Keyname
	} else {
		id, err := strconv.Atoi(host.Keyname)
		if err != nil {
			return err
		}
		sshKey.ID = id
	}
	droplet, _, err := p.client.Droplets.Create(&godo.DropletCreateRequest{
		Name:   host.Name,
		Region: host.Region,
		Size:   host.Flavor,
		Image: godo.DropletCreateImage{
			Slug: host.Image,
		},
		SSHKeys:  []godo.DropletCreateSSHKey{sshKey},
		UserData: host.Userdata,
	})
	if err != nil {
		return err
	}
	for {
		droplet, _, err = p.client.Droplets.Get(droplet.ID)
		if droplet != nil && droplet.Status == "active" {
			return nil
		}
		if err != nil {
			return err
		}
		time.Sleep(1 * time.Second)
	}
}

func (p *digitalOceanProvider) Destroy(name string) error {
	droplets, _, err := p.client.Droplets.List(nil)
	if err != nil {
		return err
	}
	for i := range droplets {
		if droplets[i].Name == name {
			_, err := p.client.Droplets.Delete(droplets[i].ID)
			if err != nil {
				return err
			}
			// TODO timeout
			for p.Get(name) != nil {
				time.Sleep(1 * time.Second)
			}
			return nil
		}
	}
	return nil
}

func (p *digitalOceanProvider) List(pattern string) []providers.Host {
	droplets, _, err := p.client.Droplets.List(nil)
	if err != nil {
		return nil
	}
	var hosts []providers.Host
	for i := range droplets {
		if ok, _ := filepath.Match(pattern, droplets[i].Name); ok {
			hosts = append(hosts, providers.Host{
				Name: droplets[i].Name,
			})
		}
	}
	return hosts
}

func (p *digitalOceanProvider) Get(name string) *providers.Host {
	droplets, _, err := p.client.Droplets.List(nil)
	if err != nil {
		return nil
	}
	for i := range droplets {
		if droplets[i].Name == name {
			var ip string
			if droplets[i].Networks != nil {
				if len(droplets[i].Networks.V4) > 0 {
					ip = droplets[i].Networks.V4[0].IPAddress
				}
			}
			return &providers.Host{
				Name: name,
				IP:   ip,
			}
		}
	}
	return nil
}
