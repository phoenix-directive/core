/*******************************************************************************
*   (c) 2018 ZondaX GmbH
*
*  Licensed under the Apache License, Version 2.0 (the "License");
*  you may not use this file except in compliance with the License.
*  You may obtain a copy of the License at
*
*      http://www.apache.org/licenses/LICENSE-2.0
*
*  Unless required by applicable law or agreed to in writing, software
*  distributed under the License is distributed on an "AS IS" BASIS,
*  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
*  See the License for the specific language governing permissions and
*  limitations under the License.
********************************************************************************/

package ledger_terra_go

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Ledger Test Mnemonic: equip will roof matter pink blind book anxiety banner elbow sun young

func Test_UserFindLedger(t *testing.T) {
	userApp, err := FindLedgerCosmosUserApp()
	if err != nil {
		t.Fatalf(err.Error())
	}

	assert.NotNil(t, userApp)
	defer userApp.Close()
}

func Test_UserGetVersion(t *testing.T) {
	userApp, err := FindLedgerCosmosUserApp()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer userApp.Close()

	userApp.api.Logging = true

	err = userApp.LoadVersion()
	require.Nil(t, err, "Detected error")

	version := userApp.version

	if userApp.appName == "Terra" {
		assert.Equal(t, uint8(0x0), version.AppMode, "TESTING MODE ENABLED!!")
		assert.Equal(t, uint8(0x1), version.Major, "Wrong Major version")
		assert.Equal(t, uint8(0x0), version.Minor, "Wrong Minor version")
		assert.Equal(t, uint8(0x0), version.Patch, "Wrong Patch version")
	} else if userApp.appName == "Cosmos" {
		assert.Equal(t, uint8(0x0), version.AppMode, "TESTING MODE ENABLED!!")
		assert.Equal(t, uint8(0x2), version.Major, "Wrong Major version")
		assert.Equal(t, uint8(0xC), version.Minor, "Wrong Minor version")
		assert.Equal(t, uint8(0x0), version.Patch, "Wrong Patch version")
	} else {
		assert.Fail(t, "MUST NOT ENTER HERE")
	}

}

func Test_UserGetPublicKey(t *testing.T) {
	userApp, err := FindLedgerCosmosUserApp()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer userApp.Close()

	userApp.api.Logging = true

	path := []uint32{44, 330, 5, 0, 21}

	pubKey, err := userApp.GetPublicKeySECP256K1(path)
	if err != nil {
		t.Fatalf("Detected error, err: %s\n", err.Error())
	}

	assert.Equal(t, 33, len(pubKey),
		"Public key has wrong length: %x, expected length: %x\n", pubKey, 65)
	fmt.Printf("PUBLIC KEY: %x\n", pubKey)

	assert.Equal(t,
		"038c2eead695e4f8e9318b8cbb6dc8b7321cbee92bb88230a3e3c3d91b8de859c0",
		hex.EncodeToString(pubKey),
		"Unexpected pubkey")
}

func Test_GetAddressPubKeySECP256K1_Zero(t *testing.T) {
	userApp, err := FindLedgerCosmosUserApp()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer userApp.Close()

	userApp.api.Logging = true

	hrp := "terra"
	path := []uint32{44, 330, 0, 0, 0}

	pubKey, addr, err := userApp.GetAddressPubKeySECP256K1(path, hrp)
	if err != nil {
		t.Fatalf("Detected error, err: %s\n", err.Error())
	}

	fmt.Printf("PUBLIC KEY : %x\n", pubKey)
	fmt.Printf("BECH32 ADDR: %s\n", addr)

	assert.Equal(t, 33, len(pubKey), "Public key has wrong length: %x, expected length: %x\n", pubKey, 65)

	assert.Equal(t, "03028f0d5a9fd41600191cdefdea05e77a68dfbce286241c0190805b9346667d07", hex.EncodeToString(pubKey), "Unexpected pubkey")
	assert.Equal(t, "terra1uayrf8zh44620zyjd052gdcjcrvjpgkkg78fux", addr, "Unexpected addr")
}

func Test_GetAddressPubKeySECP256K1(t *testing.T) {
	userApp, err := FindLedgerCosmosUserApp()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer userApp.Close()

	userApp.api.Logging = true

	hrp := "terra"
	path := []uint32{44, 330, 5, 0, 21}

	pubKey, addr, err := userApp.GetAddressPubKeySECP256K1(path, hrp)
	if err != nil {
		t.Fatalf("Detected error, err: %s\n", err.Error())
	}

	fmt.Printf("PUBLIC KEY : %x\n", pubKey)
	fmt.Printf("BECH32 ADDR: %s\n", addr)

	assert.Equal(t, 33, len(pubKey), "Public key has wrong length: %x, expected length: %x\n", pubKey, 65)

	assert.Equal(t, "038c2eead695e4f8e9318b8cbb6dc8b7321cbee92bb88230a3e3c3d91b8de859c0", hex.EncodeToString(pubKey), "Unexpected pubkey")
	assert.Equal(t, "terra1ql0ggyut5pkytffp94q3xv5zzwezq2cwv5qxc0", addr, "Unexpected addr")
}

