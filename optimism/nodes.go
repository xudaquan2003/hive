package optimism

import "github.com/ethereum/hive/hivesim"

var (
	L1_HTTP_PORT = uint16(8545)
	L1_WS_PORT   = uint16(8546)

	L2_HTTP_PORT = uint16(8545)
	L2_WS_PORT   = uint16(8546)
	L2_AUTH_PORT = uint16(8551)

	OP_HTTP_PORT = uint16(8547)
)

type Eth1Node struct {
	*hivesim.Client
	HTTPPort uint16
	WSPort   uint16
}

type DeployerNode struct {
	*hivesim.Client
	// HTTPPort uint16
	// WSPort   uint16
}

type L2Node struct {
	*hivesim.Client
	HTTPPort    uint16
	WSPort      uint16
	AuthrpcPort uint16
}

type OpNode struct {
	*hivesim.Client
	HTTPPort uint16
}

type L2OSNode struct {
	*hivesim.Client
}

type BSSNode struct {
	*hivesim.Client
}
