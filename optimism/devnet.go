package optimism

import (
	"context"
	"fmt"
	"os"

	"math/big"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/hive/hivesim"
)

type Devnet struct {
	T *hivesim.T

	L1       *Eth1Node
	Deployer *DeployerNode
	L2       *L2Node
	Rollup   *OpNode
	Verifier *OpNode
	Proposer *L2OSNode
	Batcher  *BSSNode

	GenesisL2 string
	Jwtsecret string

	GenesisTimestamp string
	L2Genesis        string
	RollupJson       string
	WalletsJson      string

	L2ToL1MessagePasserJSON          string
	L2CrossDomainMessengerJSON       string
	OptimismMintableTokenFactoryJSON string
	L2StandardBridgeJSON             string
	L1BlockJSON                      string

	L2OutputOracle string
	OptimismPortal string

	Nodes map[string]*hivesim.ClientDefinition
	Ctx   context.Context
}

func (d *Devnet) Start(l2ChainID *big.Int) {
	clientTypes, err := d.T.Sim.ClientTypes()
	if err != nil {
		d.T.Fatal(err)
	}
	var eth1, deployer, l2, op, l2os, bss *hivesim.ClientDefinition
	for _, client := range clientTypes {
		if client.HasRole("op-l1") {
			eth1 = client
		}
		if client.HasRole("op-deployer") {
			deployer = client
		}
		if client.HasRole("op-l2") {
			l2 = client
		}
		if client.HasRole("op-node") {
			op = client
		}
		if client.HasRole("op-proposer") {
			l2os = client
		}
		if client.HasRole("op-batcher") {
			bss = client
		}
	}

	// if eth1 == nil || deployer == nil || l2 == nil || op == nil || l2os == nil || bss == nil {
	// 	d.T.Fatal("op-l1, op-l2, op-node, op-proposer, op-batcher required")
	// }

	var eth1ConfigOpt, eth1Bundle hivesim.Params
	execNodeOpts := hivesim.Params{
		"HIVE_CATALYST_ENABLED": "1",
		"HIVE_LOGLEVEL":         os.Getenv("HIVE_LOGLEVEL"),

		"HIVE_L1_HTTP_PORT": fmt.Sprintf("%d", L1_HTTP_PORT),
		"HIVE_L1_WS_PORT":   fmt.Sprintf("%d", L1_WS_PORT),
	}
	executionOpts := hivesim.Bundle(eth1ConfigOpt, eth1Bundle, execNodeOpts)

	d.T.Logf("eth1.Name: %s", eth1.Name)
	d.T.Log(eth1)

	opts := []hivesim.StartOption{executionOpts}
	d.L1 = &Eth1Node{d.T.StartClient(eth1.Name, opts...), L1_HTTP_PORT, L1_WS_PORT}
	d.Wait()

	l1_rpc_url := fmt.Sprintf("http://%s:%d", d.L1.IP, d.L1.HTTPPort)
	var deployerConfigOpt, deployerBundle hivesim.Params
	deployerNodeOpts := hivesim.Params{
		"HIVE_L1_RPC_URL":  l1_rpc_url,
		"HIVE_L2_CHAIN_ID": l2ChainID.String(),
	}
	deployerOpts := hivesim.Bundle(deployerConfigOpt, deployerBundle, deployerNodeOpts)
	opts = []hivesim.StartOption{deployerOpts}

	// d.Deployer = &DeployerNode{d.T.StartClient(deployer.Name, opts...), 8545, 8546}
	d.Deployer = &DeployerNode{d.T.StartClient(deployer.Name, opts...)}

	d.Nodes["op-l1"] = eth1
	d.Nodes["op-deployer"] = deployer
	d.Nodes["op-l2"] = l2
	d.Nodes["op-node"] = op
	d.Nodes["op-proposer"] = l2os
	d.Nodes["op-batcher"] = bss
}

func (d *Devnet) Wait() error {
	// TODO: wait until rpc connects
	client := ethclient.NewClient(d.L1.Client.RPC())
	chainID, err := client.ChainID(d.Ctx)
	d.T.Logf("d.L1.Client, chainID: %s", chainID.String())
	return err
}

func (d *Devnet) Cat(path string) (string, error) {
	execInfo, err := d.Deployer.Client.Exec("cat.sh", path)
	if err != nil {
		return "", err
	}
	return execInfo.Stdout, nil
}

func (d *Devnet) InitL2(l2ChainID *big.Int) error {
	genesisL2, err := d.Cat(fmt.Sprintf("/network-data/genesis-%s.json", l2ChainID.String()))
	if err != nil {
		return err
	}
	d.GenesisL2 = genesisL2
	// d.T.Logf("genesisL2:\n %s", genesisL2)

	jwtsecret, err := d.Cat("/jwt/jwtsecret")
	if err != nil {
		return err
	}
	d.Jwtsecret = jwtsecret
	d.T.Logf("jwtsecret:\n %s", jwtsecret)

	return nil
}

