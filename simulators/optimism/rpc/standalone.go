package main

import (
	"fmt"
	"net"
	"strings"

	"github.com/ethereum/hive/hivesim"
	"github.com/ethereum/hive/optimism"
)

func runTestStandAlone(t *hivesim.T) {
	t.Log("running all tests standalone")
	client := &hivesim.Client{IP: net.ParseIP("67.213.115.211")}
	l2 := &optimism.L2Node{Client: client, HTTPPort: 9545, WSPort: 9546, AuthrpcPort: 9551}

	t.Logf("L2.Client.HTTP_URL:\n %s\n", fmt.Sprintf("http://%v:%d", l2.Client.IP, l2.HTTPPort))

	vault := newVault()

	s := newSemaphore(16)
	for _, test := range tests {
		test := test
		s.get()
		go func() {
			defer s.put()
			t.Run(hivesim.TestSpec{
				Name:        fmt.Sprintf("%s (%s)", test.Name, "mega-test"),
				Description: test.About,
				Run: func(t *hivesim.T) {
					switch test.Name[:strings.IndexByte(test.Name, '/')] {
					case "http":
						runHTTP(t, l2, vault, test.Run)
					case "ws":
						runWS(t, l2, vault, test.Run)
					default:
						panic("bad test prefix in name " + test.Name)
					}
				},
			})
		}()
	}
	s.drain()
}
