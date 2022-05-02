package leader_election

import (
	"fmt"
	"log"
	"pocket/consensus/leader_election/sortition"
	"pocket/consensus/leader_election/vrf"
	consensuconsensus "pocket/consensus/types"
)

/*
References:
- https://github.com/algorand/go-algorand/tree/041e1f92d9c190bdc6d6c78b1dd04ef19b8ec03b/data/committee/sortition
- https://community.algorand.org/blog/the-intuition-behind-algorand-cryptographic-sortition/
*/

type LeaderCandidate struct {
	nodeId          consensuconsensus.NodeId
	verificationKey *vrf.VerificationKey
	vrfProof        vrf.VRFProof
	vrfOut          vrf.VRFOutput
	sortitionResult sortition.SortitionResult
	// TODO: SHould we include `height`, `round` and `prevBlockHash` here?
}

/*
	VERY IMPORTANT

	The verificationKey of each validator MUST be entered into the system BEFORE the block at height h.
	The security of the sortition lies in the fact that `prevBlockHash` is unknown and only determined
	after the VRF keys are refreshed across the network.

	Reference: https://medium.com/algorand/algorand-releases-first-open-source-code-of-verifiable-random-function-93c2960abd61
*/
func IsLeaderCandidate(
	validator *consensuconsensus.Validator,
	height consensuconsensus.BlockHeight,
	round consensuconsensus.Round,
	prevBlockHash string,
	votingPower float64,
	totalVotingPower float64,
	numCandidatesLeadersPerRound float64,
	vrfSecretKey *vrf.SecretKey,
) (*LeaderCandidate, error) {
	seed := sortition.FormatSeed(height, round, prevBlockHash)

	vrfOut, vrfProof, err := vrfSecretKey.Prove(seed)
	if err != nil {
		return nil, err
	}

	sortitionResult := sortition.Sortition(votingPower, totalVotingPower, numCandidatesLeadersPerRound, vrfOut)
	if sortitionResult == 0 {
		return nil, nil
	}

	verificationKey, err := vrfSecretKey.VerificationKey()
	if err != nil {
		return nil, err
	}

	return &LeaderCandidate{
		nodeId:          validator.NodeId,
		verificationKey: verificationKey,
		vrfProof:        vrfProof,
		vrfOut:          vrfOut,
		sortitionResult: sortitionResult,
	}, nil
}

func ElectLeader(
	leaderCandidates []*LeaderCandidate,
	height consensuconsensus.BlockHeight,
	round consensuconsensus.Round,
	prevBlockHash string,
) (consensuconsensus.NodeId, error) {
	seed := sortition.FormatSeed(height, round, prevBlockHash)

	var leaderCandidate *LeaderCandidate = nil
	for _, candidate := range leaderCandidates {
		if candidate.sortitionResult == 0 {
			continue
		}

		verification, err := candidate.verificationKey.Verify(seed, candidate.vrfProof, candidate.vrfOut)
		if err != nil || !verification {
			log.Printf("[WARN] Candidate leader failed verification: NodeId: %d; Error: %v", candidate.nodeId, err)
			continue
		}

		// TODO(Discuss): Should we be using `vrfOutProb(p1)` or `sortitionResult`` to break ties?
		// if highProof == nil || vrfOutProb(p1) >= vrfOutProb(p2) {
		if leaderCandidate == nil || candidate.sortitionResult >= leaderCandidate.sortitionResult {
			leaderCandidate = candidate
		}
	}

	if leaderCandidate == nil {
		return 0, fmt.Errorf("leader could not be selected")
	}
	return leaderCandidate.nodeId, nil
}
