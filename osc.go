package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"strconv"
	"time"
)

// OscValue is a single OSC argument with an explicit type tag.
// Supported tags: 'i' int32, 'h' int64, 'f' float32, 'd' float64,
// 's' string, 'T' true, 'F' false, 'N' nil.
type OscValue struct {
	Type byte
	I    int32
	I64  int64
	F    float32
	D    float64
	S    string
}

// oscString writes an OSC string: the raw UTF-8 bytes followed by a null
// terminator, padded with additional nulls so the total length is a multiple
// of 4 bytes. This mirrors python-osc's osc_types.write_string.
func oscString(b *bytes.Buffer, s string) {
	b.WriteString(s)
	n := len(s)
	// pad is in 1..4, guaranteeing at least one null terminator byte.
	pad := 4 - (n % 4)
	b.Write(make([]byte, pad))
}

// buildMessage encodes an OSC message into its wire format:
// address, type-tag string, then each argument, all padded to 4 bytes.
func buildMessage(address string, args []OscValue) []byte {
	var b bytes.Buffer
	oscString(&b, address)

	tag := ","
	for _, a := range args {
		tag += string(a.Type)
	}
	oscString(&b, tag)

	for _, a := range args {
		switch a.Type {
		case 'i':
			_ = binary.Write(&b, binary.BigEndian, a.I)
		case 'h':
			_ = binary.Write(&b, binary.BigEndian, a.I64)
		case 'f':
			_ = binary.Write(&b, binary.BigEndian, a.F)
		case 'd':
			_ = binary.Write(&b, binary.BigEndian, a.D)
		case 's':
			oscString(&b, a.S)
		case 'T', 'F', 'N':
			// These types carry no payload bytes.
		}
	}
	return b.Bytes()
}

// readOscString reads a null-terminated, 4-byte-padded OSC string starting at idx.
func readOscString(d []byte, idx int) (string, int, error) {
	start := idx
	n := 0
	for idx < len(d) && d[idx] != 0 {
		idx++
		n++
	}
	if idx >= len(d) {
		return "", 0, fmt.Errorf("osc: unterminated string")
	}
	s := string(d[start : start+n])
	idx += 4 - (n % 4) // skip terminator null plus padding
	if idx > len(d) {
		return "", 0, fmt.Errorf("osc: short datagram")
	}
	return s, idx, nil
}

// parseMessage decodes an OSC message, returning its address and arguments.
func parseMessage(d []byte) (string, []OscValue, error) {
	addr, idx, err := readOscString(d, 0)
	if err != nil {
		return "", nil, err
	}
	if idx >= len(d) {
		return addr, nil, nil
	}

	tag, idx, err := readOscString(d, idx)
	if err != nil {
		return "", nil, err
	}
	if len(tag) > 0 && tag[0] == ',' {
		tag = tag[1:]
	}

	args := make([]OscValue, 0, len(tag))
	for _, t := range []byte(tag) {
		v := OscValue{Type: t}
		switch t {
		case 'i':
			if idx+4 > len(d) {
				return "", nil, fmt.Errorf("osc: short int32")
			}
			v.I = int32(binary.BigEndian.Uint32(d[idx : idx+4]))
			idx += 4
		case 'h':
			if idx+8 > len(d) {
				return "", nil, fmt.Errorf("osc: short int64")
			}
			v.I64 = int64(binary.BigEndian.Uint64(d[idx : idx+8]))
			idx += 8
		case 'f':
			if idx+4 > len(d) {
				return "", nil, fmt.Errorf("osc: short float32")
			}
			v.F = math.Float32frombits(binary.BigEndian.Uint32(d[idx : idx+4]))
			idx += 4
		case 'd':
			if idx+8 > len(d) {
				return "", nil, fmt.Errorf("osc: short float64")
			}
			v.D = math.Float64frombits(binary.BigEndian.Uint64(d[idx : idx+8]))
			idx += 8
		case 's':
			s, ni, err := readOscString(d, idx)
			if err != nil {
				return "", nil, err
			}
			v.S = s
			idx = ni
		case 'T', 'F', 'N':
			// No payload.
		case 'b': // blob: int32 length + data, padded
			if idx+4 > len(d) {
				return "", nil, fmt.Errorf("osc: short blob")
			}
			size := int(int32(binary.BigEndian.Uint32(d[idx : idx+4])))
			idx += 4 + size
			idx += (4 - size%4) % 4
			if idx > len(d) {
				return "", nil, fmt.Errorf("osc: short blob data")
			}
		case 'r', 'm': // rgba / midi: 4 bytes
			if idx+4 > len(d) {
				return "", nil, fmt.Errorf("osc: short rgba/midi")
			}
			idx += 4
		case 't': // timetag: 8 bytes
			if idx+8 > len(d) {
				return "", nil, fmt.Errorf("osc: short timetag")
			}
			idx += 8
		default:
			return "", nil, fmt.Errorf("osc: unknown type tag %q", t)
		}
		args = append(args, v)
	}
	return addr, args, nil
}

