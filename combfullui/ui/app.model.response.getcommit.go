package main

type CommitedUtxoTag struct {
	Commit string `json:"Commit"`

	UtxoTag
}
type UtxoTag struct {
	Combbase bool `json:"Combbase"`

	Height uint32 `json:"Height"`
	Number uint32 `json:"CommitNum"`
	Tx     uint16 `json:"TxNum"`
	Out    uint16 `json:"OutNum"`
}

type AppModelResponseGetCommitments struct {
	Commitments []CommitedUtxoTag `json:"Commitments"`
	Bloom       []byte            `json:"Bloom"`
	Count       uint64            `json:"Count"`
	Height      uint64            `json:"Height"`
}

const strictly_monotonic_vouts_bugfix_fork_height = 620000

func utag_cmp(l *UtxoTag, r *UtxoTag) int {
	if l.Height != r.Height {
		return int(l.Height) - int(r.Height)
	}

	// at this point the l and r heights are the same, use Natasha's fork
	// that is at the fork height we start to compare by commitnum
	if l.Height >= strictly_monotonic_vouts_bugfix_fork_height {
		return int(l.Number) - int(r.Number)
	}

	if l.Tx != r.Tx {
		return int(l.Tx) - int(r.Tx)
	}
	return int(l.Out) - int(r.Out)
}
