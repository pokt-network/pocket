from collections import defaultdict
from collections import deque
from pptree import Node, print_tree

import math
import warnings

warnings.filterwarnings("ignore", category=DeprecationWarning)

# Helpers
def shrink_list(l, i1, i2):
    if i1 <= i2:
        return l[i1:i2]
    return l[i1:] + l[:i2]


def get_msg_sent(l, i, t):
    s = "[ "
    for idx, n in enumerate(l):
        if idx == i:
            s += f"({n}), "
        elif idx == t:
            s += f"**{n}**, "
        else:
            s += f"{n}, "
    return f"{s[:-2]} ]"


def agg_dicts(d1, d2):
    return {k: d1.get(k, 0) + d2.get(k, 0) for k in set(d1) | set(d2)}


def run_simulations(
    S, N, X, Y, Z, should_print
):  # S = num simulations, N = num nodes, X = 1st message, Y = 2nd message, Z = shrinkage
    # Reset global params for simulation
    global global_msg_send_counter
    global global_set_nodes_reached
    global global_map_msg_rec_counter
    global global_map_msg_send_counter
    global global_map_depth_reached_counter
    global global_prop_queue
    global global_missing_nodes

    # Flags are used for visualization (not proving max depth)
    global global_enforce_full_prop
    global global_max_allowed_depth

    # Core logic
    def prop(
        addr, book, depth, X, Y, Z, node, sender
    ):  # addr => curr_add, book => curr_addr_book, depth => curr_depth
        global global_msg_send_counter
        global global_set_nodes_reached
        global global_map_msg_rec_counter
        global global_map_msg_send_counter
        global global_map_depth_reached_counter
        global global_prop_queue
        global global_missing_nodes

        # Flags are used for visualization (not proving max depth)
        global global_enforce_full_prop
        global global_max_allowed_depth

        if len(book) == 0:
            return

        # if depth >= global_max_allowed_depth:
        #     global_map_depth_reached_counter[depth] += 1
        #     print("Max depth reached")
        #     return

        # Add node to the tree
        # node = Node(addr) if parent is None else Node(addr, parent)

        if len(global_missing_nodes) == 0:
            global_map_depth_reached_counter[depth] += 1
            if not global_enforce_full_prop or depth >= global_max_allowed_depth:
                return

        # A network message was made - track it
        global_missing_nodes.discard(addr)
        global_set_nodes_reached.add(addr)
        if addr != sender:
            global_map_msg_rec_counter[addr] += 1

        if should_print:
            print("-----")  # Separator to make reading easier

        n = len(book)
        i = book.index(addr)
        # x = (i + math.ceil(n * X)) % n
        # y = (i + math.ceil(n * Y)) % n
        # z = (i + math.ceil(n * Z)) % n
        x = (i + int(n * X)) % n
        y = (i + int(n * Y)) % n
        z = (i + int(n * Z)) % n

        x_addr = book[x]
        y_addr = book[y]

        if x_addr == y_addr:
            y_addr = None
        if x_addr == addr:
            x_addr = None

        # print(f"Global missing nodes ({len(global_missing_nodes)}): ", global_missing_nodes)

        # Send to first target
        if x_addr is not None:
            global_msg_send_counter += 1
            x_book = book.copy()
            x_z = (x + int(n * Z)) % n
            # x_z = (x  + math.ceil(n * Z)) % n
            x_book_s = shrink_list(x_book, x, x_z)
            global_prop_queue.append(
                (x_addr, x_book_s, depth + 1, X, Y, Z, Node(x_addr, node), addr)
            )

            # Assumes successfull propagation
            global_missing_nodes.discard(x_addr)
            global_set_nodes_reached.add(x_addr)

            global_map_msg_send_counter[addr] += 1
            if should_print:
                print(f"Msg 1: {get_msg_sent(book, i, x)}")

        # Send to second target
        if y_addr is not None:
            global_msg_send_counter += 1
            y_book = book.copy()
            y_z = (y + int(n * Z)) % n
            # y_z = (y  + math.ceil(n * Z)) % n
            y_book_s = shrink_list(y_book, y, y_z)
            global_prop_queue.append(
                (y_addr, y_book_s, depth + 1, X, Y, Z, Node(y_addr, node), addr)
            )
            global_map_msg_send_counter[addr] += 1
            if should_print:
                print(f"Msg 2: {get_msg_sent(book, i, y)}")

            # Assumes successfull propagation
            global_missing_nodes.discard(y_addr)
            global_set_nodes_reached.add(y_addr)

        # This is a demote (not a send) so we do not increment `global_msg_send_counter`
        book_s = shrink_list(book, i, z)
        if len(book_s) > 1:
            global_prop_queue.append(
                (addr, book_s, depth + 1, X, Y, Z, Node(addr, node), addr)
            )

        # return node

    ## Simulations counters
    msg_send_counter_acc = 0
    map_msg_rec_counter_acc = defaultdict(int)
    depth_counter_acc = defaultdict(int)

    # global_addr_book = sorted([chr(ord('A') + i) for i in range(N)])
    global_addr_book = sorted(
        [f"val_{i+1}" for i in range(N)], key=lambda x: int(x.split("_")[1])
    )

    # Flags are used for visualization (not proving max depth)
    global_enforce_full_prop = True
    global_max_allowed_depth = math.log(N, 3)
    print("!!!!!!!!", global_max_allowed_depth)

    for _ in range(S):
        # Reset global params for simulation
        global_msg_send_counter = 0
        global_set_nodes_reached = set()
        global_map_msg_rec_counter = defaultdict(int)
        global_map_msg_send_counter = defaultdict(int)
        global_map_depth_reached_counter = defaultdict(int)
        global_prop_queue = deque()
        global_missing_nodes = set(global_addr_book)

        # Start simulation
        orig_addr = "val_1"
        # orig_addr = 'O'
        # orig_addr = random.choice(global_addr_book)
        # orig_addr = global_addr_book[0]
        root = Node(orig_addr)
        global_prop_queue.append(
            (orig_addr, global_addr_book, 0, X, Y, Z, root, orig_addr)
        )
        # root = None
        while len(global_prop_queue) > 0:
            prop(*global_prop_queue.popleft())
            # node = prop(*global_prop_queue.popleft())
            # root = node if root is None else root

        # Print results
        if should_print:
            print("###################")
            print_tree(root, horizontal=False)
            print(f"Target Coverage: {Y}")
            print(f"Num nodes: {N}")
            print(f"Global Send Counter: {global_msg_send_counter}")
            print(f"Global Set Reached: {sorted(list(global_set_nodes_reached))}")
            print(
                f"Global # Times Received: {dict(dict(sorted(global_map_msg_rec_counter.items(), key=lambda item: -item[1])))}"
            )
            print(
                f"Nodes not reached: {global_set_nodes_reached.difference(global_addr_book)}"
            )

        # Aggregate results
        depth_counter_acc = agg_dicts(
            depth_counter_acc, global_map_depth_reached_counter
        )
        msg_send_counter_acc += global_msg_send_counter
        map_msg_rec_counter_acc = agg_dicts(
            map_msg_rec_counter_acc, global_map_msg_rec_counter
        )

    msg_send_counter_acc /= S
    depth_counter_acc = {k: round(i / S, 3) for k, i in depth_counter_acc.items()}
    map_msg_rec_counter_acc = {
        k: round(i / S, 3) for k, i in map_msg_rec_counter_acc.items()
    }
    map_msg_rec_counter_acc = dict(
        sorted(map_msg_rec_counter_acc.items(), key=lambda item: -item[1])
    )

    return (msg_send_counter_acc, depth_counter_acc, map_msg_rec_counter_acc)


# Algo Params
S = 1000

# N = 18 # Num nodes
# N = 27 # Num nodes
N = 4  # Num nodes
X = 1 / 3  # 1st message
Y = 2 / 3  # 2nd Message
Z = 2 / 3  # Shrinkage

(msg_count, depth_acc, msg_acc) = run_simulations(S, N, X, Y, Z, True)
test_generator = {}

for k, v in global_map_msg_rec_counter.items():
    test_generator[k] = (global_map_msg_rec_counter[k], global_map_msg_send_counter[k])
for i in range(N):
    k = f"val_{i+1}"
    read = global_map_msg_rec_counter[k]
    write = global_map_msg_send_counter[k]
    print(f"validatorId({i+1}):  {{{read+1}, {write}}},")
# print(test_generator)
# print(global_map_msg_rec_counter)
# print(global_map_msg_send_counter)

print("Simulation results")
print((msg_count, depth_acc, msg_acc))
