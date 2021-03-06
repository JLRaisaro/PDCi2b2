package serviceI2B2dc

import (
	"time"

	"github.com/JLRgithub/PrivateDCi2b2/lib"
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/crypto.v0/config"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/log"
	"gopkg.in/dedis/onet.v1/network"
)

// API represents a client with the server to which he is connected and its public/private key pair.
type API struct {
	*onet.Client
	clientID   string
	entryPoint *network.ServerIdentity
	public     abstract.Point
	private    abstract.Scalar
}

// NewClient constructor of a client.
func NewClient(entryPoint *network.ServerIdentity, clientID string) *API {
	keys := config.NewKeyPair(network.Suite)

	newClient := &API{
		Client:     onet.NewClient(ServiceName),
		clientID:   clientID,
		entryPoint: entryPoint,
		public:     keys.Public,
		private:    keys.Secret,
	}
	return newClient
}

// Send Query
//______________________________________________________________________________________________________________________

// SendQuery creates a query based on a set of entities (servers) and a query description.
func (c *API) SendQuery(entities *onet.Roster, queryID QueryID, clientPubKey abstract.Point, locations, time, concepts, groupBy []string) (*QueryID, error) {
	log.Lvl1(c, " creates a query with input: location=", locations, "time=", time, "concept=", concepts, "groupBy=", groupBy)

	var newQueryID QueryID

	cq := CreationQueryDC{
		QueryID:      queryID,
		Roster:       *entities,
		ClientPubKey: clientPubKey,

		// query statement
		Locations: locations,
		Times:     time,
		Concepts:  concepts,
		GroupBy:   groupBy,
	}
	resp := ServiceState{}
	err := c.SendProtobuf(c.entryPoint, &cq, &resp)
	if err != nil {
		return nil, err
	}
	log.Lvl1(c, " receives confirmation from server for query with ID: ", resp.QueryID)
	newQueryID = resp.QueryID

	return &newQueryID, nil
}

// SendResultsQuery to get the result from associated server and decrypt the response using its private key.
func (c *API) ExecuteQuery(queryID QueryID) (*[]string, *[]int64, error) {
	log.Lvl1(c, " asks the server to run the query with ID: ", queryID)
	resp := ServiceResult{}
	err := c.SendProtobuf(c.entryPoint, &ResultsQueryDC{false, queryID, c.public}, &resp)
	if err != nil {
		return nil, nil, err
	}

	log.Lvl1(c, " receives the query results from ", c.entryPoint)

	//grpClear := make([][]int64, len(resp.Results))
	start := time.Now()
	aggr := lib.DecryptIntVector(c.private, &((*resp.Results)[0].AggregatingAttributes))
	log.LLvl1("Decryption Time:", time.Since(start))
	groups := *resp.Groups

	return &groups, &aggr, nil
}

// String permits to have the string representation of a client.
func (c *API) String() string {
	return "[Client-" + c.clientID + "]"
}
