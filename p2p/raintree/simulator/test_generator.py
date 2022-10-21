from simulator import Counters, print_results, simulate

TEST_FORMAT = """
func TestRainTreeComplete{}Nodes(t *testing.T) {
	originatorNode := validatorId(t, 1)
	var expectedCalls = TestRainTreeCommConfig{
		validatorId(t, 1): {0, 0}, // Originator
		validatorId(t, 2): {1, 1},
		validatorId(t, 3): {1, 1},
		validatorId(t, 4): {1, 1},
		validatorId(t, 5): {1, 1},
		validatorId(t, 6): {1, 1},
		validatorId(t, 7): {1, 1},
		validatorId(t, 8): {1, 1},
		validatorId(t, 9): {1, 1},
	}
	testRainTreeCalls(t, originatorNode, expectedCalls, false)
}
"""


def prepare_test(orig_addr: str, counters: Counters) -> str:
    test_generator = {}
    test = ""

    for k, _ in counters.msgs_rec_map.items():
        test_generator[k] = (
            counters.msgs_rec_map[k],
            counters.msgs_sent_map[k],
        )

    for i in range(num_nodes):
        k = f"val_{i+1}"
        read = counters.msgs_rec_map[k]
        write = counters.msgs_sent_map[k]
        test += f"validatorId({i+1}):  {{{read+1}, {write}}},\n"

    print(test_generator)
    return test


# Simulation Parameters
num_simulations = 1
target1_per = 1 / 3
target2_per = 2 / 3
shrinkage_factor = 2 / 3
num_nodes = 27
addr_book = sorted(
    [f"val_{i+1}" for i in range(num_nodes)], key=lambda x: int(x.split("_")[1])
)
orig_addr = addr_book[0]

# Run Simulation
root_node, counters = simulate(
    orig_addr, addr_book, target1_per, target2_per, shrinkage_factor
)

# Print Results
print_results(
    root_node, counters, target1_per, target2_per, shrinkage_factor, num_nodes
)

# Prepare Test
print(prepare_test(orig_addr, counters))
