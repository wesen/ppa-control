package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"ppa-control/lib/protocol"
	"time"

	"github.com/augustoroman/hexdump"
)

const maxBufferSize = 1024
const timeout = 10 * time.Second

func server(ctx context.Context, address string) (err error) {
	pc, err := net.ListenPacket("udp", address)
	if err != nil {
		return
	}
	defer pc.Close()

	doneChan := make(chan error, 1)
	buffer := make([]byte, maxBufferSize)

	go func() {
		for {
			fmt.Printf("server: waiting\n")
			n, addr, err := pc.ReadFrom(buffer)
			if err != nil {
				doneChan <- err
				return
			}

			fmt.Printf("server: packet-received: bytes=%d from=%s\n", n, addr.String())
			deadline := time.Now().Add(timeout)
			err = pc.SetWriteDeadline(deadline)
			if err != nil {
				doneChan <- err
				return
			}

			n, err = pc.WriteTo(buffer[:n], addr)
			if err != nil {
				doneChan <- err
				return
			}
			fmt.Printf("server: packet-written: bytes=%d to=%s\nserver: %s\n",
				n, addr.String(),
				hexdump.Dump(buffer[:n]))
		}
	}()

	select {
	case <-ctx.Done():
		fmt.Println("server: cancelled")
		err = ctx.Err()
	case err = <-doneChan:
		if err != nil {
			fmt.Printf("server: got error: %s\n", err)
		}
	}

	return
}

func client(ctx context.Context, address string, reader io.Reader) (err error) {
	raddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return
	}
	defer conn.Close()

	doneChan := make(chan error, 1)

	go func() {
		n, err := io.Copy(conn, reader)
		if err != nil {
			doneChan <- err
			return
		}
		fmt.Printf("client: packet-written: bytes=%d\n", n)
		buffer := make([]byte, maxBufferSize)
		deadline := time.Now().Add(timeout)
		err = conn.SetReadDeadline(deadline)
		if err != nil {
			doneChan <- err
			return
		}

		nRead, addr, err := conn.ReadFrom(buffer)
		if err != nil {
			doneChan <- err
			return
		}

		fmt.Printf("client: packet-received: bytes=%d from=%s\nclient: %s\n",
			nRead, addr.String(), hexdump.Dump(buffer[:nRead]))
		doneChan <- nil
	}()

	select {
	case <-ctx.Done():
		fmt.Println("client: cancelled")
		err = ctx.Err()
	case err = <-doneChan:
		if err != nil {
			fmt.Printf("client: got error: %s\n", err)
		}
	}

	return
}

var (
	address        = flag.String("address", "127.0.0.1", "server address")
	port           = flag.Uint("port", 5151, "server port")
	runServer      = flag.Bool("run-server", false, "Run as server too")
	presetPosition = flag.Int("position", 1, "preset")
	componentId    = flag.Int("component-id", 0xff, "component ID (default: 0xff)")
)

func main() {
	flag.Parse()

	ctx := context.Background()
	serverString := fmt.Sprintf("%s:%d", *address, *port)
	fmt.Printf("Connecting to %s\n", serverString)

	if *runServer {
		fmt.Printf("Starting test server")
		go server(ctx, serverString)
		time.Sleep(1 * time.Second)
	}

	seqCmd := uint16(1)

	for {
		buf := new(bytes.Buffer)
		bh := protocol.NewBasicHeader(
			protocol.MessageTypePresetRecall,
			protocol.StatusCommand,
			[4]byte{0, 0, 0, 0},
			seqCmd,
			byte(*componentId),
		)
		pr := protocol.NewPresetRecall(protocol.RecallByPresetIndex, 0, byte(*presetPosition))
		seqCmd += 1

		protocol.EncodeHeader(buf, bh)
		protocol.EncodePresetRecall(buf, pr)

		client(ctx, serverString, buf)
		time.Sleep(1 * time.Second)
	}
}
