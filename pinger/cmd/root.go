package cmd

import (
	"fmt"
	"os"

	"github.com/Vishy70/custom-ping-utility-Vishy70/pinger/helpers"
	"github.com/spf13/cobra"
)

var (
	v4Flag    bool
	v6Flag    bool
	ifaceFlag string
	ttlFlag   int8
	cntFlag   int8
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "pinger",
	Short: "Pinger- A custom ping clone to send ICMP ECHO_REQUEST to network hosts, built in Golang",
	Long: `pinger is a custom ping clone to send ICMP ECHO_REQUEST to network hosts, built in Golang! 
It supports: 
- IPv4, IPv6 [-4|-6]
- Sending to a specific network interface[-I <iface-name>]
- Number of echo requests [-c <number>]
- Setting Time to Live [-t <ttl>].`,
	Args: cobra.ExactArgs(1),
	Example: `./pinger -I wlp45s0 -c 4 -4 nitk.ac.in

(You will likely need root privileges, since pinger opens raw sockets...)`,
	// Single action for this application
	Run: func(cmd *cobra.Command, args []string) {
		addr := args[0]

		addrOptions := helpers.AddrOptions{
			V4: v4Flag,
			V6: v6Flag,
		}

		verified, err := helpers.AddrResolution(addr, addrOptions)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		ipaddr, isIPv6 := verified.Addr, verified.IsIPv6

		icmpInfo := helpers.ICMPInfo{
			IP:    ipaddr,
			Iface: ifaceFlag,
			TTL:   int(ttlFlag),
			CNT:   int(cntFlag),
		}

		if !isIPv6 {
			helpers.ICMP4Handler(icmpInfo)
		} else {
			helpers.ICMP6Handler(icmpInfo)
		}

	},
}

// Adds all child commands to the root command and sets flags appropriately. Called by main.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {

	rootCmd.PersistentFlags().BoolVarP(&v4Flag, "ipv4", "4", false, "Use IPv4 for address / hostname resolution")
	rootCmd.PersistentFlags().BoolVarP(&v6Flag, "ipv6", "6", false, "Use IPv6 for address / hostname resolution")
	rootCmd.PersistentFlags().StringVarP(&ifaceFlag, "iface", "I", "", "Specify the network device name")
	rootCmd.PersistentFlags().Int8VarP(&ttlFlag, "ttl", "t", 64, "Define the time to live")
	rootCmd.PersistentFlags().Int8VarP(&cntFlag, "count", "c", 5, "Stop after <count tries>")
}
