package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func FuzzHeight_ToStringDeterministic(f *testing.F) {
	for i := 0; i < 100; i++ {
		f.Add(uint64(i))
	}
	f.Fuzz(func(t *testing.T, i uint64) {
		height := &Height{
			RevisionNumber: i,
			RevisionHeight: i,
		}
		str := height.ToString()
		require.Equal(t, str, fmt.Sprintf("%d-%d", i, i))
	})
}

func TestHeight_IsZero(t *testing.T) {
	testCases := []struct {
		name     string
		height   *Height
		expected bool
	}{
		{
			name: "zero height",
			height: &Height{
				RevisionNumber: 0,
				RevisionHeight: 0,
			},
			expected: true,
		},
		{
			name: "non-zero height: zero revision number",
			height: &Height{
				RevisionNumber: 0,
				RevisionHeight: 1,
			},
			expected: false,
		},
		{
			name: "non-zero height: zero revision height",
			height: &Height{
				RevisionNumber: 1,
				RevisionHeight: 0,
			},
			expected: false,
		},
		{
			name: "non-zero height",
			height: &Height{
				RevisionNumber: 1,
				RevisionHeight: 1,
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.height.IsZero())
		})
	}
}

func TestHeight_Increment(t *testing.T) {
	height := &Height{
		RevisionNumber: 1,
		RevisionHeight: 1,
	}
	newHeight := height.Increment()
	require.Equal(t, uint64(1), height.GetRevisionNumber())
	require.Equal(t, uint64(2), newHeight.GetRevisionHeight())

	newHeight = newHeight.Increment()
	require.Equal(t, uint64(1), height.GetRevisionNumber())
	require.Equal(t, uint64(3), newHeight.GetRevisionHeight())
}

func TestHeight_Decrement(t *testing.T) {
	height := &Height{
		RevisionNumber: 1,
		RevisionHeight: 2,
	}
	newHeight := height.Decrement()
	require.Equal(t, uint64(1), height.GetRevisionNumber())
	require.Equal(t, uint64(1), newHeight.GetRevisionHeight())

	newHeight = newHeight.Decrement()
	require.Equal(t, uint64(1), height.GetRevisionNumber())
	require.Equal(t, uint64(0), newHeight.GetRevisionHeight())

	newHeight = newHeight.Decrement()
	require.Equal(t, uint64(1), height.GetRevisionNumber())
	require.Equal(t, uint64(0), newHeight.GetRevisionHeight())
}

