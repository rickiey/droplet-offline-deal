package main

import (
	"context"
	"crypto/rand"
	"droplet-offline-deal/api"
	"droplet-offline-deal/signer"
	"fmt"

	"github.com/filecoin-project/go-address"
	cborutil "github.com/filecoin-project/go-cbor-util"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/venus/venus-shared/types"
	"github.com/google/uuid"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multiaddr"
	"github.com/rickiey/loggo"

	types2 "github.com/ipfs-force-community/droplet/v2/types"
	"github.com/libp2p/go-libp2p"
	crypto2 "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	inet "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

func init() {

}

func main() {

	address.CurrentNetwork = address.Testnet

	// test-root-cid: bafykbzacec33tbo7q3uo7txi3hgr6okk6cpatzlfa2e6kejz3bu3sjuajxwy6
	// test-piece-cid: baga6ea4seaqcvjkpqx6mgr7lxnptmfkxiwghwa77fmzkxr6jz7ja5pgfl6fa4ba
	// test-piece-size: 34359738368
	api.InitChainAPI("https://rpc.ankr.com/filecoin_testnet")
	// api.InitChainAPI("https://rpc.ankr.com/filecoin")

	dur := 2880 * 180
	pieceSize := 34359738368
	datarootCid, err := cid.Parse("bafykbzacec33tbo7q3uo7txi3hgr6okk6cpatzlfa2e6kejz3bu3sjuajxwy6")
	if err != nil {
		panic(err)
	}

	pieceCid, err := cid.Parse("baga6ea4seaqcvjkpqx6mgr7lxnptmfkxiwghwa77fmzkxr6jz7ja5pgfl6fa4ba")
	if err != nil {
		panic(err)
	}
	k, err := signer.DecodePricateKey("")
	if err != nil {
		panic(err)
	}
	minerAddr, err := address.NewFromString("t01000")
	if err != nil {
		panic(err)
	}
	fmt.Println(minerAddr)
	err = ddeal(
		context.Background(),
		datarootCid,
		pieceCid,
		minerAddr,
		*k,
		uint64(dur),
		uint64(pieceSize),
		false,
	)
	if err != nil {
		panic(err)
	}
}

func NewHost() (host.Host, error) {

	pk, _, err := crypto2.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return nil, err
	}
	h, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
		libp2p.Identity(pk),
	)
	if err != nil {
		return nil, err
	}

	return h, nil
}

func ddeal(ctx context.Context, dataRootCid, pieceCid cid.Cid, minerAddr address.Address,
	clientPvk signer.Key, dur, pieceSize uint64, isVerified bool) error {

	// 180 day
	// dur := 518400

	if dur < 518400 {
		loggo.Warn("deal duration less than min 518400, set 518400")
		dur = 518400
	}
	h, err := NewHost()
	if err != nil {
		return err
	}

	minerInfo, err := api.StateMinerInfo(minerAddr)
	if err != nil {
		return fmt.Errorf("failed to StateMinerInfo %v: %w", minerInfo, err)

	}

	fmt.Println(minerInfo.Multiaddrs)

	mutaddrs, err := ConvertMultiaddr(minerInfo.Multiaddrs)
	if err != nil {
		return fmt.Errorf("failed to cConvertMultiaddr %v: %w", minerInfo.Multiaddrs, err)
	}

	addrInfo := &peer.AddrInfo{ID: *minerInfo.PeerId, Addrs: mutaddrs}

	if err := h.Connect(ctx, *addrInfo); err != nil {
		return fmt.Errorf("failed to connect to peer %s: %w", addrInfo.ID, err)
	}
	x, err := h.Peerstore().FirstSupportedProtocol(addrInfo.ID, types2.DealProtocolv121ID)
	if err != nil {
		return fmt.Errorf("getting protocols for peer %s: %w", addrInfo.ID, err)
	}

	if len(x) == 0 {
		return fmt.Errorf("cannot make a deal with storage provider %s because it does not support protocol version 1.2.0", minerAddr)
	}

	dealUuid := uuid.New()

	transfer := types2.Transfer{
		Type: storagemarket.TTManual,
	}

	var providerCollateral abi.TokenAmount

	providerCollateral = abi.NewTokenAmount(0)

	tipset, err := api.ChainHead()
	if err != nil {
		return fmt.Errorf("cannot get chain head: %w", err)
	}
	head := tipset.Height
	startEpoch := head + abi.ChainEpoch(2880*8) // head + 8 days

	s, err := h.NewStream(ctx, addrInfo.ID, types2.DealProtocolv120ID)
	if err != nil {
		return fmt.Errorf("failed to open stream to peer %s: %w", addrInfo.ID, err)
	}
	defer s.Close() // nolint

	dealProposal, err := dealProposal(ctx, clientPvk, dataRootCid, pieceCid, abi.PaddedPieceSize(pieceSize),
		minerAddr, startEpoch, int(dur), isVerified, providerCollateral, abi.NewTokenAmount(0))
	if err != nil {
		return err
	}

	dealParams := types2.DealParams{
		DealUUID:           dealUuid,
		ClientDealProposal: *dealProposal,
		DealDataRoot:       dataRootCid,
		IsOffline:          true,
		Transfer:           transfer,
		RemoveUnsealedCopy: false,
		SkipIPNIAnnounce:   false,
	}

	// log.Debug("about to submit deal proposal", "uuid", dealUuid.String())

	var resp types2.DealResponse
	if err := doRpc(ctx, s, &dealParams, &resp); err != nil {
		return fmt.Errorf("send proposal rpc: %w", err)
	}

	if !resp.Accepted {
		return fmt.Errorf("deal proposal rejected: %s", resp.Message)
	}

	fmt.Println("deal uuid: ", dealUuid)

	proposalNd, err := cborutil.AsIpld(dealProposal)
	if err == nil {
		fmt.Println("proposal cid: ", proposalNd.Cid())
	}
	return nil
}

