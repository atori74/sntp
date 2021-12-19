package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

func main() {
	// sntp client

	if len(os.Args) != 2 {
		log.Println("invalid args")
		return
	}

	serverIP := os.Args[1]
	now, err := sntpNow(serverIP)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(now)
}

type ntpTimestamp struct {
	originateTimestamp   *time.Time
	transmitTimestamp    *time.Time
	receiveTimestamp     *time.Time
	destinationTimestamp *time.Time
}

func sntpNow(serverIP string) (*time.Time, error) {
	udpAddr, err := net.ResolveUDPAddr("udp4", serverIP+":123")
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	resCh := make(chan ntpTimestamp)

	go func() {
		res := make([]byte, 64)
		_, err := conn.Read(res)
		destinationTS := time.Now()
		if err != nil {
			close(resCh)
			return
		}

		var buf bytes.Buffer
		buf.Write(res[0:])
		buf.Next(24)

		// originateTS
		buf.Next(8)

		// receiveTS
		receiveSeconds := binary.BigEndian.Uint32(buf.Next(4))
		receiveFrac := binary.BigEndian.Uint32(buf.Next(4))
		receiveTS := timestampToTime(int(receiveSeconds), int(receiveFrac))

		// transmitTS
		transmitSeconds := binary.BigEndian.Uint32(buf.Next(4))
		transmitFrac := binary.BigEndian.Uint32(buf.Next(4))
		transmitTS := timestampToTime(int(transmitSeconds), int(transmitFrac))

		ts := ntpTimestamp{
			originateTimestamp:   nil,
			receiveTimestamp:     &receiveTS,
			transmitTimestamp:    &transmitTS,
			destinationTimestamp: &destinationTS,
		}
		resCh <- ts
	}()

	req := make([]byte, 48)
	req[0] = 0x1b
	originateTS := time.Now()
	_, err = conn.Write(req)
	if err != nil {
		return nil, err
	}

	ts, ok := <-resCh
	if !ok {
		return nil, errors.New("error in receiving ntp packets.")
	}
	ts.originateTimestamp = &originateTS

	offset := ntpOffset(ts)
	ntpNow := time.Now().Add(offset)

	return &ntpNow, nil
}

func ntpOffset(ts ntpTimestamp) time.Duration {
	// offset = ((t2-t1)+(t3-t4))/2
	offset := (ts.receiveTimestamp.Sub(*ts.originateTimestamp) + ts.transmitTimestamp.Sub(*ts.destinationTimestamp)) / 2
	return offset
}

func timestampToTime(intg, frac int) time.Time {
	layout := "2006-01-02 03:04:05"
	epoch, _ := time.Parse(layout, "1900-01-01 00:00:00")
	d, _ := time.ParseDuration(fmt.Sprintf("%d.%ds", intg, frac))

	return epoch.Add(d)
}
