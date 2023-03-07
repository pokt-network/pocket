package slip

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	// Test Vectors
	testVector1SeedHex = "000102030405060708090a0b0c0d0e0f"
	testVector2SeedHex = "fffcf9f6f3f0edeae7e4e1dedbd8d5d2cfccc9c6c3c0bdbab7b4b1aeaba8a5a29f9c999693908d8a8784817e7b7875726f6c696663605d5a5754514e4b484542"

	// SLIPS-0010
	testSeedHex = "045e8380086abc6f6e941d6fe47ca93b86723bc246ec8c4beee411b410028675"
)

func TestSlip_DeriveChild_TestVectors(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		seed        string
		wantPrivHex string
		wantPubHex  string

		wantErr bool
	}{
		// https://github.com/satoshilabs/slips/blob/master/slip-0010.md#test-vector-1-for-ed25519
		// Note that ed25519 public keys normally don't have a "00" prefix, but we are reflecting the
		// test vectors from the spec which do
		{
			name:        "TestVector1 Key derivation is deterministic for path `m` (master key)",
			path:        "m",
			seed:        testVector1SeedHex,
			wantPrivHex: "2b4be7f19ee27bbf30c667b642d5f4aa69fd169872f8fc3059c08ebae2eb19e7",
			wantPubHex:  "00a4b2856bfec510abab89753fac1ac0e1112364e7d250545963f135f2a33188ed",
			wantErr:     false,
		},
		{
			name:        "TestVector1 Key derivation is deterministic for path `m/0'`",
			path:        "m/0'",
			seed:        testVector1SeedHex,
			wantPrivHex: "68e0fe46dfb67e368c75379acec591dad19df3cde26e63b93a8e704f1dade7a3",
			wantPubHex:  "008c8a13df77a28f3445213a0f432fde644acaa215fc72dcdf300d5efaa85d350c",
			wantErr:     false,
		},
		{
			name:        "TestVector1 Key derivation is deterministic for path `m/0'/1'`",
			path:        "m/0'/1'",
			seed:        testVector1SeedHex,
			wantPrivHex: "b1d0bad404bf35da785a64ca1ac54b2617211d2777696fbffaf208f746ae84f2",
			wantPubHex:  "001932a5270f335bed617d5b935c80aedb1a35bd9fc1e31acafd5372c30f5c1187",
			wantErr:     false,
		},
		{
			name:        "TestVector1 Key derivation is deterministic for path `m/0'/1'/2'`",
			path:        "m/0'/1'/2'",
			seed:        testVector1SeedHex,
			wantPrivHex: "92a5b23c0b8a99e37d07df3fb9966917f5d06e02ddbd909c7e184371463e9fc9",
			wantPubHex:  "00ae98736566d30ed0e9d2f4486a64bc95740d89c7db33f52121f8ea8f76ff0fc1",
			wantErr:     false,
		},
		{
			name:        "TestVector1 Key derivation is deterministic for path `m/0'/1'/2'/2'`",
			path:        "m/0'/1'/2'/2'",
			seed:        testVector1SeedHex,
			wantPrivHex: "30d1dc7e5fc04c31219ab25a27ae00b50f6fd66622f6e9c913253d6511d1e662",
			wantPubHex:  "008abae2d66361c879b900d204ad2cc4984fa2aa344dd7ddc46007329ac76c429c",
			wantErr:     false,
		},
		{
			name:        "TestVector1 Key derivation is deterministic for path `m/0'/1'/2'/2'/1000000000'`",
			path:        "m/0'/1'/2'/2'/1000000000'",
			seed:        testVector1SeedHex,
			wantPrivHex: "8f94d394a8e8fd6b1bc2f3f49f5c47e385281d5c17e65324b0f62483e37e8793",
			wantPubHex:  "003c24da049451555d51a7014a37337aa4e12d41e485abccfa46b47dfb2af54b7a",
			wantErr:     false,
		},
		{
			name:        "TestVector1 Key derivation fails with invalid path `m/0`",
			path:        "m/0",
			seed:        testVector1SeedHex,
			wantPrivHex: "",
			wantErr:     true,
		},
		// https://github.com/satoshilabs/slips/blob/master/slip-0010.md#test-vector-2-for-ed25519
		{
			name:        "TestVector2 Key derivation is deterministic for path `m` (master key)",
			path:        "m",
			seed:        testVector2SeedHex,
			wantPrivHex: "171cb88b1b3c1db25add599712e36245d75bc65a1a5c9e18d76f9f2b1eab4012",
			wantPubHex:  "008fe9693f8fa62a4305a140b9764c5ee01e455963744fe18204b4fb948249308a",
			wantErr:     false,
		},
		{
			name:        "TestVector2 Key derivation is deterministic for path `m/0'`",
			path:        "m/0'",
			seed:        testVector2SeedHex,
			wantPrivHex: "1559eb2bbec5790b0c65d8693e4d0875b1747f4970ae8b650486ed7470845635",
			wantPubHex:  "0086fab68dcb57aa196c77c5f264f215a112c22a912c10d123b0d03c3c28ef1037",
			wantErr:     false,
		},
		{
			name:        "TestVector2 Key derivation is deterministic for path `m/0'/2147483647'`",
			path:        "m/0'/2147483647'",
			seed:        testVector2SeedHex,
			wantPrivHex: "ea4f5bfe8694d8bb74b7b59404632fd5968b774ed545e810de9c32a4fb4192f4",
			wantPubHex:  "005ba3b9ac6e90e83effcd25ac4e58a1365a9e35a3d3ae5eb07b9e4d90bcf7506d",
			wantErr:     false,
		},
		{
			name:        "TestVector2 Key derivation is deterministic for path `m/0'/2147483647'/1'`",
			path:        "m/0'/2147483647'/1'",
			seed:        testVector2SeedHex,
			wantPrivHex: "3757c7577170179c7868353ada796c839135b3d30554bbb74a4b1e4a5a58505c",
			wantPubHex:  "002e66aa57069c86cc18249aecf5cb5a9cebbfd6fadeab056254763874a9352b45",
			wantErr:     false,
		},
		{
			name:        "TestVector2 Key derivation is deterministic for path `m/0'/2147483647'/1'/2147483646'`",
			path:        "m/0'/2147483647'/1'/2147483646'",
			seed:        testVector2SeedHex,
			wantPrivHex: "5837736c89570de861ebc173b1086da4f505d4adb387c6a1b1342d5e4ac9ec72",
			wantPubHex:  "00e33c0f7d81d843c572275f287498e8d408654fdf0d1e065b84e2e6f157aab09b",
			wantErr:     false,
		},
		{
			name:        "TestVector2 Key derivation is deterministic for path `m/0'/2147483647'/1'/2147483646'/2'`",
			path:        "m/0'/2147483647'/1'/2147483646'/2'",
			seed:        testVector2SeedHex,
			wantPrivHex: "551d333177df541ad876a60ea71f00447931c0a9da16f227c11ea080d7391b8d",
			wantPubHex:  "0047150c75db263559a70d5778bf36abbab30fb061ad69f69ece61a72b0cfa4fc0",
			wantErr:     false,
		},
		// Pocket specific test vectors
		{
			name:        "PoktTestVector Key derivation is deterministic for path `m` (master key)",
			path:        "m",
			seed:        testSeedHex,
			wantPrivHex: "849e1cf075e4ba4a310c2f89036a2a7eb2bb18ccfcfdd793cdc48a28978e588e",
			wantPubHex:  "009ab1c51476edbb1f71f91b33d3bb342503cba9adc8cd4a4c6e62a8cc7d27859b",
			wantErr:     false,
		},
		{
			name:        "PoktTestVector Key derivation is deterministic for path `m/44'`",
			path:        "m/44'",
			seed:        testSeedHex,
			wantPrivHex: "35852227a2289bcba3f0a428e08cc864f2ed35f605552c378232856a2abf5564",
			wantPubHex:  "00d0b243e9de9f2f12410275025395bd088cf57b0272794cc66d66093a51579797",
			wantErr:     false,
		},
		{
			name:        "PoktTestVector Key derivation is deterministic for path `m/44'/635'`",
			path:        "m/44'/635'",
			seed:        testSeedHex,
			wantPrivHex: "9af532407b4dba729962f9d57496394e7b82ea7a995575d0bfcc8fd4845debe6",
			wantPubHex:  "00a428992b7afb7716cf339d2b54bede7f2932cf4cf3c542f5662dcebd7e3abaee",
			wantErr:     false,
		},
		{
			name:        "PoktTestVector Child key derivation is deterministic for index `0` (first child)",
			path:        fmt.Sprintf(PoktAccountPathFormat, 0),
			seed:        testSeedHex,
			wantPrivHex: "e7e0f734311bdf2f446821a464b490ae8c005f6af92a26f42a39067103726346",
			wantPubHex:  "000848875b5836bc71fecb10f563b84b9b8a63b6d7a1b16b1e88eeaca6a7ad5852",
			wantErr:     false,
		},
		{
			name:        "PoktTestVector Child key derivation is deterministic for index `1000000`",
			path:        fmt.Sprintf(PoktAccountPathFormat, 1000000),
			seed:        testSeedHex,
			wantPrivHex: "f208bf7d7afa1a12a0b74e8fbbf1bb15bdfcaa3d03507ec338d7ecd77331e964",
			wantPubHex:  "00e7d8f01dfbc0eb1638d9853b95cdec650ad47d72fac1d2d1c97c8f935d2cbf90",
			wantErr:     false,
		},
		{
			name:        "PoktTestVector Child key derivation is deterministic for index `2147483647` (last child)",
			path:        fmt.Sprintf(PoktAccountPathFormat, 2147483647),
			seed:        testSeedHex,
			wantPrivHex: "20e061dcfae5cc90cba8ee374afc2d67d518b5f542f3217b8c830293c3dbb7e6",
			wantPubHex:  "0008a399a3cdc1ee9a50c922d018daa98b5069bfea37944fde14a73d83ca3ec08c",
			wantErr:     false,
		},
		{
			name:        "PoktTestVector Child index is too large to derive ed25519 key for index `2147483648` ",
			path:        fmt.Sprintf(PoktAccountPathFormat, 2147483648),
			seed:        testSeedHex,
			wantPrivHex: "",
			wantPubHex:  "",
			wantErr:     true,
		},
		{
			name:        "PoktTestVector Child index is too large to derive ed25519 key for index `4294967295` ",
			path:        fmt.Sprintf(PoktAccountPathFormat, ^uint32(0)),
			seed:        testSeedHex,
			wantPrivHex: "",
			wantPubHex:  "",
			wantErr:     true,
		},
	}
	for _, tv := range tests {
		t.Run(tv.name, func(t *testing.T) {
			seed, err := hex.DecodeString(tv.seed)
			require.NoError(t, err)
			childKey, err := DeriveChild(tv.path, seed)
			if tv.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if err != nil {
				return
			}

			// Slip-0010 private keys in test vector are only the seed of the full private key
			// This is equivalent to the SecretKey of the HMAC key used to generate the ed25519 key
			privSeed, err := childKey.GetSeed("")
			require.NoError(t, err)
			privHex := hex.EncodeToString(privSeed)
			require.Equal(t, tv.wantPrivHex, privHex)

			// Slip-0010 keys are prefixed with "00" in the test vectors
			pubHex := childKey.GetPublicKey().String()
			require.Equal(t, tv.wantPubHex, "00"+pubHex)
		})
	}
}
