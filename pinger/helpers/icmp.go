package helpers

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

const (
	protocolICMP   = 1  // IPv4 ICMP protocol number
	protocolICMPv6 = 58 // IPv6 ICMP protocol number
	pingDataSize   = 64 // Standard ping packet size
	defaultTTL     = 64 // Default time to live
)

// ICMPInfo is everything user - configurable of a PINGER
type ICMPInfo struct {
	IP    string
	Iface string
	TTL   int
	CNT   int
}

// getInterface checks if interfaceName device exists,
// and returns a pointer to it if it does, else nil
func getInterface(interfaceName string) *net.Interface {
	var hostIface *net.Interface

	if interfaceName == "" {
		return nil
	}

	hostIface, err := net.InterfaceByName(interfaceName)
	if err != nil {
		fmt.Printf("Error finding interface %s: %v\n", interfaceName, err)
		os.Exit(1)
	}

	return hostIface
}

// constructMarshalledMessage handles populating icmp.Message struct,
// and marshalls it into []binary, to send on the wire
func constructMarshalledMessage(msgType icmp.Type, seqNum int) ([]byte, error) {
	// Construct message
	request := icmp.Message{
		Type: msgType,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  seqNum,
			Data: []byte("Never stop learning because life never stops teaching!!!"),
		},
	}

	// Marshal the message
	binRequest, err := request.Marshal(nil)

	return binRequest, err
}

// sendICMPRequest sends the request v4/6 Echo Request to the given ipaddr, via the given iface,
// using the "icmp socket" conn
func sendICMPRequest(ipaddr string, iface *net.Interface, ifaceName string, conn *icmp.PacketConn, request []byte, proto int) (time.Time, error) {

	var (
		start time.Time
		err   error
	)
	destination := &net.IPAddr{IP: net.ParseIP(ipaddr), Zone: ifaceName}

	switch proto {
	case protocolICMP:
		var controlRequest ipv4.ControlMessage

		if iface != nil {
			controlRequest.IfIndex = iface.Index
			_, err = conn.IPv4PacketConn().WriteTo(request, &controlRequest, destination)
		} else {
			_, err = conn.WriteTo(request, destination)
		}

		start = time.Now()

	case protocolICMPv6:
		var controlRequest ipv6.ControlMessage
		if iface != nil {
			controlRequest.IfIndex = iface.Index
			_, err = conn.IPv6PacketConn().WriteTo(request, &controlRequest, destination)
		} else {
			_, err = conn.WriteTo(request, destination)
		}
		start = time.Now()

	}

	return start, err
}

// recvICMPRequest receives the v4/6 Echo Reply from the given "icmp socket" conn
// and *immediately* calculates the elapsed time since sending the Echo Request
func recvICMPRequest(startTime time.Time, proto int, conn *icmp.PacketConn) ([]byte, float64, int, net.Addr, error) {

	var (
		receivedTTL = defaultTTL
		numBytes    int
		binReply    = make([]byte, 1500)
		peerAddr    net.Addr
		err         error
	)

	switch proto {
	case protocolICMP:
		// Read ttl from reply IP header
		// Handled by this control message
		var controlMessage *ipv4.ControlMessage

		// Receive response
		numBytes, controlMessage, peerAddr, err = conn.IPv4PacketConn().ReadFrom(binReply)

		if controlMessage != nil {
			receivedTTL = controlMessage.TTL
		}

	case protocolICMPv6:
		// Read ttl from reply IP header
		// Handled by this control message
		var controlMessage *ipv6.ControlMessage
		// Receive response
		numBytes, controlMessage, peerAddr, err = conn.IPv6PacketConn().ReadFrom(binReply)
		if controlMessage != nil {
			receivedTTL = controlMessage.HopLimit
		}
	}

	// End timer
	elapsed := time.Since(startTime)
	elapsedMs := float64(elapsed.Microseconds()) / 1000.0 // Convert to milliseconds

	return binReply[:numBytes], elapsedMs, receivedTTL, peerAddr, err
}

