/*
minimal, read-only TFTP server baed in pin/tftp library
  https://github.com/pin/tftp/pull/67
- basically imported the Example from the README.md
- enable single port mode for the server to work on Fly.io
- prevent Dir tree / path traversal aka "dot-dot-dash" attacks.
  Use simple and pragmatic solution, as we serve very few files from
  a single directory ("prefix") and one/few extensions ("suffixes") only:
	1. cut all prefixes and prepend our prefix /app/tftp
	2. check suffxes using allowedSuffixes from teaftp:
*/

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	// single-port mode option requires v2.2.0+
	// $ go get github.com/pin/tftp@2.2.0		// fails to get v2.2.0
	// $ go get github.com/pin/tftp@master		// work around until fixed upstream
	// see https://github.com/pin/tftp/pull/67
	"github.com/pin/tftp"
	//"github.com/pin/tftp@2.2.0"				// fails
	//"github.com/pin/tftp/v2"					// once PR#67 is committed upstream
)

var allowedSuffixes = []string{".md", ".mod", ".kpxe", ".txt"}

// checks if a string is present in a slice, is case sensitive
//  https://freshman.tech/snippets/go/check-if-slice-contains-element/
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

// readHandler is called when client starts file download from server
func readHandler(filenameRRQ string, rf io.ReaderFrom) error {
	// print/log Local and Remote Address
	//  https://github.com/pin/tftp#local-and-remote-address
	raddr := rf.(tftp.OutgoingTransfer).RemoteAddr()
	laddr := rf.(tftp.RequestPacketInfo).LocalIP()

	// chop off path & enforce it for file to read & served
	filename := filepath.Join("/app/tftp", filepath.Base(filepath.Clean(filenameRRQ)))

	fmt.Printf("RRQ %s > %s from %s  to %s \n", filenameRRQ, filename, raddr.String(), laddr.String())

	// Check if the read request is allowed, or not
	if !contains(allowedSuffixes, filepath.Ext(filename)) {
		fmt.Fprintf(os.Stderr, "[DENIED] RRQ %s : suffix not allowed\n", filenameRRQ)
		return fmt.Errorf("[DENIED] RRQ %s ", filenameRRQ)
	}

	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
		return err
	}
	n, err := rf.ReadFrom(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
		return err
	}
	fmt.Printf("[SENT] %d bytes\n", n)
	return nil
}

func main() {
	// use nil in place of handler to disable read or write operations
	//s := tftp.NewServer(readHandler, writeHandler)
	s := tftp.NewServer(readHandler, nil) // read-only server
	s.SetTimeout(5 * time.Second)         // optional
	// enable single-port mode (experimental)
	//  https://github.com/pin/tftp/blob/0161c5dd2e967493da88cfdf9426b9337afb60ee/server.go#L112
	s.EnableSinglePort()
	s.SetBlockSize(8192) // 512 (default) to 65465, advisory only, clamped by client, MTU, & single-port mode to 1372!
	//	err := s.ListenAndServe(":69")		// blocks until s.Shutdown() is called
	//	err := s.ListenAndServe(":6969")	// blocks until s.Shutdown() is called
	// special for UDP on Fly.io, must bind to addr fly-global-services, see
	//  https://fly.io/docs/app-guides/udp-and-tcp/#let-s-see-some-code
	//err := s.ListenAndServe(fmt.Sprintf("fly-global-services:%d", 6969))	// blocks until s.Shutdown() is called
	err := s.ListenAndServe("fly-global-services:69") // blocks until s.Shutdown() is called
	//	err := s.ListenAndServe("fly-global-services:6969")	// blocks until s.Shutdown() is called
	if err != nil {
		fmt.Fprintf(os.Stdout, "[ERROR] tftServer: %v\n", err)
		os.Exit(1)
	}
}