func (d *Devnet) StartL2(accountOpts hivesim.StartOption) error {
	l2 := d.Nodes["op-l2"]

	executionOpts := hivesim.Params{
		"HIVE_CHECK_LIVE_PORT": fmt.Sprintf("%d", L2_HTTP_PORT),
		"HIVE_LOGLEVEL":        os.Getenv("HIVE_LOGLEVEL"),
		"HIVE_NODETYPE":        "full",
		"HIVE_NETWORK_ID":      networkID.String(),
		"HIVE_CHAIN_ID":        chainID.String(),

		"HIVE_L2_HTTP_PORT": fmt.Sprintf("%d", L2_HTTP_PORT),
		"HIVE_L2_WS_PORT":   fmt.Sprintf("%d", L2_WS_PORT),
		"HIVE_L2_AUTH_PORT": fmt.Sprintf("%d", L2_AUTH_PORT),
	}

	genesisL2Opt := hivesim.WithDynamicFile("/genesis.json", bytesSource([]byte(d.GenesisL2)))
	jwtsecretOpt := hivesim.WithDynamicFile("/jwtsecret", bytesSource([]byte(d.Jwtsecret)))
	opts := []hivesim.StartOption{executionOpts, genesisL2Opt, accountOpts, jwtsecretOpt}

	d.L2 = &L2Node{d.T.StartClient(l2.Name, opts...), L2_HTTP_PORT, L2_WS_PORT, L2_AUTH_PORT}
	return nil
}

func (d *Devnet) InitOp(l2ChainID *big.Int) error {
	rollup, err := d.Cat(fmt.Sprintf("/network-data/rollup-%s.json", l2ChainID.String()))
	if err != nil {
		return err
	}
	d.RollupJson = rollup
	// d.T.Logf("RollupJson:\n %s", rollup)

	wallets, err := d.Cat("/network-data/wallets.json")
	if err != nil {
		return err
	}
	d.WalletsJson = wallets
	// d.T.Logf("WalletsJson:\n %s", wallets)
	return nil
}

func (d *Devnet) StartOp() error {
	op := d.Nodes["op-node"]

	executionOpts := hivesim.Params{
		"HIVE_CHECK_LIVE_PORT": fmt.Sprintf("%d", OP_HTTP_PORT),

		"HIVE_OP_HTTP_PORT": fmt.Sprintf("%d", OP_HTTP_PORT),

		"HIVE_L1_HTTP_URL": fmt.Sprintf("http://%s:%d", d.L1.IP, d.L1.HTTPPort),
		"HIVE_L2_AUTH_URL": fmt.Sprintf("http://%s:%d", d.L2.IP, d.L2.AuthrpcPort),
		"HIVE_L2_HTTP_URL": fmt.Sprintf("http://%s:%d", d.L2.IP, d.L2.HTTPPort),
	}

	rollupOpt := hivesim.WithDynamicFile("/rollup.json", bytesSource([]byte(d.RollupJson)))
	jwtsecretOpt := hivesim.WithDynamicFile("/jwtsecret", bytesSource([]byte(d.Jwtsecret)))
	walletsOpt := hivesim.WithDynamicFile("/wallets.json", bytesSource([]byte(d.WalletsJson)))

	opts := []hivesim.StartOption{executionOpts, rollupOpt, jwtsecretOpt, walletsOpt}
	d.Rollup = &OpNode{d.T.StartClient(op.Name, opts...), 8547}
	return nil
}

func (d *Devnet) StartL2OS() error {
	l2os := d.Nodes["op-proposer"]

	executionOpts := hivesim.Params{
		"HIVE_CHECK_LIVE_PORT":  "0",
		"HIVE_CATALYST_ENABLED": "1",
		"HIVE_LOGLEVEL":         os.Getenv("HIVE_LOGLEVEL"),
		"HIVE_NODETYPE":         "full",

		"HIVE_L1_ETH_RPC_FLAG": fmt.Sprintf("--l1-eth-rpc=http://%s:%d", d.L1.IP, d.L1.HTTPPort),
		"HIVE_L2_ETH_RPC_FLAG": fmt.Sprintf("--l2-eth-rpc=http://%s:%d", d.L2.IP, d.L2.HTTPPort),
		"HIVE_ROLLUP_RPC_FLAG": fmt.Sprintf("--rollup-rpc=http://%s:%d", d.Rollup.IP, d.Rollup.HTTPPort),
	}

	l2OutputOracleOpt := hivesim.WithDynamicFile("/L2OutputOracleProxy.json", bytesSource([]byte(d.L2OutputOracle)))
	opts := []hivesim.StartOption{executionOpts, l2OutputOracleOpt}
	d.Proposer = &L2OSNode{d.T.StartClient(l2os.Name, opts...)}
	return nil
}

func (d *Devnet) StartBSS() error {
	bss := d.Nodes["op-batcher"]

	executionOpts := hivesim.Params{
		"HIVE_CHECK_LIVE_PORT": "0",

		"HIVE_L1_HTTP_URL":     fmt.Sprintf("http://%s:%d", d.L1.IP, d.L1.HTTPPort),
		"HIVE_L2_HTTP_URL":     fmt.Sprintf("http://%s:%d", d.L2.IP, d.L2.HTTPPort),
		"HIVE_ROLLUP_HTTP_URL": fmt.Sprintf("http://%s:%d", d.Rollup.IP, d.Rollup.HTTPPort),
	}

	walletsOpt := hivesim.WithDynamicFile("/wallets.json", bytesSource([]byte(d.WalletsJson)))
	opts := []hivesim.StartOption{executionOpts, walletsOpt}
	d.Batcher = &BSSNode{d.T.StartClient(bss.Name, opts...)}
	return nil
}
