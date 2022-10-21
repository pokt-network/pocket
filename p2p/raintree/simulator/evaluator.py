from collections import defaultdict
from typing import List

from . import simulate


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

    orig_addr = addr_book[0]
    for _ in range(num_simulations):
        simulate(orig_addr, addr_book, t1, t2, shrinkage)

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