func ConvertMultiaddr(addrs [][]byte) ([]multiaddr.Multiaddr, error) {
	multiaddrs := make([]multiaddr.Multiaddr, 0, len(addrs))
	for _, a := range addrs {
		maddr, err := multiaddr.NewMultiaddrBytes(a)
		if err != nil {
			return nil, err
		}
		multiaddrs = append(multiaddrs, maddr)
	}

	return multiaddrs, nil
}

func dealProposal(ctx context.Context,
	clientPvk signer.Key,
	rootCid cid.Cid,
	pieceCid cid.Cid,
	pieceSize abi.PaddedPieceSize,
	minerAddr address.Address,
	startEpoch abi.ChainEpoch,
	duration int,
	verified bool,
	providerCollateral abi.TokenAmount,
	storagePrice abi.TokenAmount,

) (*types.ClientDealProposal, error) {
	endEpoch := startEpoch + abi.ChainEpoch(duration)
	// deal proposal expects total storage price for deal per epoch, therefore we
	// multiply pieceSize * storagePrice (which is set per epoch per GiB) and divide by 2^30
	storagePricePerEpochForDeal := big.Div(big.Mul(big.NewInt(int64(pieceSize)), storagePrice), big.NewInt(int64(1<<30)))
	l, err := types.NewLabelFromString(rootCid.String())
	if err != nil {
		return nil, err
	}
	proposal := types.DealProposal{
		PieceCID:             pieceCid,
		PieceSize:            pieceSize,
		VerifiedDeal:         verified,
		Client:               clientPvk.Address,
		Provider:             minerAddr,
		Label:                l,
		StartEpoch:           startEpoch,
		EndEpoch:             endEpoch,
		StoragePricePerEpoch: storagePricePerEpochForDeal,
		ProviderCollateral:   providerCollateral,
	}

	buf, err := cborutil.Dump(&proposal)
	if err != nil {
		return nil, err
	}

	// sig, err := signer.WalletSign(ctx, clientAddr, buf, types.MsgMeta{Type: types.MTDealProposal, Extra: buf})
	// if err != nil {
	// 	return nil, fmt.Errorf("wallet sign failed: %w", err)
	// }

	sig, err := signer.ClientSignDeal(clientPvk, buf, types.MsgMeta{Type: types.MTDealProposal, Extra: buf})
	if err != nil {
		return nil, fmt.Errorf("wallet sign failed: %w", err)
	}
	return &types.ClientDealProposal{
		Proposal:        proposal,
		ClientSignature: *sig,
	}, nil
}

func doRpc(ctx context.Context, s inet.Stream, req interface{}, resp interface{}) error {
	errc := make(chan error)
	go func() {
		if err := cborutil.WriteCborRPC(s, req); err != nil {
			errc <- fmt.Errorf("failed to send request: %w", err)
			return
		}

		if err := cborutil.ReadCborRPC(s, resp); err != nil {
			errc <- fmt.Errorf("failed to read response: %w", err)
			return
		}

		errc <- nil
	}()

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