// printICMPResponse handles the different types of ICMP replies received
func printICMPResponse(proto int, data []byte, peer net.Addr, seq int, receivedTTL int, elapsedMs float64, stats *PingStats) {

	// Parse the response
	reply, err := icmp.ParseMessage(proto, data)
	if err != nil {
		fmt.Printf("Error parsing ICMP response: %v\n", err)
		stats.errors++
		return
	}

	switch reply.Type {
	// Expected case
	case ipv4.ICMPTypeEchoReply, ipv6.ICMPTypeEchoReply:
		// data parse
		echo, ok := reply.Body.(*icmp.Echo)
		if !ok {
			fmt.Printf("Invalid ICMP echo reply\n")
			stats.errors++
			return
		}

		stats.received++
		// valid receipt => update statistics
		stats.iterativeStats(elapsedMs)
		// Print to stdout
		fmt.Printf("%d bytes from %s: icmp_seq=%d ttl=%d time=%.3f ms\n",
			len(data), peer.String(), echo.Seq, receivedTTL, elapsedMs)

	case ipv4.ICMPTypeDestinationUnreachable, ipv6.ICMPTypeDestinationUnreachable:

		// error receipt => do nothing
		stats.errors++
		// Print to stdout
		fmt.Printf("From %s icmp_seq=%d: Destination Host Unreachable\n",
			peer.String(), seq)

	case ipv4.ICMPTypeTimeExceeded, ipv6.ICMPTypeTimeExceeded:

		// error receipt => do nothing
		stats.errors++
		// Print to stdout
		if proto == protocolICMP {
			fmt.Printf("From %s icmp_seq=%d: Time To Live Exceeded\n",
				peer.String(), seq)
		} else {
			fmt.Printf("From %s icmp_seq=%d: Hop Limit Exceeded\n",
				peer.String(), seq)
		}

	case ipv6.ICMPTypeNeighborAdvertisement, ipv6.ICMPTypeNeighborSolicitation, ipv6.ICMPTypeRouterAdvertisement, ipv6.ICMPTypeRouterSolicitation:

		//BUG: Unknown if the actual ICMPv6 reply is lost, due to this meta control-message received
		fmt.Printf("From %s icmp_seq=%d: IPv6 specific information: %v\n",
			peer.String(), seq, reply.Type)

	default:
		// Uncaught error...
		stats.errors++
		// Print to stdout
		fmt.Printf("From %s icmp_seq=%d: ICMP type: %v\n",
			peer.String(), seq, reply.Type)
	}
}

// printReadError is an error handler for receiving the ICMP(4/6) reply
func printReadError(err error, seq int, stats *PingStats) {
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		fmt.Printf("Request timeout for icmp_seq %d\n", seq)
		stats.errors++
	} else {
		fmt.Printf("Error reading ICMP response: %v\n", err)
		stats.errors++
	}
}

