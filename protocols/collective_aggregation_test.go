package protocols_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/JLRgithub/PrivateI2B2DQ/lib"
	"github.com/JLRgithub/PrivateI2B2DQ/protocols"
	"github.com/stretchr/testify/assert"
	"gopkg.in/dedis/crypto.v0/random"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/log"
	"gopkg.in/dedis/onet.v1/network"
)

var suite = network.Suite
var clientPrivate = suite.Scalar().Pick(random.Stream)
var clientPublic = suite.Point().Mul(suite.Point().Base(), clientPrivate)
var grpattr1 = lib.DeterministCipherText{Point: suite.Point().Base()}
var grpattr2 = lib.DeterministCipherText{Point: suite.Point().Null()}
var groupingAttrA = lib.DeterministCipherVector{grpattr1, grpattr1}
var groupingAttrB = lib.DeterministCipherVector{grpattr2, grpattr2}
var groupingAttrC = lib.DeterministCipherVector{grpattr1, grpattr2}

//TestCollectiveAggregation tests collective aggregation protocol
func TestCollectiveAggregation(t *testing.T) {
	local := onet.NewLocalTest()

	// You must register this protocol before creating the servers
	onet.GlobalProtocolRegister("CollectiveAggregationTest", NewCollectiveAggregationTest)
	_, _, tree := local.GenTree(10, true)
	defer local.CloseAll()

	p, err := local.CreateProtocol("CollectiveAggregationTest", tree)
	if err != nil {
		t.Fatal("Couldn't start protocol:", err)
	}
	protocol := p.(*protocols.CollectiveAggregationProtocol)

	//run protocol
	go protocol.Start()
	timeout := network.WaitRetry * time.Duration(network.MaxRetryConnect*5*2) * time.Millisecond

	feedback := protocol.FeedbackChannel

	//verify results
	expectedGroups := map[lib.GroupingKey][]int64{groupingAttrA.Key(): []int64{1, 1},
		groupingAttrB.Key(): []int64{1, 2},
		groupingAttrC.Key(): []int64{3, 3}}

	expectedResults := map[lib.GroupingKey][]int64{groupingAttrA.Key(): {3, 5, 7, 9, 11},
		groupingAttrB.Key(): {1, 2, 3, 4, 5},
		groupingAttrC.Key(): {1, 1, 1, 1, 1}}

	select {
	case encryptedResult := <-feedback:
		log.Lvl1("Recieved results:")
		resultData := make(map[lib.GroupingKey][]int64)
		for k, v := range encryptedResult.GroupedData {
			resultData[k] = lib.DecryptIntVector(clientPrivate, &v.AggregatingAttributes)

			log.Lvl1(k, resultData[k])
		}
		for k, v1 := range expectedGroups {
			if v2, ok := encryptedResult.GroupedData[k]; ok {
				assert.True(t, ok)
				_ = v1
				_ = v2
				assert.True(t, reflect.DeepEqual(v1, lib.DecryptIntVector(clientPrivate, &v2.GroupByEnc)))
				delete(encryptedResult.GroupedData, k)
			}
		}
		assert.Empty(t, encryptedResult.GroupedData)
		assert.Equal(t, expectedResults, resultData)
	case <-time.After(timeout):
		t.Fatal("Didn't finish in time")
	}
}

// NewCollectiveAggregationTest is a test specific protocol instance constructor that injects test data.
func NewCollectiveAggregationTest(tni *onet.TreeNodeInstance) (onet.ProtocolInstance, error) {

	pi, err := protocols.NewCollectiveAggregationProtocol(tni)
	protocol := pi.(*protocols.CollectiveAggregationProtocol)

	testCVMap := make(map[lib.GroupingKey]lib.FilteredResponse)

	switch tni.Index() {
	case 0:
		log.Lvl1("0")
		testCVMap[groupingAttrA.Key()] = lib.FilteredResponse{GroupByEnc: *lib.EncryptIntVector(clientPublic, []int64{1, 1}), AggregatingAttributes: *lib.EncryptIntVector(clientPublic, []int64{1, 2, 3, 4, 5})}
		testCVMap[groupingAttrB.Key()] = lib.FilteredResponse{GroupByEnc: *lib.EncryptIntVector(clientPublic, []int64{1, 2}), AggregatingAttributes: *lib.EncryptIntVector(clientPublic, []int64{0, 0, 0, 0, 0})}
	case 1:
		log.Lvl1("1")
		testCVMap[groupingAttrB.Key()] = lib.FilteredResponse{GroupByEnc: *lib.EncryptIntVector(clientPublic, []int64{1, 2}), AggregatingAttributes: *lib.EncryptIntVector(clientPublic, []int64{1, 2, 3, 4, 5})}
	case 2:
		log.Lvl1("2")
		testCVMap[groupingAttrA.Key()] = lib.FilteredResponse{GroupByEnc: *lib.EncryptIntVector(clientPublic, []int64{1, 1}), AggregatingAttributes: *lib.EncryptIntVector(clientPublic, []int64{1, 1, 1, 1, 1})}
	case 9:
		log.Lvl1("9")
		testCVMap[groupingAttrC.Key()] = lib.FilteredResponse{GroupByEnc: *lib.EncryptIntVector(clientPublic, []int64{3, 3}), AggregatingAttributes: *lib.EncryptIntVector(clientPublic, []int64{1, 0, 1, 0, 1})}
		testCVMap[groupingAttrA.Key()] = lib.FilteredResponse{GroupByEnc: *lib.EncryptIntVector(clientPublic, []int64{1, 1}), AggregatingAttributes: *lib.EncryptIntVector(clientPublic, []int64{1, 2, 3, 4, 5})}
	case 5:
		log.Lvl1("5")
		testCVMap[groupingAttrC.Key()] = lib.FilteredResponse{GroupByEnc: *lib.EncryptIntVector(clientPublic, []int64{3, 3}), AggregatingAttributes: *lib.EncryptIntVector(clientPublic, []int64{0, 1, 0, 1, 0})}

	default:
	}
	protocol.GroupedData = &testCVMap

	return protocol, err
}
