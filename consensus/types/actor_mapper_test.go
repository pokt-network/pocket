package types

import (
	"reflect"
	"testing"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

func makeTestValidatorWithAddress(address string) *coreTypes.Actor {
	return &coreTypes.Actor{
		ActorType:       0,
		Address:         address,
		PublicKey:       "",
		Chains:          []string{},
		GenericParam:    "",
		StakedAmount:    "",
		PausedHeight:    0,
		UnstakingHeight: 0,
		Output:          "",
	}
}

func Test_actorMapper_GetValidatorMap(t *testing.T) {
	type args struct {
		validators []*coreTypes.Actor
	}
	tests := []struct {
		name string
		args args
		want ValidatorMap
	}{
		{
			name: "empty validator slice should return empty map",
			args: args{
				validators: []*coreTypes.Actor{},
			},
			want: map[string]*coreTypes.Actor{},
		},
		{
			name: "one validator should return map with one entry",
			args: args{
				validators: []*coreTypes.Actor{
					makeTestValidatorWithAddress("0x1"),
				},
			},
			want: map[string]*coreTypes.Actor{
				"0x1": makeTestValidatorWithAddress("0x1"),
			},
		},
		{
			// Note: the order of the validators is irrelevant since we are dealing with maps in this particular test, maps are not sorted in Go anyway
			name: "multiple validators should return map with all of them",
			args: args{
				validators: []*coreTypes.Actor{
					makeTestValidatorWithAddress("0x3"),
					makeTestValidatorWithAddress("0x2"),
					makeTestValidatorWithAddress("0x1"),
					makeTestValidatorWithAddress("0x4"),
					makeTestValidatorWithAddress("0x5"),
				},
			},
			want: map[string]*coreTypes.Actor{
				"0x1": makeTestValidatorWithAddress("0x1"),
				"0x2": makeTestValidatorWithAddress("0x2"),
				"0x3": makeTestValidatorWithAddress("0x3"),
				"0x4": makeTestValidatorWithAddress("0x4"),
				"0x5": makeTestValidatorWithAddress("0x5"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			am := NewActorMapper(tt.args.validators)

			if got := am.GetValidatorMap(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("actorMapper.GetValidatorMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_actorMapper_GetValAddrToIdMap(t *testing.T) {
	type args struct {
		validators []*coreTypes.Actor
	}
	tests := []struct {
		name string
		args args
		want ValAddrToIdMap
	}{
		{
			name: "empty validator slice should return empty map",
			args: args{},
			want: map[string]NodeId{},
		},
		{
			name: "one validator should return map with one entry",
			args: args{
				validators: []*coreTypes.Actor{
					makeTestValidatorWithAddress("0x1"),
				},
			},
			want: map[string]NodeId{
				"0x1": 1,
			},
		},
		{
			// Note: this test is important because it tests the sorting of the validators by address that's used for generating tds
			name: "multiple validators should return map with all of them and with the correct NodeIds",
			args: args{
				validators: []*coreTypes.Actor{
					makeTestValidatorWithAddress("0x5"),
					makeTestValidatorWithAddress("0x1"),
					makeTestValidatorWithAddress("0x3"),
					makeTestValidatorWithAddress("0x2"),
					makeTestValidatorWithAddress("0x4"),
				},
			},
			want: map[string]NodeId{
				"0x1": 1,
				"0x2": 2,
				"0x3": 3,
				"0x4": 4,
				"0x5": 5,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			am := NewActorMapper(tt.args.validators)

			if got := am.GetValAddrToIdMap(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("actorMapper.GetValAddrToIdMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_actorMapper_GetIdToValAddrMap(t *testing.T) {
	type args struct {
		validators []*coreTypes.Actor
	}
	tests := []struct {
		name string
		args args
		want IdToValAddrMap
	}{
		{
			name: "empty validator slice should return empty map",
			args: args{
				validators: []*coreTypes.Actor{},
			},
			want: map[NodeId]string{},
		},
		{
			name: "one validator should return map with one entry",
			args: args{
				validators: []*coreTypes.Actor{
					makeTestValidatorWithAddress("0x1"),
				},
			},
			want: map[NodeId]string{
				1: "0x1",
			},
		},
		{
			// Note: this test is important because it tests the sorting of the validators by address that's used for generating NodeIds
			name: "multiple validators should return map with all of them and with the correct NodeIds",
			args: args{
				validators: []*coreTypes.Actor{
					makeTestValidatorWithAddress("0x5"),
					makeTestValidatorWithAddress("0x1"),
					makeTestValidatorWithAddress("0x3"),
					makeTestValidatorWithAddress("0x2"),
					makeTestValidatorWithAddress("0x4"),
				},
			},
			want: IdToValAddrMap{
				1: "0x1",
				2: "0x2",
				3: "0x3",
				4: "0x4",
				5: "0x5",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			am := NewActorMapper(tt.args.validators)

			if got := am.GetIdToValAddrMap(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("actorMapper.GetIdToValAddrMap() = %v, want %v", got, tt.want)
			}
		})
	}
}
