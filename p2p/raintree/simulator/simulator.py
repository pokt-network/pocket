import math
import warnings
from collections import defaultdict, deque
from dataclasses import dataclass
from typing import Dict, List

from pptree import Node, print_tree

warnings.filterwarnings("ignore", category=DeprecationWarning)

# ~~~ Helpers ~~~

# Returns a subset of the list with items from `i1` to `i2` in `l` wrapping around if needed
def shrink_list(l: List[str], i1: int, i2: int) -> List[str]:
    if i1 <= i2:
        return l[i1:i2]
    return l[i1:] + l[:i2]


# Formats a string showing who the sender and receivers were
def format_send_message(l: List[str], self_index: int, target: int) -> str:
    s = "[ "
    for idx, n in enumerate(l):
        if idx == self_index:
            s += f"({n}), "
        elif idx == target:
            s += f"**{n}**, "
        else:
            s += f"{n}, "
    return f"{s[:-2]} ]"


# Sum the values from two dictionaries where the keys overlap
def agg_dicts(d1: Dict[str, int], d2: Dict[str, int]) -> Dict[str, int]:
    return {k: d1.get(k, 0) + d2.get(k, 0) for k in set(d1) | set(d2)}


# ~~~ Data Types ~~~


@dataclass
class Counters:
    msgs_sent: int  # Total num of messages sent by RainTree propagating
    nodes_reached: set[str]  # Nodes reached by current RainTree propagating
    nodes_missing: set[str]  # Nodes not yet reached by current RainTree propagating
    msgs_rec_map: defaultdict[str, int]  # Num messages received by node addr
    msgs_sent_map: defaultdict[str, int]  # Num messages sent by node addr
    depth_reached_map: defaultdict[str, int]  # Addr -> depth reached by node addr
    max_theoretical_depth: int  # Theoretical max depth, used to end propagating early

    def __init__(self, nodes: List[str], max_allowed_depth: int):
        self.msgs_sent = 0
        self.nodes_reached = set()
        self.nodes_missing = set(nodes)
        self.msgs_rec_map = defaultdict(int)
        self.msgs_sent_map = defaultdict(int)
        self.depth_reached_map = defaultdict(int)
        self.max_theoretical_depth = max_allowed_depth


@dataclass
class PropagationQueueElement:
    addr: str
    addr_book: List[str]
    depth: int  # current depth
    t1: float  # target 1 percentage factor
    t2: float  # target 2 percentage factor
    shrinkage: float  # addr book shrinkage coefficient
    node: Node  # current node
    sender: str  # sender addr

    def __iter__(self):
        return iter(
            (
                self.addr,
                self.addr_book,
                self.depth,
                self.t1,
                self.t2,
                self.shrinkage,
                self.node,
                self.sender,
            )
        )


# A single RainTree propagation step
def propagate(
    p: PropagationQueueElement,
    counters: Counters,
    queue: deque[PropagationQueueElement],
) -> None:
    addr, addr_book, depth, t1, t2, s, node, sender = p

    # Return if the addr boo is empty
    if len(addr_book) == 0:
        return

    # If the theoretical depth was reached and no nodes are missing, return
    if len(counters.nodes_missing) == 0:
        counters.depth_reached_map[depth] += 1
        if depth >= counters.max_theoretical_depth:
            return

    # A network message was sent
    counters.nodes_missing.discard(addr)
    counters.nodes_reached.add(addr)
    if addr != sender:
        counters.msgs_rec_map[addr] += 1

    # Configure who the current node should send messages to
    n = len(addr_book)
    i = addr_book.index(addr)
    x = (i + int(n * t1)) % n
    y = (i + int(n * t2)) % n
    z = (i + int(n * s)) % n

    x_addr = addr_book[x]
    y_addr = addr_book[y]

    if x_addr == y_addr:
        y_addr = None
    if x_addr == addr:
        x_addr = None

    # Send a message to the first target
    if x_addr is not None:
        counters.msgs_sent += 1
        x_z = (x + int(n * s)) % n
        x_book_s = shrink_list(addr_book.copy(), x, x_z)
        queue.append(
            (
                PropagationQueueElement(
                    x_addr,
                    x_book_s,
                    depth + 1,
                    t1,
                    t2,
                    s,
                    Node(x_addr, node),
                    addr,
                ),
                counters,
                queue,
            ),
        )

        counters.nodes_missing.discard(x_addr)
        counters.nodes_reached.add(x_addr)
        counters.msgs_sent_map[addr] += 1
        print(f"Msg 1: {format_send_message(addr_book, i, x)}")

    # Send a message to the second target
    if y_addr is not None:
        counters.msgs_sent += 1
        y_z = (y + int(n * s)) % n
        y_book_s = shrink_list(addr_book.copy(), y, y_z)
        queue.append(
            (
                PropagationQueueElement(
                    y_addr,
                    y_book_s,
                    depth + 1,
                    t1,
                    t2,
                    s,
                    Node(y_addr, node),
                    addr,
                ),
                counters,
                queue,
            )
        )

        counters.nodes_missing.discard(y_addr)
        counters.nodes_reached.add(y_addr)
        counters.msgs_sent_map[addr] += 1
        print(f"Msg 2: {format_send_message(addr_book, i, y)}")

    # Demote - not incrementing `msg_send_counter` since it's not a send
    addr_book_s = shrink_list(addr_book, i, z)
    if len(addr_book_s) > 1:
        queue.append(
            (
                PropagationQueueElement(
                    addr,
                    addr_book_s,
                    depth + 1,
                    t1,
                    t2,
                    s,
                    Node(addr, node),
                    addr,
                ),
                counters,
                queue,
            )
        )