func Test_UserPK_HDPaths(t *testing.T) {
	userApp, err := FindLedgerCosmosUserApp()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer userApp.Close()

	userApp.api.Logging = true

	path := []uint32{44, 330, 0, 0, 0}

	expected := []string{
		"03028f0d5a9fd41600191cdefdea05e77a68dfbce286241c0190805b9346667d07",
		"025488611a1e1127703e6c2a0dcc27063035d16536da11feae655ac58fd3f46eb4",
		"02af5763cc2bb3e83ec410bcabb57d622b1eacd4fc4aa180fabfb3c879b05e4828",
		"03eb7aed84a5c31264cf19e8c03b487c1485824415def544f0c48be815f98c4af4",
		"020f80062dc2696157ff04b66772478ed01141c369dc2d53fafbb2468dffae64fa",
		"02792cddcb672fa88c8bd13aff708eb6552b2dd4a9d347a844f1d591c202f58c1e",
		"024dbb0781c9b46e073dd8816280942b6497cfbbb486882180be7e6fbd48def104",
		"0349938f2c1a5468b1137bf70e73cdc4820b22e9dd2a10750026139075e2140aec",
		"029fc277428deeae4afbe18aa852c0acd4ef4d9115bd0ba3ae352e72dcb4474650",
		"03e07031287989c036e1ad68811981ef3bdb026241e5ff0e21c0d043f3c4790ed1",
	}

	for i := uint32(0); i < 10; i++ {
		path[4] = i

		pubKey, err := userApp.GetPublicKeySECP256K1(path)
		if err != nil {
			t.Fatalf("Detected error, err: %s\n", err.Error())
		}

		assert.Equal(
			t,
			33,
			len(pubKey),
			"Public key has wrong length: %x, expected length: %x\n", pubKey, 65)

		assert.Equal(
			t,
			expected[i],
			hex.EncodeToString(pubKey),
			"Public key 44'/330'/0'/0/%d does not match\n", i)

		_, err = btcec.ParsePubKey(pubKey[:], btcec.S256())
		require.Nil(t, err, "Error parsing public key err: %s\n", err)

	}
}

func getDummyTx() []byte {
	dummyTx := `{
		"account_number": 1,
		"chain_id": "some_chain",
		"fee": {
			"amount": [{"amount": 10, "denom": "DEN"}],
			"gas": 5
		},
		"memo": "MEMO",
		"msgs": ["SOMETHING"],
		"sequence": 3
	}`
	dummyTx = strings.Replace(dummyTx, " ", "", -1)
	dummyTx = strings.Replace(dummyTx, "\n", "", -1)
	dummyTx = strings.Replace(dummyTx, "\t", "", -1)

	return []byte(dummyTx)
}

func Test_UserSign(t *testing.T) {
	userApp, err := FindLedgerCosmosUserApp()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer userApp.Close()

	userApp.api.Logging = true

	path := []uint32{44, 118, 0, 0, 5}

	message := getDummyTx()
	signature, err := userApp.SignSECP256K1(path, message, 0)
	if err != nil {
		t.Fatalf("[Sign] Error: %s\n", err.Error())
	}

	// Verify Signature
	pubKey, err := userApp.GetPublicKeySECP256K1(path)
	if err != nil {
		t.Fatalf("Detected error, err: %s\n", err.Error())
	}

	if err != nil {
		t.Fatalf("[GetPK] Error: " + err.Error())
		return
	}

	pub2, err := btcec.ParsePubKey(pubKey[:], btcec.S256())
	if err != nil {
		t.Fatalf("[ParsePK] Error: " + err.Error())
		return
	}

	sig2, err := btcec.ParseDERSignature(signature[:], btcec.S256())
	if err != nil {
		t.Fatalf("[ParseSig] Error: " + err.Error())
		return
	}

	hash := sha256.Sum256(message)
	verified := sig2.Verify(hash[:], pub2)
	if !verified {
		t.Fatalf("[VerifySig] Error verifying signature: " + err.Error())
		return
	}
}

func Test_UserSign_Fails(t *testing.T) {
	userApp, err := FindLedgerCosmosUserApp()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer userApp.Close()

	userApp.api.Logging = true

	path := []uint32{44, 118, 0, 0, 5}

	message := getDummyTx()
	garbage := []byte{65}
	message = append(garbage, message...)

	_, err = userApp.SignSECP256K1(path, message, 0)
	assert.Error(t, err)
	errMessage := err.Error()

	if errMessage != "Invalid character in JSON string" && errMessage != "Unexpected characters" {
		assert.Fail(t, "Unexpected error message returned: "+errMessage)
	}
}
