// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// Package les implements the Light Ethereum Subprotocol.
package les

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Venachain/Venachain/internal/debug"

	"github.com/Venachain/Venachain/p2p/discover"

	"github.com/Venachain/Venachain/accounts"
	"github.com/Venachain/Venachain/common"
	"github.com/Venachain/Venachain/common/hexutil"
	"github.com/Venachain/Venachain/consensus"
	"github.com/Venachain/Venachain/core"
	"github.com/Venachain/Venachain/core/bloombits"
	"github.com/Venachain/Venachain/core/types"
	"github.com/Venachain/Venachain/event"
	"github.com/Venachain/Venachain/internal/venaapi"
	"github.com/Venachain/Venachain/light"
	"github.com/Venachain/Venachain/log"
	"github.com/Venachain/Venachain/node"
	"github.com/Venachain/Venachain/p2p"
	"github.com/Venachain/Venachain/p2p/discv5"
	"github.com/Venachain/Venachain/params"
	rpc "github.com/Venachain/Venachain/rpc"
	"github.com/Venachain/Venachain/vena"
	"github.com/Venachain/Venachain/vena/downloader"
	"github.com/Venachain/Venachain/vena/filters"
	"github.com/Venachain/Venachain/vena/gasprice"
)

type LightEthereum struct {
	lesCommons

	odr         *LesOdr
	relay       *LesTxRelay
	chainConfig *params.ChainConfig
	// Channel for shutting down the service
	shutdownChan chan bool

	// Handlers
	peers      *peerSet
	txPool     *light.TxPool
	blockchain *light.LightChain
	serverPool *serverPool
	reqDist    *requestDistributor
	retriever  *retrieveManager

	bloomRequests chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer  *core.ChainIndexer

	ApiBackend *LesApiBackend

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	networkId     uint64
	netRPCService *venaapi.PublicNetAPI

	wg sync.WaitGroup
}

type Node struct {
	Addr       string `json:"addr"`
	ExpireTime string `json:"expiretime"`
}

func New(ctx *node.ServiceContext, config *vena.Config) (*LightEthereum, error) {
	chainDb, err := vena.CreateDB(ctx, config, "lightchaindata")
	if err != nil {
		return nil, err
	}

	extDb, err := vena.CreateExtDB(ctx, config, "extdb")

	chainConfig, _, genesisErr := core.SetupGenesisBlock(chainDb, config.Genesis)

	if genesisErr != nil {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	peers := newPeerSet()
	quitSync := make(chan struct{})

	leth := &LightEthereum{
		lesCommons: lesCommons{
			chainDb: chainDb,
			config:  config,
			iConfig: light.DefaultClientIndexerConfig,
		},
		chainConfig:    chainConfig,
		eventMux:       ctx.EventMux,
		peers:          peers,
		reqDist:        newRequestDistributor(peers, quitSync),
		accountManager: ctx.AccountManager,
		engine:         vena.CreateConsensusEngine(ctx, chainConfig, chainDb),
		shutdownChan:   make(chan bool),
		networkId:      config.NetworkId,
		bloomRequests:  make(chan chan *bloombits.Retrieval),
		bloomIndexer:   vena.NewBloomIndexer(chainDb, params.BloomBitsBlocksClient, params.HelperTrieConfirmations),
	}

	leth.relay = NewLesTxRelay(peers, leth.reqDist)
	leth.serverPool = newServerPool(chainDb, quitSync, &leth.wg)
	leth.retriever = newRetrieveManager(peers, leth.reqDist, leth.serverPool)

	leth.odr = NewLesOdr(chainDb, light.DefaultClientIndexerConfig, leth.retriever)
	leth.chtIndexer = light.NewChtIndexer(chainDb, leth.odr, params.CHTFrequencyClient, params.HelperTrieConfirmations)
	leth.bloomTrieIndexer = light.NewBloomTrieIndexer(chainDb, leth.odr, params.BloomBitsBlocksClient, params.BloomTrieFrequency)
	leth.odr.SetIndexers(leth.chtIndexer, leth.bloomTrieIndexer, leth.bloomIndexer)

	// Note: NewLightChain adds the trusted checkpoint so it needs an ODR with
	// indexers already set but not started yet
	if leth.blockchain, err = light.NewLightChain(leth.odr, leth.chainConfig, leth.engine); err != nil {
		return nil, err
	}
	// Note: AddChildIndexer starts the update process for the child
	leth.bloomIndexer.AddChildIndexer(leth.bloomTrieIndexer)
	leth.chtIndexer.Start(leth.blockchain)
	leth.bloomIndexer.Start(leth.blockchain)

	leth.txPool = light.NewTxPool(leth.chainConfig, leth.blockchain, leth.relay)
	if leth.protocolManager, err = NewProtocolManager(leth.chainConfig, light.DefaultClientIndexerConfig, true, config.NetworkId, leth.eventMux, leth.engine, leth.peers, leth.blockchain, nil, chainDb, extDb, leth.odr, leth.relay, leth.serverPool, quitSync, &leth.wg); err != nil {
		return nil, err
	}
	leth.ApiBackend = &LesApiBackend{leth, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.MinerGasPrice
	}
	leth.ApiBackend.gpo = gasprice.NewOracle(leth.ApiBackend, gpoParams)

	if chainConfig.LicenseCheck {
		log.Info("license", "enable", chainConfig.LicenseCheck)
		log.Info("Start license check right now.")

		checked, expireTime := licenseCheck(leth.config.Etherbase)
		if checked {
			go func() {
				remainingSecond := expireTime - time.Now().Unix()
				timeout := time.After(time.Second * time.Duration(remainingSecond))

				select {
				case <-timeout:
					//rawdb.ReadHeadBlockHash(eth.chainDb) //todo read timestamp in head block and compare.

					log.Info("License expired: stopping the node right now.")

					go leth.Stop()

					debug.Exit() // ensure trace and CPU profile data is flushed.
					debug.LoudPanic("boom")
				}
			}()
		}
	}

	return leth, nil
}

func lesTopic(genesisHash common.Hash, protocolVersion uint) discv5.Topic {
	var name string
	switch protocolVersion {
	case lpv1:
		name = "LES"
	case lpv2:
		name = "LES2"
	default:
		panic(nil)
	}
	return discv5.Topic(name + "@" + common.Bytes2Hex(genesisHash.Bytes()[0:8]))
}

type LightDummyAPI struct{}

// Etherbase is the address that mining rewards will be send to
func (s *LightDummyAPI) Etherbase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("not supported")
}

