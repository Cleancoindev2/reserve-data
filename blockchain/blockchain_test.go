package blockchain_test

import (
	"math/big"
	"testing"

	"github.com/KyberNetwork/reserve-data/blockchain"
	"github.com/KyberNetwork/reserve-data/cmd/configuration"
	"github.com/KyberNetwork/reserve-data/common"
	cBlockChain "github.com/KyberNetwork/reserve-data/common/blockchain"
	"github.com/KyberNetwork/reserve-data/settings"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type mockSettings struct{}

func (*mockSettings) GetInternalTokenByID(tokenID string) (common.Token, error) {
	panic("implement me")
}

func (*mockSettings) GetInternalTokens() ([]common.Token, error) {
	panic("implement me")
}

func (*mockSettings) ETHToken() common.Token {
	panic("implement me")
}

func (*mockSettings) GetAddress(addressName settings.AddressName) (ethereum.Address, error) {
	switch addressName {
	case settings.Wrapper:
		return ethereum.HexToAddress("0x6172AFC8c00c46E0D07ce3AF203828198194620a"), nil
	case settings.Reserve:
		return ethereum.HexToAddress("0x63825c174ab367968EC60f061753D3bbD36A0D8F"), nil
	case settings.Pricing:
		return ethereum.HexToAddress("0x798AbDA6Cc246D0EDbA912092A2a3dBd3d11191B"), nil
	case settings.Burner:
		return ethereum.HexToAddress("0xed4f53268bfdFF39B36E8786247bA3A02Cf34B04"), nil
	case settings.Network:
		return ethereum.HexToAddress("0x818E6FECD516Ecc3849DAf6845e3EC868087B755"), nil
	case settings.Whitelist:
		return ethereum.HexToAddress("0x6e106a75d369d09a9ea1dcc16da844792aa669a3"), nil
	}
	panic("implement me")
}

func TestGetStepDetailStepFunctionData(t *testing.T) {
	const blockNum = 6233684
	const token = "0xd26114cd6EE289AccF82350c8d8487fedB8A0C07" // OMG
	var chainType = configuration.GetChainType(common.RunningMode())

	endpoints := []string{
		"https://mainnet.infura.io",
	}
	var callClients []*ethclient.Client
	for _, endpoint := range endpoints {
		client, err := ethclient.Dial(endpoint)
		if err != nil {
			t.Fatal(err)
		}
		callClients = append(callClients, client)
	}
	caller := cBlockChain.NewContractCaller(callClients, endpoints)

	base := cBlockChain.NewBaseBlockchain(
		nil,
		nil,
		map[string]*cBlockChain.Operator{},
		nil,
		nil,
		chainType,
		caller)

	bc, err := blockchain.NewBlockchain(base, &mockSettings{})
	if err != nil {
		t.Fatal(err)
	}

	opts := bc.GetCallOpts(blockNum)
	// https://etherscan.io/address/0x798AbDA6Cc246D0EDbA912092A2a3dBd3d11191B#readContract
	// -361025371366737575936
	result, err := bc.GeneratedGetStepFunctionData(opts, ethereum.HexToAddress(token), big.NewInt(13), big.NewInt(1))
	if err != nil {
		t.Fatal(err)
	}

	if result.Cmp(big.NewInt(0)) >= 0 {
		t.Errorf("expected result to be <0, got: %s", result.String())
	}
}