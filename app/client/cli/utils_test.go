package cli

import (
	"bytes"
	"io"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/stretchr/testify/require"
)

func Test_parseEd25519PrivateKeyFromReader(t *testing.T) {
	type args struct {
		reader io.Reader
	}

	validPKString := `"e7760141c2672178b28360a8cf80ff3a9d5fd579990317b9afcb2091426ffe75dc12b26584c057be33fcc8e891a483250581e38fe2bc9d62c1a1341c5e85b667"`

	pk, err := strconv.Unquote(validPKString)
	require.NoError(t, err)

	validPk, err := crypto.NewPrivateKey(pk)
	require.NoError(t, err)

	tests := []struct {
		name    string
		args    args
		wantPk  crypto.Ed25519PrivateKey
		wantErr bool
	}{
		{
			name: "should err if invalid: nil",
			args: args{
				reader: nil,
			},
			wantPk:  nil,
			wantErr: true,
		},
		{
			name: "should err if invalid: empty byteArr",
			args: args{
				reader: bytes.NewReader([]byte{}),
			},
			wantPk:  nil,
			wantErr: true,
		},
		{
			name: "should return valid private key",
			args: args{
				reader: strings.NewReader(validPKString),
			},
			wantPk:  validPk.Bytes(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPk, err := parseEd25519PrivateKeyFromReader(tt.args.reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseEd25519PrivateKeyFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotPk, tt.wantPk) {
				t.Errorf("parseEd25519PrivateKeyFromFile() = %v, want %v", gotPk, tt.wantPk)
			}
		})
	}
}