// Coinbase is the address that mining rewards will be send to (alias for Etherbase)
func (s *LightDummyAPI) Coinbase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("not supported")
}

// Hashrate returns the POW hashrate
func (s *LightDummyAPI) Hashrate() hexutil.Uint {
	return 0
}

// Mining returns an indication if this node is currently mining.
func (s *LightDummyAPI) Mining() bool {
	return false
}

// APIs returns the collection of RPC services the ethereum package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *LightEthereum) APIs() []rpc.API {
	return append(venaapi.GetAPIs(s.ApiBackend), []rpc.API{
		{
			Namespace: "eth",
			Version:   "1.0",
			Service:   &LightDummyAPI{},
			Public:    true,
		}, {
			Namespace: "eth",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "eth",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.ApiBackend, true),
			Public:    true,
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)
}

func (s *LightEthereum) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *LightEthereum) BlockChain() *light.LightChain      { return s.blockchain }
func (s *LightEthereum) TxPool() *light.TxPool              { return s.txPool }
func (s *LightEthereum) Engine() consensus.Engine           { return s.engine }
func (s *LightEthereum) LesVersion() int                    { return int(ClientProtocolVersions[0]) }
func (s *LightEthereum) Downloader() *downloader.Downloader { return s.protocolManager.downloader }
func (s *LightEthereum) EventMux() *event.TypeMux           { return s.eventMux }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *LightEthereum) Protocols() []p2p.Protocol {
	return s.makeProtocols(ClientProtocolVersions)
}

