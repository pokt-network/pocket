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
    depth_reached_map: defaultdict[str, int]  # Max depth reached by node addr
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
    addr: str  # The addr of the node propagating the message
    addr_book: List[str]  # the addr's current addr book
    depth: int  # the current depth in the propagations
    t1: float  # % of location in addr book for target 1
    t2: float  # % of location in addr book for target 1
    shrinkage: float  # addr book shrinkage coefficient
    node: Node  # current node
    sender: str  # sender addr (who sent the message to addr)

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
    addr, addr_book, depth, t1_per, t2_per, s_per, node, sender = p

    # Return if the addr book is empty
    if len(addr_book) == 0:
        return

    # Not a demote - real message over a network
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
    t1 = (i + int(n * t1_per)) % n
    t2 = (i + int(n * t2_per)) % n
    s = (i + int(n * s_per)) % n

    t1_addr = addr_book[t1]
    t2_addr = addr_book[t2]

    if t1_addr == t2_addr:
        t2_addr = None
    if t1_addr == addr:
        t1_addr = None

    def send(t: int, t_addr: str) -> None:
        counters.msgs_sent += 1
        t_s = (t + int(n * s_per)) % n
        t_book_s = shrink_list(addr_book.copy(), t, t_s)
        queue.append(
            (
                PropagationQueueElement(
                    t_addr,
                    t_book_s,
                    depth + 1,
                    t1_per,
                    t2_per,
                    s_per,
                    Node(t_addr, node),
                    addr,
                ),
                counters,
                queue,
            )
        )

        counters.nodes_missing.discard(t_addr)
        counters.nodes_reached.add(t_addr)
        counters.msgs_sent_map[addr] += 1
        print(f"Msg: {format_send_message(addr_book, i, t)}")

    # Send a message to the first target
    if t1_addr is not None:
        send(t1, t1_addr)

    # Send a message to the second target
    if t2_addr is not None:
        send(t2, t2_addr)

    # Demote - not incrementing `msg_send_counter` since it's not a send
    addr_book_s = shrink_list(addr_book, i, s)
    if len(addr_book_s) > 1:
        queue.append(
            (
                PropagationQueueElement(
                    addr,
                    addr_book_s,
                    depth + 1,
                    t1_per,
                    t2_per,
                    s_per,
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
    max_allowed_depth = math.log(
        num_nodes, 3
    )  # 3 comes from the fact that we use a ternary tree
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
