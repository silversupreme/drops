package server

import (
	"bufio"
	"fmt"
	"net"
	"testing"

	"github.com/benbjohnson/clock"
)

type interaction struct {
	send   string
	expect string
}

// Simple things that don't need complex client interactions.
var simpleCmdTestCases = []struct {
	name         string
	interactions []interaction
}{
	{
		name: "BlankListCmd",
		interactions: []interaction{
			{"LIST", "LIST"},
		},
	},
	{
		name: "ListCmdEnforces0Args",
		interactions: []interaction{
			{"LIST SOMETHING", "ERR"},
		},
	},
	{
		name: "RegisterListCmd",
		interactions: []interaction{
			{"REGISTER water source", "ACK"},
			{"LIST", "LIST water:source"},
		},
	},
	{
		name: "RegisterErr",
		interactions: []interaction{
			{"REGISTER water", "ERR"},
		},
	},
	{
		name: "MetricsRequireRegistration",
		interactions: []interaction{
			{"METRIC test 10.000", "ERR"},
		},
	},
	{
		name: "MetricRegistration",
		interactions: []interaction{
			{"REGISTER water source", "ACK"},
			{"METRIC level 91.120", "ACK"},
			{"METRICS water", "METRICS water level"},
		},
	},
	{
		name: "MetricsRequireFloat",
		interactions: []interaction{
			{"REGISTER water source", "ACK"},
			{"METRIC level something", "ERR"},
		},
	},
	{
		name: "MetricsList",
		interactions: []interaction{
			{"REGISTER water source", "ACK"},
			{"METRIC level 1", "ACK"},
			{"METRIC level 2", "ACK"},
			{"METRIC level 3", "ACK"},
			{"METRICS water level", "METRICS water level 0:1.00 0:2.00 0:3.00"},
		},
	},
	{
		name: "DoubleRegistrationFails",
		interactions: []interaction{
			{"REGISTER water source", "ACK"},
			{"REGISTER water barrel", "ERR"},
		},
	},
	{
		name: "UnknownMetricFails",
		interactions: []interaction{
			{"REGISTER water source", "ACK"},
			{"METRICS water level", "ERR"},
		},
	},
	{
		name: "MaxMetricCount",
		interactions: []interaction{
			{"REGISTER water source", "ACK"},
			{"METRIC level 1", "ACK"},
			{"METRIC level 2", "ACK"},
			{"METRIC level 3", "ACK"},
			{"METRIC level 4", "ACK"},
			{"METRIC level 5", "ACK"},
			{"METRICS water level", "METRICS water level 0:2.00 0:3.00 0:4.00 0:5.00"},
		},
	},
	{
		name: "UnknownCommand",
		interactions: []interaction{
			{"DOODLE", "ERR UNRECOGNIZED CMD"},
		},
	},
	{
		name: "Blank",
		interactions: []interaction{
			{"", "ERR UNRECOGNIZED CMD"},
		},
	},
}

func TestSimpleCmds(t *testing.T) {
	for _, test := range simpleCmdTestCases {
		t.Run(test.name, func(t *testing.T) {
			// Listen on a random port for each test.
			listener, err := net.Listen("tcp", ":0")
			if err != nil {
				t.Fatal(err)
			}

			addr := listener.Addr()
			mock := clock.NewMock()
			server := New(listener, 4, mock)
			go server.Serve()

			conn, err := net.Dial("tcp", addr.String())
			if err != nil {
				t.Fatal(err)
			}

			for _, i := range test.interactions {
				toSend := []byte(fmt.Sprintf("%s\n", i.send))
				if _, err := conn.Write(toSend); err != nil {
					t.Fatal(err)
				}

				connReader := bufio.NewReader(conn)
				output, err := connReader.ReadString('\n')
				if err != nil {
					t.Fatal(err)
				}

				toExpect := fmt.Sprintf("%s\n", i.expect)
				if output != toExpect {
					t.Fatalf("`%s` expected `%s`, got %s", i.send, i.expect, output)
				}
			}

			conn.Close()
		})
	}
}