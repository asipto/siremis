package main

import (
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// {"jsonrpc": "2.0", "method": "command", "params": [p1, "p2"], "id": 1}
type GMJSONRPCMethod struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  []any  `json:"Params,omitempty"`
	Id      int    `json:"id"`
}

func GMJSONRPCExec(sCommand string) (string, bool) {
	if len(GMConfigV.JSONRPC.Protocol) == 0 {
		return "", false
	}
	sArr := strings.SplitN(sCommand, " ", 2)
	sMethod := sArr[0]
	sParams := ""
	if len(sArr) > 1 {
		sParams = sArr[1]
	}
	var jCmd = GMJSONRPCMethod{}
	jCmd.JSONRPC = "2.0"
	jCmd.Method = sMethod
	jCmd.Id = 1 + rand.Intn(1000000)
	if len(sParams) > 0 {
		pTokens := strings.Split(sParams, " ")
		if len(pTokens) > 0 {
			for _, v := range pTokens {
				if strings.HasPrefix(v, "i:") || strings.HasPrefix(v, "n:") {
					if n, err := strconv.Atoi(v[2:]); err == nil {
						jCmd.Params = append(jCmd.Params, n)
					} else {
						jCmd.Params = append(jCmd.Params, v)
					}
				} else if strings.HasPrefix(v, "s:") {
					jCmd.Params = append(jCmd.Params, v[2:])
				} else {
					if n, err := strconv.Atoi(v); err == nil {
						jCmd.Params = append(jCmd.Params, n)
					} else {
						jCmd.Params = append(jCmd.Params, v)
					}
				}
			}
		}
	}
	bCmd, err := json.Marshal(&jCmd)
	if err != nil {
		log.Printf("failed to generate jsonrpc command: '%s'\n", sMethod)
		return "", false
	}
	if GMConfigV.JSONRPC.Protocol == "udp" {
		log.Printf("udp - jsonrpc command: '%s'\n", sMethod)
		LocalAddr, err := net.ResolveUDPAddr(GMConfigV.JSONRPC.Protocol, GMConfigV.JSONRPC.LAddress)
		if err != nil {
			log.Printf("failed to resolve udp local address: '%s'\n", GMConfigV.JSONRPC.LAddress)
			return "", false
		}
		RemoteAddr, err := net.ResolveUDPAddr(GMConfigV.JSONRPC.Protocol, GMConfigV.JSONRPC.RAddress)
		if err != nil {
			log.Printf("failed to resolve udp remote address: '%s'\n", GMConfigV.JSONRPC.RAddress)
			return "", false
		}
		conn, err := net.DialUDP(GMConfigV.JSONRPC.Protocol, LocalAddr, RemoteAddr)
		if err != nil {
			log.Printf("failed to connect to udp remote address: '%s'\n", GMConfigV.JSONRPC.RAddress)
			return "", false
		}
		conn.Write(bCmd)
		defer conn.Close()
		bRes, err := io.ReadAll(conn)
		if err != nil {
			log.Printf("failed to read from udp remote address: '%s'\n", GMConfigV.JSONRPC.RAddress)
			return "", false
		}
		return string(bRes), true
	} else if GMConfigV.JSONRPC.Protocol == "unixgram" {
		log.Printf("unixgram - jsonrpc command: '%s'\n", sMethod)
		// remove any existing local socket
		os.Remove(GMConfigV.JSONRPC.LAddress)
		LocalAddr, err := net.ResolveUnixAddr(GMConfigV.JSONRPC.Protocol, GMConfigV.JSONRPC.LAddress)
		if err != nil {
			log.Printf("failed to resolve unixgram local address: '%s'\n", GMConfigV.JSONRPC.LAddress)
			return "", false
		}

		RemoteAddr, err := net.ResolveUnixAddr(GMConfigV.JSONRPC.Protocol, GMConfigV.JSONRPC.RAddress)
		if err != nil {
			log.Printf("failed to resolve unixgram remote address: '%s'\n", GMConfigV.JSONRPC.RAddress)
			return "", false
		}
		rConn, err := net.DialUnix(GMConfigV.JSONRPC.Protocol, LocalAddr, RemoteAddr)
		if err != nil {
			log.Printf("failed to connect to unixgram remote address: '%s' (%v)\n",
				GMConfigV.JSONRPC.RAddress, err)
			return "", false
		}
		defer rConn.Close()
		defer os.Remove(GMConfigV.JSONRPC.LAddress)

		rConn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		rConn.SetReadDeadline(time.Now().Add(5 * time.Second))

		_, err = rConn.Write(bCmd)
		if err != nil {
			log.Printf("failed to write to unixgram remote address: '%s' (%v)\n",
				GMConfigV.JSONRPC.RAddress, err)
			return "", false
		}
		bRes := make([]byte, 8*1024)
		_, err = rConn.Read(bRes)
		if err != nil {
			log.Printf("failed to read from unixgram remote address: '%s' (%v)\n",
				GMConfigV.JSONRPC.RAddress, err)
			return "", false
		}
		return string(bRes), true
	}
	return "", false
}
