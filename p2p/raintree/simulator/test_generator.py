import argparse
import sys

import stringcase
from num2words import num2words
from pptree import Node, print_tree

from simulator import Counters, print_results, simulate

TEST_FORMAT = """
func TestRainTreeComplete{0}Nodes(t *testing.T) {{
    originatorNode := validatorId({1})
	var expectedCalls = TestNetworkSimulationConfig{{
		{2}
	}}
	testRainTreeCalls(t, originatorNode, expectedCalls, false)
}}
"""


def prepare_test(
    orig_addr: str,
    counters: Counters,
    num_nodes: int,
    root_node: Node,
    filename: str,
) -> None:
    test_generator = {}
    test = ""

    for k, _ in counters.msgs_rec_map.items():
        test_generator[k] = (
            counters.msgs_rec_map[k],
            counters.msgs_sent_map[k],
        )

    for i in range(num_nodes):
        k = f"val_{i+1}"
        if k == orig_addr:
            test += f"originatorNode:"
        else:
            test += f"		validatorId({i+1}):"
        read = counters.msgs_rec_map[k]
        write = counters.msgs_sent_map[k]
        test += f"  {{{read+1}, {write}}},\n"

    # print(test_generator)
    num_nodes_str = stringcase.camelcase(num2words(i).replace("-", "_")).capitalize()
    go_test = TEST_FORMAT.format(num_nodes_str, i, test)
    with open(filename, "w") as sys.stdout:
        print_tree(root_node, horizontal=False)
        print(go_test)


def main(args):
    # Simulation Parameters
    num_simulations = 1
    target1_per = 1 / 3
    target2_per = 2 / 3
    shrinkage_factor = 2 / 3
    num_nodes = args.num_nodes
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
    prepare_test(orig_addr, counters, num_nodes, root_node, args.output_file)


if __name__ == "__main__":

    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--num_nodes",
        dest="num_nodes",
        type=int,
        default=12,
        help="Number of nodes to simulate in RainTree",
    )
    parser.add_argument(
        "--output_file",
        dest="output_file",
        type=str,
        default="raintree_single_test.go",
        help="File where the go test should be written to",
    )
    args = parser.parse_args()

    main(args)