# A single RainTree Simulation
def simulate(
    addr_book: List[str],
    t1: float,
    t2: float,
    shrinkage: float,
) -> Counters:
    num_nodes = len(addr_book)

    # Configure Simulation
    prop_queue = deque()
    max_allowed_depth = math.log(num_nodes, 3)
    counters = Counters(addr_book, max_allowed_depth)

    # Prepare Simulation
    orig_addr = "val_1"
    root_node = Node(orig_addr)
    prop_queue.append(
        (
            PropagationQueueElement(
                orig_addr,
                addr_book,
                0,
                t1,
                t2,
                shrinkage,
                root_node,
                orig_addr,
            ),
            counters,
            prop_queue,
        )
    )

    # Run Simulation to completion
    while len(prop_queue) > 0:
        propagate(*prop_queue.popleft())

    print("\n###################\n")
    print_tree(root_node, horizontal=False)
    print("\n###################\n")
    print(f"Coefficients used: t1: {t1:.3f}, t2: {t2:.3f}, shrinkage: {shrinkage:.3f}")
    print(f"Num messages sent: {counters.msgs_sent}")
    print(f"Num nodes reached: {len(counters.nodes_reached)}/ {num_nodes}")
    print(
        f"Messages received: {dict(dict(sorted(counters.msgs_rec_map.items(), key=lambda item: -item[1])))}"
    )
    print(
        f"Messages sent: {dict(dict(sorted(counters.msgs_sent_map.items(), key=lambda item: -item[1])))}"
    )

    return counters


# A series of RainTree simulations to collect data
def evaluate(
    num_simulations: int,
    addr_book: List[str],
    t1: float,
    t2: float,
    shrinkage: float,
):
    ## Evaluation accumulators
    msg_send_counter_acc = 0
    map_msg_rec_counter_acc = defaultdict(int)
    depth_counter_acc = defaultdict(int)

    for _ in range(num_simulations):
        simulate(addr_book, t1, t2, shrinkage)

        # Aggregate results
        # depth_counter_acc = agg_dicts(
        #     depth_counter_acc, global_map_depth_reached_counter
        # )
        # msg_send_counter_acc += global_msg_send_counter
        # map_msg_rec_counter_acc = agg_dicts(
        #     map_msg_rec_counter_acc, global_map_msg_rec_counter
        # )

    # msg_send_counter_acc /= S
    # depth_counter_acc = {k: round(i / S, 3) for k, i in depth_counter_acc.items()}
    # map_msg_rec_counter_acc = {
    #     k: round(i / S, 3) for k, i in map_msg_rec_counter_acc.items()
    # }
    # map_msg_rec_counter_acc = dict(
    #     sorted(map_msg_rec_counter_acc.items(), key=lambda item: -item[1])
    # )

    return (msg_send_counter_acc, depth_counter_acc, map_msg_rec_counter_acc)


# Algo Params
num_simulations = 1
target1_per = 1 / 3
target2_per = 2 / 3
shrinkage_factor = 2 / 3
num_nodes = 27
addr_book = sorted(
    [f"val_{i+1}" for i in range(num_nodes)], key=lambda x: int(x.split("_")[1])
)

# (msg_count, depth_acc, msg_acc) =
counters = simulate(addr_book, target1_per, target2_per, shrinkage_factor)

test_generator = {}

# for k, v in global_map_msg_rec_counter.items():
#     test_generator[k] = (global_map_msg_rec_counter[k], global_map_msg_send_counter[k])
# for i in range(num_nodes):
#     k = f"val_{i+1}"
#     read = global_map_msg_rec_counter[k]
#     write = global_map_msg_send_counter[k]
#     print(f"validatorId({i+1}):  {{{read+1}, {write}}},")
# print(test_generator)
# print(global_map_msg_rec_counter)
# print(global_map_msg_send_counter)
# print((msg_count, depth_acc, msg_acc))
