// Package protocols contains the adding/removing protocol which permits to change the encryption of data.
// It allows to remove/add a server contribution to the encryption of ciphertexts.
// We assume that the server joining/leaving the cothority participates in the process.
package protocols

import (
	"github.com/JLRgithub/PrivateDCi2b2/lib"
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/log"
	"gopkg.in/dedis/onet.v1/network"
)

// AddRmServerProtocolName is the registered name for the local aggregation protocol.
const AddRmServerProtocolName = "AddRmServer"

func init() {
	onet.GlobalProtocolRegister(AddRmServerProtocolName, NewAddRmProtocol)
}

// Protocol
//______________________________________________________________________________________________________________________

// AddRmServerProtocol is a struct holding the state of a protocol instance.
type AddRmServerProtocol struct {
	*onet.TreeNodeInstance

	// Protocol feedback channel
	FeedbackChannel chan []lib.DpResponse

	// Protocol state data
	TargetOfTransformation []lib.DpResponse
	KeyToRm                abstract.Scalar
	Proofs                 bool
	Add                    bool
}

// NewAddRmProtocol is constructor of add/rm protocol instances.
func NewAddRmProtocol(n *onet.TreeNodeInstance) (onet.ProtocolInstance, error) {
	pvp := &AddRmServerProtocol{
		TreeNodeInstance: n,
		FeedbackChannel:  make(chan []lib.DpResponse),
	}

	return pvp, nil
}

var finalResultAddrm = make(chan []lib.DpResponse)

// Start is called at the root to start the execution of the Add/Rm protocol.
func (p *AddRmServerProtocol) Start() error {

	log.Lvl1(p.Name(), "starts a server adding/removing Protocol")
	roundComput := lib.StartTimer(p.Name() + "_AddRmServer(PROTOCOL)")

	result := make([]lib.DpResponse, len(p.TargetOfTransformation))

	wg := lib.StartParallelize(len(p.TargetOfTransformation))
	for i, v := range p.TargetOfTransformation {
		if lib.PARALLELIZE {
			go func(i int, v lib.DpResponse) {
				defer wg.Done()
				result[i] = changeEncryption(v, p.KeyToRm, p.Add)
			}(i, v)
		} else {
			result[i] = changeEncryption(v, p.KeyToRm, p.Add)
		}

	}

	lib.EndParallelize(wg)
	lib.EndTimer(roundComput)

	roundProof := lib.StartTimer(p.Name() + "_AddRmServer(PROOFS)")
	pubs := make([]lib.PublishedAddRmProof, 0)
	if p.Proofs {
		wg := lib.StartParallelize(len(result))
		for i, v := range result {
			if lib.PARALLELIZE {
				go func(i int, v lib.DpResponse) {
					defer wg.Done()
					proofsCreation(pubs, p.TargetOfTransformation[i], v, p.KeyToRm, p.Add)
				}(i, v)

			} else {
				proofsCreation(pubs, p.TargetOfTransformation[i], v, p.KeyToRm, p.Add)
			}

		}
		lib.EndParallelize(wg)
	}

	lib.EndTimer(roundProof)

	roundProof = lib.StartTimer(p.Name() + "_AddRmServer(PROOFSVerif)")
	wg = lib.StartParallelize(len(pubs))
	for _, v := range pubs {
		if lib.PARALLELIZE {
			go func(v lib.PublishedAddRmProof) {
				defer wg.Done()
				lib.PublishedAddRmCheckProof(v)
			}(v)
		} else {
			lib.PublishedAddRmCheckProof(v)
		}

	}
	lib.EndParallelize(wg)
	lib.EndTimer(roundProof)

	finalResultAddrm <- result
	return nil
}

// Dispatch is called on each node. It waits for incoming messages and handle them.
func (p *AddRmServerProtocol) Dispatch() error {
	aux := <-finalResultAddrm
	p.FeedbackChannel <- aux
	return nil
}

func changeEncryptionKeyMapCipherTexts(cv map[string]lib.CipherText, serverAddRmKey abstract.Scalar, toAdd bool) map[string]lib.CipherText {
	result := make(map[string]lib.CipherText, len(cv))
	for j, w := range cv {
		tmp := network.Suite.Point().Mul(w.K, serverAddRmKey)
		copy := result[j]
		copy.K = w.K
		if toAdd {
			copy.C = network.Suite.Point().Add(w.C, tmp)

		} else {
			copy.C = network.Suite.Point().Sub(w.C, tmp)
		}
		result[j] = copy
	}
	return result
}

func changeEncryption(response lib.DpResponse, keyToRm abstract.Scalar, add bool) lib.DpResponse {
	result := lib.DpResponse{}

	result.GroupByClear = response.GroupByClear
	result.GroupByEnc = changeEncryptionKeyMapCipherTexts(response.GroupByEnc, keyToRm, add)
	result.WhereClear = response.WhereClear
	result.WhereEnc = changeEncryptionKeyMapCipherTexts(response.WhereEnc, keyToRm, add)
	result.AggregatingAttributesEnc = changeEncryptionKeyMapCipherTexts(response.AggregatingAttributesEnc, keyToRm, add)
	result.AggregatingAttributesClear = response.AggregatingAttributesClear
	return result
}

func proofsCreation(pubs []lib.PublishedAddRmProof, target, v lib.DpResponse, keyToRm abstract.Scalar, add bool) {
	targetAggregatingAttributesEnc := target.AggregatingAttributesEnc
	targetGroupingAttributes := target.GroupByEnc
	targetWhereAttributes := target.WhereEnc

	prfAggr := lib.VectorAddRmProofCreation(targetAggregatingAttributesEnc, v.AggregatingAttributesEnc, keyToRm, add)
	prfGrp := lib.VectorAddRmProofCreation(targetGroupingAttributes, v.GroupByEnc, keyToRm, add)
	prfWhere := lib.VectorAddRmProofCreation(targetWhereAttributes, v.WhereEnc, keyToRm, add)
	ktopub := network.Suite.Point().Mul(network.Suite.Point().Base(), keyToRm)
	pub1 := lib.PublishedAddRmProof{Arp: prfAggr, VectBefore: targetAggregatingAttributesEnc, VectAfter: v.AggregatingAttributesEnc, Krm: ktopub, ToAdd: add}
	pub2 := lib.PublishedAddRmProof{Arp: prfGrp, VectBefore: v.GroupByEnc, VectAfter: v.GroupByEnc, Krm: ktopub, ToAdd: add}
	pub3 := lib.PublishedAddRmProof{Arp: prfWhere, VectBefore: v.WhereEnc, VectAfter: v.WhereEnc, Krm: ktopub, ToAdd: add}

	pubs = append(pubs, pub1, pub2, pub3)
}
