package main

import (
	"fmt"
	"os"

	"github.com/MattAitchison/env"
	"github.com/getsentry/raven-go"
	"github.com/spf13/cobra"
)

var (
	Version   string
	SentryDSN string

	providerName = env.String("HOSTCTL_PROVIDER", "digitalocean", "cloud provider")
	defaultName  = env.String("HOSTCTL_NAME", "", "optional default name")
	namespace    = env.String("HOSTCTL_NAMESPACE", "", "optional namespace for names")

	hostImage    = env.String("HOSTCTL_IMAGE", "", "vm image")
	hostFlavor   = env.String("HOSTCTL_FLAVOR", "", "vm flavor")
	hostRegion   = env.String("HOSTCTL_REGION", "", "vm region")
	hostKeyname  = env.String("HOSTCTL_KEYNAME", "", "vm keyname")
	hostUserdata = env.String("HOSTCTL_USERDATA", "", "vm user data")

	user = env.String("HOSTCTL_USER", os.Getenv("USER"), "ssh user")
)

func init() {
	raven.SetDSN(SentryDSN)
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			capture(r)
			panic(r)
		}
	}()
	fatal(Hostctl.Execute())
}

var Hostctl = &cobra.Command{
	Use:   "hostctl",
	Short: "An opinionated tool for provisioning cloud VMs",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(Version)
	},
}
