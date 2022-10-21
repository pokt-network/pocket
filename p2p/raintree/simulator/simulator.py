import math
import warnings
from collections import defaultdict, deque
from dataclasses import dataclass
from typing import Dict, List, Tuple

# TODO(olshansky): Consider investigating this library as well since it has custom typing: https://github.com/liwt31/print_tree
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

    # Theoretical max depth, used to end propagating early
    max_theoretical_depth: int

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

    if addr != sender:
        counters.msgs_rec_map[addr] += 1

    # If the theoretical depth was reached and no nodes are missing, return
    if len(counters.nodes_missing) == 0:
        counters.depth_reached_map[depth] += 1
        if depth >= counters.max_theoretical_depth:
            return

    # A network message was sent
    counters.nodes_missing.discard(addr)
    counters.nodes_reached.add(addr)

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
    orig_addr: str,
    addr_book: List[str],
    t1: float,
    t2: float,
    shrinkage: float,
) -> Tuple[Node, Counters]:
    num_nodes = len(addr_book)

    # Configure Simulation
    prop_queue = deque()
    max_allowed_depth = math.log(num_nodes, 3)
    counters = Counters(addr_book, max_allowed_depth)

    # Prepare Simulation
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

    return root_node, counters


def print_results(
    node: Node,
    counters: Counters,
    t1: float,
    t2: float,
    shrinkage: float,
    num_nodes: int,
) -> None:
    print("\n###################\n")
    print_tree(node, horizontal=False)
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