// String returns a human-readable representation of the value.
func (v OscValue) String() string {
	switch v.Type {
	case 'i':
		return strconv.FormatInt(int64(v.I), 10)
	case 'h':
		return strconv.FormatInt(v.I64, 10)
	case 'f':
		return strconv.FormatFloat(float64(v.F), 'g', -1, 32)
	case 'd':
		return strconv.FormatFloat(v.D, 'g', -1, 64)
	case 's':
		return v.S
	case 'T':
		return "true"
	case 'F':
		return "false"
	case 'N':
		return "null"
	default:
		return "?"
	}
}

// toJSON converts the value to a Go value suitable for encoding/json.
func (v OscValue) toJSON() interface{} {
	switch v.Type {
	case 'i':
		return v.I
	case 'h':
		return v.I64
	case 'f':
		return float64(v.F)
	case 'd':
		return v.D
	case 's':
		return v.S
	case 'T':
		return true
	case 'F':
		return false
	case 'N':
		return nil
	default:
		return nil
	}
}

// Client is a minimal OSC client for AbletonOSC.
type Client struct {
	Host         string
	Port         int
	ResponsePort int
	Timeout      time.Duration
}

// Send transmits a single OSC message without waiting for a reply.
func (c *Client) Send(address string, args []OscValue) error {
	conn, err := net.Dial("udp", net.JoinHostPort(c.Host, strconv.Itoa(c.Port)))
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Write(buildMessage(address, args))
	return err
}

// Query sends a message and waits for a matching reply.
// AbletonOSC replies on the response port, addressed to the source host.
func (c *Client) Query(address string, args []OscValue) (string, []OscValue, error) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: c.ResponsePort})
	if err != nil {
		return "", nil, fmt.Errorf("bind reply port %d: %w", c.ResponsePort, err)
	}
	defer conn.Close()

	remote := &net.UDPAddr{IP: net.ParseIP(c.Host), Port: c.Port}
	if _, err := conn.WriteToUDP(buildMessage(address, args), remote); err != nil {
		return "", nil, err
	}

	_ = conn.SetReadDeadline(time.Now().Add(c.Timeout))
	buf := make([]byte, 65536)

	var fallbackAddr string
	var fallbackVals []OscValue
	haveFallback := false

	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			if haveFallback {
				return fallbackAddr, fallbackVals, nil
			}
			return "", nil, fmt.Errorf(
				"no reply within %s (is Live running with AbletonOSC selected as a Control Surface?)",
				c.Timeout)
		}
		a, vals, err := parseMessage(buf[:n])
		if err != nil {
			continue
		}
		if a == "/live/error" || a == address {
			return a, vals, nil
		}
		if !haveFallback {
			fallbackAddr, fallbackVals, haveFallback = a, vals, true
		}
	}
}

// Stream sends a start_listen message, forwards every incoming OSC message to
// onMessage for the given duration, then sends a stop_listen message. It holds
// the response port for the whole duration.
func (c *Client) Stream(startAddr string, startArgs []OscValue, stopAddr string, stopArgs []OscValue, duration time.Duration, onMessage func(addr string, vals []OscValue) error) error {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: c.ResponsePort})
	if err != nil {
		return fmt.Errorf("bind reply port %d: %w", c.ResponsePort, err)
	}
	defer conn.Close()

	remote := &net.UDPAddr{IP: net.ParseIP(c.Host), Port: c.Port}
	if _, err := conn.WriteToUDP(buildMessage(startAddr, startArgs), remote); err != nil {
		return err
	}

	_ = conn.SetReadDeadline(time.Now().Add(duration))
	buf := make([]byte, 65536)
	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			// Deadline (or any read error) ends the stream normally.
			break
		}
		a, vals, err := parseMessage(buf[:n])
		if err != nil {
			continue
		}
		if err := onMessage(a, vals); err != nil {
			return err
		}
	}

	if stopAddr != "" {
		_, _ = conn.WriteToUDP(buildMessage(stopAddr, stopArgs), remote)
	}
	return nil
}