// ICMP6Handler handles PINGER when using AF_INET6
func ICMP6Handler(info ICMPInfo) {
	// iteratively calculated statistics
	stats := PingStats{min: -1}

	// returned pointer may be nil...
	var hostIface *net.Interface = getInterface(info.Iface)

	// Set up signal handling for graceful termination: usual ending with Ctl + C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		stats.finalStats()
		PrintStatistics(&stats)
		os.Exit(0)
	}()

	// Start pinging
	if info.Iface != "" {
		fmt.Printf("PINGERING %s: %d data bytes (via %s)\n", info.IP, pingDataSize, info.Iface)
	} else {
		fmt.Printf("PINGERING %s: %d data bytes\n", info.IP, pingDataSize)
	}

	// abstracted "socket" information
	var (
		proto      int    = protocolICMPv6
		network    string = "ip6:ipv6-icmp"
		listenAddr string = "::"
	)

	// setup one end of connection
	conn, err := icmp.ListenPacket(network, listenAddr)
	if err != nil {
		fmt.Printf("Error creating ICMPv6 connection: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	// Set Hop Limit
	conn.IPv6PacketConn().SetHopLimit(info.TTL)
	// **Set control message flags to receive hop limit info**
	conn.IPv6PacketConn().SetControlMessage(ipv6.FlagHopLimit|ipv6.FlagInterface, true)

	//Send ICMPv6 packet loop
	for i := range info.CNT {
		stats.transmitted++

		// Construct the required message
		request, err := constructMarshalledMessage(ipv6.ICMPTypeEchoRequest, i)
		if err != nil {
			fmt.Printf("Error generating ICMP message: %v\n", err)
			stats.errors++
			continue
		}

		// Set read deadline
		// Linux has it as 1 second, MS as 4 seconds, so may modify...
		conn.SetReadDeadline(time.Now().Add(4 * time.Second))

		startTime, err := sendICMPRequest(info.IP, hostIface, info.Iface, conn, request, proto)
		if err != nil {
			fmt.Printf("Error sending ICMP packet: %v\n", err)
			stats.errors++
			continue
		}

		// Receive the required response
		reply, elapsedMs, receivedTTL, peerAddr, err := recvICMPRequest(startTime, proto, conn)
		if err != nil {
			printReadError(err, i, &stats)
			continue
		}

		//Format what was received
		printICMPResponse(proto, reply, peerAddr, i, receivedTTL, elapsedMs, &stats)
		time.Sleep(time.Second)
	}

	stats.finalStats()
	PrintStatistics(&stats)
}

// ICMP4Handler handles PINGER when using AF_INET
func ICMP4Handler(info ICMPInfo) {

	stats := PingStats{}

	// returned pointer may be nil...
	var hostIface *net.Interface = getInterface(info.Iface)

	// Set up signal handling for graceful termination: usual ending with Ctl + C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		stats.finalStats()
		PrintStatistics(&stats)
		os.Exit(0)
	}()

	// Start pinging
	if info.Iface != "" {
		fmt.Printf("PINGERING %s: %d data bytes (via %s)\n", info.IP, pingDataSize, info.Iface)
	} else {
		fmt.Printf("PINGERING %s: %d data bytes\n", info.IP, pingDataSize)
	}

	// abstracted "socket" information
	var (
		proto      int    = protocolICMP
		network    string = "ip4:icmp"
		listenAddr string = "0.0.0.0"
	)

	// setup one end of connection
	conn, err := icmp.ListenPacket(network, listenAddr)
	if err != nil {
		fmt.Printf("Error creating ICMP connection: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	// Set TTL
	conn.IPv4PacketConn().SetTTL(info.TTL)

	// **Set control message flags to receive TTL info**
	conn.IPv4PacketConn().SetControlMessage(ipv4.FlagTTL|ipv4.FlagInterface, true)

	//Send ICMPv4 packet loop
	for i := range info.CNT {
		stats.transmitted++

		// Construct the required message
		request, err := constructMarshalledMessage(ipv4.ICMPTypeEcho, i)
		if err != nil {
			fmt.Printf("Error generating ICMP message: %v\n", err)
			stats.errors++
			continue
		}

		// Set read deadline
		// Linux has it as 1 second, MS as 4 seconds, so may modify...
		//TODO: Change location
		conn.SetReadDeadline(time.Now().Add(4 * time.Second))

		startTime, err := sendICMPRequest(info.IP, hostIface, info.Iface, conn, request, proto)

		if err != nil {
			fmt.Printf("Error sending ICMP packet: %v\n", err)
			stats.errors++
			continue
		}

		// Receive the required response
		reply, elapsedMs, receivedTTL, peerAddr, err := recvICMPRequest(startTime, proto, conn)
		if err != nil {
			printReadError(err, i, &stats)
			continue
		}

		//Format what was received
		printICMPResponse(proto, reply, peerAddr, i, receivedTTL, elapsedMs, &stats)
		time.Sleep(time.Second)
	}

	stats.finalStats()
	PrintStatistics(&stats)
}