func TestHeight_Comparisons(t *testing.T) {
	testCases := []struct {
		name     string
		op       string
		height   *Height
		other    *Height
		expected bool
	}{
		{
			name: "LT: height < other",
			op:   "LT",
			height: &Height{
				RevisionNumber: 1,
				RevisionHeight: 1,
			},
			other: &Height{
				RevisionNumber: 2,
				RevisionHeight: 2,
			},
			expected: true,
		},
		{
			name: "LT: height == other",
			op:   "LT",
			height: &Height{
				RevisionNumber: 1,
				RevisionHeight: 1,
			},
			other: &Height{
				RevisionNumber: 1,
				RevisionHeight: 1,
			},
			expected: false,
		},
		{
			name: "LT: height > other",
			op:   "LT",
			height: &Height{
				RevisionNumber: 2,
				RevisionHeight: 2,
			},
			other: &Height{
				RevisionNumber: 1,
				RevisionHeight: 1,
			},
			expected: false,
		},
		{
			name: "LT: height < other (same revision number)",
			op:   "LT",
			height: &Height{
				RevisionNumber: 1,
				RevisionHeight: 1,
			},
			other: &Height{
				RevisionNumber: 1,
				RevisionHeight: 2,
			},
			expected: true,
		},
		{
			name: "LT: height > other (same revision number)",
			op:   "LT",
			height: &Height{
				RevisionNumber: 1,
				RevisionHeight: 2,
			},
			other: &Height{
				RevisionNumber: 1,
				RevisionHeight: 1,
			},
			expected: false,
		},
		{
			name: "LT: height > other (same revision height)",
			op:   "LT",
			height: &Height{
				RevisionNumber: 2,
				RevisionHeight: 1,
			},
			other: &Height{
				RevisionNumber: 1,
				RevisionHeight: 1,
			},
			expected: false,
		},
		{
			name: "LT: height < other (same revision height)",
			op:   "LT",
			height: &Height{
				RevisionNumber: 1,
				RevisionHeight: 1,
			},
			other: &Height{
				RevisionNumber: 2,
				RevisionHeight: 1,
			},
			expected: true,
		},
		{
			name: "LTE: height < other",
			op:   "LTE",
			height: &Height{
				RevisionNumber: 1,
				RevisionHeight: 1,
			},
			other: &Height{
				RevisionNumber: 2,
				RevisionHeight: 2,
			},
			expected: true,
		},
		{
			name: "LTE: height == other",
			op:   "LTE",
			height: &Height{
				RevisionNumber: 1,
				RevisionHeight: 1,
			},
			other: &Height{
				RevisionNumber: 1,
				RevisionHeight: 1,
			},
			expected: true,
		},
		{
			name: "LTE: height > other",
			op:   "LTE",
			height: &Height{
				RevisionNumber: 2,
				RevisionHeight: 2,
			},
			other: &Height{
				RevisionNumber: 1,
				RevisionHeight: 1,
			},
			expected: false,
		},
		{
			name: "GT: height < other",
			op:   "GT",
			height: &Height{
				RevisionNumber: 1,
				RevisionHeight: 1,
			},
			other: &Height{
				RevisionNumber: 2,
				RevisionHeight: 2,
			},
			expected: false,
		},
		{
			name: "GT: height == other",
			op:   "GT",
			height: &Height{
				RevisionNumber: 1,
				RevisionHeight: 1,
			},
			other: &Height{
				RevisionNumber: 1,
				RevisionHeight: 1,
			},
			expected: false,
		},
		{
			name: "GT: height > other",
			op:   "GT",
			height: &Height{
				RevisionNumber: 2,
				RevisionHeight: 2,
			},
			other: &Height{
				RevisionNumber: 1,
				RevisionHeight: 1,
			},
			expected: true,
		},
		{
			name: "GT: height < other (same revision number)",
			op:   "GT",
			height: &Height{
				RevisionNumber: 1,
				RevisionHeight: 1,
			},
			other: &Height{
				RevisionNumber: 1,
				RevisionHeight: 2,
			},
			expected: false,
		},
		{
			name: "GT: height > other (same revision number)",
			op:   "GT",
			height: &Height{
				RevisionNumber: 1,
				RevisionHeight: 2,
			},
			other: &Height{
				RevisionNumber: 1,
				RevisionHeight: 1,
			},
			expected: true,
		},
		{
			name: "GT: height > other (same revision height)",
			op:   "GT",
			height: &Height{
				RevisionNumber: 2,
				RevisionHeight: 1,
			},
			other: &Height{
				RevisionNumber: 1,
				RevisionHeight: 1,
			},
			expected: true,
		},
		{
			name: "GT: height < other (same revision height)",
			op:   "GT",
			height: &Height{
				RevisionNumber: 1,
				RevisionHeight: 1,
			},
			other: &Height{
				RevisionNumber: 2,
				RevisionHeight: 1,
			},
			expected: false,
		},
		{
			name: "GTE: height < other",
			op:   "GTE",
			height: &Height{
				RevisionNumber: 1,
				RevisionHeight: 1,
			},
			other: &Height{
				RevisionNumber: 2,
				RevisionHeight: 2,
			},
			expected: false,
		},
		{
			name: "GTE: height == other",
			op:   "GTE",
			height: &Height{
				RevisionNumber: 1,
				RevisionHeight: 1,
			},
			other: &Height{
				RevisionNumber: 1,
				RevisionHeight: 1,
			},
			expected: true,
		},
		{
			name: "GTE: height > other",
			op:   "GTE",
			height: &Height{
				RevisionNumber: 2,
				RevisionHeight: 2,
			},
			other: &Height{
				RevisionNumber: 1,
				RevisionHeight: 1,
			},
			expected: true,
		},
		{
			name: "EQ: height < other",
			op:   "EQ",
			height: &Height{
				RevisionNumber: 1,
				RevisionHeight: 1,
			},
			other: &Height{
				RevisionNumber: 2,
				RevisionHeight: 2,
			},
			expected: false,
		},
		{
			name: "EQ: height == other",
			op:   "EQ",
			height: &Height{
				RevisionNumber: 1,
				RevisionHeight: 1,
			},
			other: &Height{
				RevisionNumber: 1,
				RevisionHeight: 1,
			},
			expected: true,
		},
		{
			name: "EQ: height > other",
			op:   "EQ",
			height: &Height{
				RevisionNumber: 2,
				RevisionHeight: 2,
			},
			other: &Height{
				RevisionNumber: 1,
				RevisionHeight: 1,
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			switch tc.op {
			case "LT":
				require.Equal(t, tc.expected, tc.height.LT(tc.other))
			case "LTE":
				require.Equal(t, tc.expected, tc.height.LTE(tc.other))
			case "GT":
				require.Equal(t, tc.expected, tc.height.GT(tc.other))
			case "GTE":
				require.Equal(t, tc.expected, tc.height.GTE(tc.other))
			case "EQ":
				require.Equal(t, tc.expected, tc.height.EQ(tc.other))
			default:
				panic(fmt.Sprintf("invalid comparison op: %s", tc.op))
			}
		})
	}
}
