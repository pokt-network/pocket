package kvstore_test

import (
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pokt-network/pocket/persistence/kvstore"
)

var count = 100

func TestPostgresKV_Get(t *testing.T) {
	type fields struct {
		Pool *pgxpool.Pool
	}
	type args struct {
		key []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &kvstore.PostgresKV{
				Pool: tt.fields.Pool,
			}

			seedKVs(t, p)

			got, err := p.Get(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("PostgresKV.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PostgresKV.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPostgresKV_Set(t *testing.T) {
	type fields struct {
		Pool *pgxpool.Pool
	}
	type args struct {
		key   []byte
		value []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "",
			fields: fields{
				Pool: p,
			},
			args: args{
				key:   []byte{},
				value: []byte{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &kvstore.PostgresKV{
				Pool: tt.fields.Pool,
			}
			if err := p.Set(tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("PostgresKV.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPostgresKV_Delete(t *testing.T) {
	type fields struct {
		Pool *pgxpool.Pool
	}
	type args struct {
		key []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &kvstore.PostgresKV{
				Pool: tt.fields.Pool,
			}
			if err := p.Delete(tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("PostgresKV.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPostgresKV_Stop(t *testing.T) {
	type fields struct {
		Pool *pgxpool.Pool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &kvstore.PostgresKV{
				Pool: tt.fields.Pool,
			}
			if err := p.Stop(); (err != nil) != tt.wantErr {
				t.Errorf("PostgresKV.Stop() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPostgresKV_GetAll(t *testing.T) {
	type fields struct {
		Pool *pgxpool.Pool
	}
	type args struct {
		prefixKey  []byte
		descending bool
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantKeys   [][]byte
		wantValues [][]byte
		wantErr    bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &kvstore.PostgresKV{
				Pool: tt.fields.Pool,
			}
			gotKeys, gotValues, err := p.GetAll(tt.args.prefixKey, tt.args.descending)
			if (err != nil) != tt.wantErr {
				t.Errorf("PostgresKV.GetAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotKeys, tt.wantKeys) {
				t.Errorf("PostgresKV.GetAll() gotKeys = %v, want %v", gotKeys, tt.wantKeys)
			}
			if !reflect.DeepEqual(gotValues, tt.wantValues) {
				t.Errorf("PostgresKV.GetAll() gotValues = %v, want %v", gotValues, tt.wantValues)
			}
		})
	}
}

func TestPostgresKV_Exists(t *testing.T) {
	type fields struct {
		Pool *pgxpool.Pool
	}
	type args struct {
		key []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &kvstore.PostgresKV{
				Pool: tt.fields.Pool,
			}
			got, err := p.Exists(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("PostgresKV.Exists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("PostgresKV.Exists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPostgresKV_ClearAll(t *testing.T) {
	type fields struct {
		Pool *pgxpool.Pool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &kvstore.PostgresKV{
				Pool: tt.fields.Pool,
			}
			if err := p.ClearAll(); (err != nil) != tt.wantErr {
				t.Errorf("PostgresKV.ClearAll() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func seedKVs(t *testing.T, kv *kvstore.PostgresKV) {
	for i := 0; i < count; i++ {
		// insert a test value
		err := kv.Set(generateRandomBytes(10), generateRandomBytes(20))
		if err != nil {
			t.Fail()
		}
	}
}

func generateRandomBytes(length int) []byte {
	bytes := make([]byte, length)
	rand.Seed(time.Now().UnixNano())
	rand.Read(bytes)
	return bytes
}

func connectToPool() *pgxpool.Pool {
	panic("not impl")
	return nil
}
