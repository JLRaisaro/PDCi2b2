package serviceI2B2dc_test

import (
	"strconv"
	"testing"

	"github.com/JLRgithub/PrivateI2B2DQ/services/i2b2dc"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/log"
)

// numberGrpAttr is the number of group attributes.
const numberGrpAttr = 3

// numberAttr is the number of attributes.
const numberAttr = 2

const proofsService = true

func TestMain(m *testing.M) {
	log.MainTest(m)
}

// TEST BATCH 1 -> encrypted or/and non-encrypted grouping attributes

func Test1(t *testing.T) {
	log.Lvl1("***************************************************************************************************")

	log.SetDebugVisible(2)
	local := onet.NewLocalTest()
	_, el, _ := local.GenTree(2, true)
	defer local.CloseAll()

	// Send a request to the service
	client := serviceI2B2dc.NewClient(el.List[0], strconv.Itoa(0))

	locations := []string{}
	times := []string{"2015"}
	concepts := []string{"ICD10:E08.52", "ICD10:E08.59"}
	groupBy := []string{"location_cd"}

	queryID, err := client.SendQuery(el, serviceI2B2dc.QueryID(""), nil, locations, times, concepts, groupBy)

	if err != nil {
		t.Fatal("Service did not start.", err)
	}

	log.Lvl1("Output query creation ", *queryID)

	grps, aggr, err := client.ExecuteQuery(*queryID)

	log.Lvl1("Output query results ", *grps, *aggr)

}
