package settings

import (
	ethereum "github.com/ethereum/go-ethereum/common"
)

func (setting *Settings) UpdateAdress(name string, address ethereum.Address) error {
	return setting.Address.Storage.UpdateOneAddress(name, address.Hex())
}

func (setting *Settings) GetAddress(name string) (ethereum.Address, error) {
	addr, err := setting.Address.Storage.GetAddress(name)
	result := ethereum.Address{}
	if err != nil {
		return result, err
	}
	return ethereum.HexToAddress(addr), err
}

func (setting *Settings) AddAddressToSet(setName string, address ethereum.Address) error {
	return setting.Address.Storage.AddAddressToSet(setName, address.Hex())
}

func (setting *Settings) GetAddresses(setName string) ([]ethereum.Address, error) {
	result := []ethereum.Address{}
	addrs, err := setting.Address.Storage.GetAddresses(setName)
	if err != nil {
		return result, err
	}
	for _, addr := range addrs {
		result = append(result, ethereum.HexToAddress(addr))
	}
	return result, nil
}