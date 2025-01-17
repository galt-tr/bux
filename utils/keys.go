package utils

import (
	"github.com/bitcoinschema/go-bitcoin/v2"
	"github.com/libsv/go-bk/bec"
	"github.com/libsv/go-bk/bip32"
)

// DeriveChildKeyFromHex derive the child extended key from the hex string
func DeriveChildKeyFromHex(hdKey *bip32.ExtendedKey, hexHash string) (*bip32.ExtendedKey, error) {
	var childKey *bip32.ExtendedKey
	childKey = hdKey

	childNums, err := GetChildNumsFromHex(hexHash)
	if err != nil {
		return nil, err
	}

	for num := range childNums {
		childKey, err = childKey.Child(uint32(num))
		if err != nil {
			return nil, err
		}
	}

	return childKey, nil
}

// DerivePublicKey will derive the internal and external address from a key
func DerivePublicKey(hdKey *bip32.ExtendedKey, chain uint32, num uint32) (*bec.PublicKey, error) {
	if hdKey == nil {
		return nil, ErrHDKeyNil
	}

	pubKeys, err := bitcoin.GetPublicKeysForPath(hdKey, num)
	if err != nil {
		return nil, err
	}

	return pubKeys[chain], nil
}

// ValidateXPub will check the xPub key for length & validation
func ValidateXPub(rawKey string) (*bip32.ExtendedKey, error) {

	// Validate the xpub (length)
	if len(rawKey) != xPubKeyLength {
		return nil, ErrXpubInvalidLength
	}

	// Parse the xPub into an HD key
	hdKey, err := bitcoin.GetHDKeyFromExtendedPublicKey(rawKey)
	if err != nil {
		return nil, err
	} else if hdKey.String() != rawKey { // Sanity check (might not be needed)
		return nil, ErrXpubNoMatch
	}
	return hdKey, nil
}

// DeriveAddresses will derive the internal and external address from a key
func DeriveAddresses(hdKey *bip32.ExtendedKey, num uint32) (external, internal string, err error) {

	// Don't panic
	if hdKey == nil {
		return "", "", ErrHDKeyNil
	}

	// Derive the address
	var addresses []string
	if addresses, err = bitcoin.GetAddressesForPath(
		hdKey, num,
	); err != nil {
		return
	} else if len(addresses) != 2 { // Sanity check might not be needed
		return "", "", ErrDeriveFailed
	}
	external = addresses[0]
	internal = addresses[1]
	return
}

// DerivePrivateKeyFromHex will derive the private key from the extended key using the hex as the derivation paths
func DerivePrivateKeyFromHex(hdKey *bip32.ExtendedKey, hexString string) (*bec.PrivateKey, error) {
	if hdKey == nil {
		return nil, ErrHDKeyNil
	}

	childKey, err := DeriveChildKeyFromHex(hdKey, hexString)
	if err != nil {
		return nil, err
	}

	var privKey *bec.PrivateKey
	if privKey, err = childKey.ECPrivKey(); err != nil {
		return nil, err
	}

	return privKey, nil
}

// DerivePublicKeyFromHex will derive the public key from the extended key using the hex as the derivation paths
func DerivePublicKeyFromHex(hdKey *bip32.ExtendedKey, hexString string) (*bec.PublicKey, error) {
	if hdKey == nil {
		return nil, ErrHDKeyNil
	}

	childKey, err := DeriveChildKeyFromHex(hdKey, hexString)
	if err != nil {
		return nil, err
	}

	var pubKey *bec.PublicKey
	if pubKey, err = childKey.ECPubKey(); err != nil {
		return nil, err
	}

	return pubKey, nil
}