// Start implements node.Service, starting all internal goroutines needed by the
// Ethereum protocol implementation.
func (s *LightEthereum) Start(srvr *p2p.Server) error {
	log.Warn("Light client mode is an experimental feature")
	s.startBloomHandlers(params.BloomBitsBlocksClient)
	s.netRPCService = venaapi.NewPublicNetAPI(srvr, s.networkId)
	// clients are searching for the first advertised protocol in the list
	protocolVersion := AdvertiseProtocolVersions[0]
	s.serverPool.start(srvr, lesTopic(s.blockchain.Genesis().Hash(), protocolVersion))
	s.protocolManager.Start(s.config.LightPeers)

	if _, ok := s.engine.(consensus.Istanbul); ok {
		for _, n := range p2p.GetBootNodes() {
			srvr.AddPeer(discover.NewNode(n.ID, n.IP, n.UDP, n.TCP))
		}
	}
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Ethereum protocol.
func (s *LightEthereum) Stop() error {
	s.odr.Stop()
	s.bloomIndexer.Close()
	s.chtIndexer.Close()
	s.blockchain.Stop()
	s.protocolManager.Stop()
	s.txPool.Stop()
	s.engine.Close()

	s.eventMux.Stop()

	time.Sleep(time.Millisecond * 200)
	s.chainDb.Close()
	s.extDb.Close()
	close(s.shutdownChan)

	return nil
}

func licenseCheck(addr common.Address) (bool, int64) {
	log.Info("Node address: ", "addr", addr.String())

	// load signature file.
	dir, _ := os.Getwd()
	log.Info(dir + "/../data/signature" + addr.String())
	fi, err := os.Open(dir + "/../data/signature" + addr.String())
	if err != nil {
		log.Info("Error: %s\n", err)
		return false, 0
	}
	defer fi.Close()

	var licenseInfo []string

	br := bufio.NewReader(fi)
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		licenseInfo = append(licenseInfo, string(a))
	}

	licenseInfoSplit := strings.Split(licenseInfo[0], " ")

	if len(licenseInfoSplit) != 4 {
		log.Info("License info doesn't enough. Parameter required 4: [node address] [expire time] [R] [S]")
		return false, 0
	}

	// check addr
	if addr.String() != licenseInfoSplit[0] {
		log.Info("Node address doesn't match the address license provided! ", "addr", addr.String(), licenseInfoSplit[0])
		return false, 0
	}

	log.Info("Node address matched.", "addr", licenseInfoSplit[0])

	// check expire time
	expireTime, err := strconv.ParseInt(licenseInfoSplit[1], 10, 64)
	if time.Now().Unix() >= expireTime {
		log.Info("License expired!")
		log.Info("The expire time is set to ", expireTime)
		return false, 0
	}

	// following: check signature
	nodeInfo := Node{
		Addr:       licenseInfoSplit[0],
		ExpireTime: licenseInfoSplit[1],
	}

	jsonR, err := json.Marshal(nodeInfo)
	if err != nil {
		log.Info("Node info marshal err: ", err)
		return false, 0
	}

	msgHash := sha256.New()
	_, err = msgHash.Write(jsonR)
	if err != nil {
		log.Info("message hash error: ", err)
		return false, 0
	}
	msgHashSum := msgHash.Sum(nil)

	//read public key info.
	publickeyInfo := `-----BEGIN ECDSA public key-----
MIGbMBAGByqGSM49AgEGBSuBBAAjA4GGAAQBlG2xio9lfJaNVXmgGJamH2iBkBxA
CUzh0qhn6F4AjPdupYVl0BFAFp8zcgf+T/CD63y82LTztJbhaMMGv67BEnEA0A2r
vfEnetVuu9nvSJYdtXLqoPwKKmeKLzHuPciWYjVN659/ghsvX5t7D9muj0a5NDLp
QN275TE7TLxctFVF0eY=
-----END ECDSA public key-----
`

	//dstFile, err := os.Create(dir + "/../data/publickey.pem")
	//if err != nil {
	//	logrus.Fatal(err)
	//}
	//
	//defer dstFile.Close()
	//dstFile.WriteString(publickeyInfo)

	tmpFile, err := ioutil.TempFile(dir+"/../data/", "tmp")
	defer os.Remove(tmpFile.Name())
	if err != nil {
		log.Info("Error when creating temp file.", err)
		return false, 0
	}
	tmpFile.WriteString(publickeyInfo)

	publicKeyfile, err := os.Open(tmpFile.Name())
	if err != nil {
		log.Info("Open public key file error: ", err)
		return false, 0
	}

	log.Info("Open public key file", tmpFile.Name())

	publicInfo, _ := publicKeyfile.Stat()
	publicBuf := make([]byte, publicInfo.Size())
	publicKeyfile.Read(publicBuf)

	publicBlock, _ := pem.Decode(publicBuf)

	publicKey, err := x509.ParsePKIXPublicKey(publicBlock.Bytes)
	if err != nil {
		log.Info("Decode public key error: ", err)
		return false, 0
	}

	publicKeyEcdsa := publicKey.(*ecdsa.PublicKey)

	R := new(big.Int)
	R, ok := R.SetString(licenseInfoSplit[2], 10) //R
	if !ok {
		log.Info("SetString: error")
		return false, 0
	}

	S := new(big.Int)
	S, ok = S.SetString(licenseInfoSplit[3], 10) //S
	if !ok {
		log.Info("SetString: error")
		return false, 0
	}

	flag := ecdsa.Verify(publicKeyEcdsa, msgHashSum, R, S)
	if flag != true {
		log.Info("could not verify signature.")
		return false, 0
	}
	log.Info("license check success!")

	return true, expireTime
}
