package serviceI2B2dc

import (
	"database/sql"
	"strings"
	"time"
	"github.com/BurntSushi/toml"
	"github.com/JLRgithub/PrivateDCi2b2/lib"
	"github.com/JLRgithub/PrivateDCi2b2/protocols"
	"github.com/btcsuite/goleveldb/leveldb/errors"
	_ "github.com/lib/pq"
	"github.com/satori/go.uuid"
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/log"
	"gopkg.in/dedis/onet.v1/network"
)

// ServiceName is the registered name for the unlynx service.
const ServiceName = "i2b2dc"

// QueryID unique ID for each query.
type QueryID string

// CreationQuery is used to trigger the creation of a query
type CreationQueryDC struct {
	QueryID      QueryID
	Roster       onet.Roster
	ClientPubKey abstract.Point

	// query statement
	Locations []string
	Times     []string
	Concepts  []string
	GroupBy   []string
}

// MsgTypes defines the Message Type ID for all the service's intra-messages.
type MsgTypes struct {
	msgCreationQueryDC network.MessageTypeID
	msgResultsQueryDC  network.MessageTypeID
}

// ResultsQueryDC is used by querier to ask for the response of the data characterization query.
type ResultsQueryDC struct {
	IntraMessage bool
	QueryID      QueryID
	ClientPublic abstract.Point
}

// ServiceState represents the service "state".
type ServiceState struct {
	QueryID QueryID
}

// DatabaseConfig represents the configuration of the database
type DatabaseConfig struct {
	Username string
	Password string
	DbName   string
	Table    string
}

// ServiceResult will contain final results of a query and be sent to querier.
type ServiceResult struct {
	Results *[]lib.FilteredResponse
	Groups  *[]string
}

//TODO: use concurrent map to deal with multiple clients
// Service defines a service in i2b2dc.
type Service struct {
	*onet.ServiceProcessor
	Query                        CreationQueryDC
	AggregatedResults            []lib.FilteredResponse
	KeySwitchedAggregatedResults []lib.FilteredResponse
	Groups                       []string
}

var msgTypes = MsgTypes{}
var dbConfig DatabaseConfig

func init() {
	onet.RegisterNewService(ServiceName, NewService)

	msgTypes.msgCreationQueryDC = network.RegisterMessage(&CreationQueryDC{})
	msgTypes.msgResultsQueryDC = network.RegisterMessage(&ResultsQueryDC{})

	network.RegisterMessage(&ServiceState{})
	network.RegisterMessage(&ServiceResult{})
}

// NewService constructor which registers the needed messages.
func NewService(c *onet.Context) onet.Service {
	newServiceInstance := &Service{
		ServiceProcessor: onet.NewServiceProcessor(c),
	}
	if cerr := newServiceInstance.RegisterHandler(newServiceInstance.HandleCreationQueryDC); cerr != nil {
		log.Fatal("Wrong Handler.", cerr)
	}
	if cerr := newServiceInstance.RegisterHandler(newServiceInstance.HandleResultsQueryDC); cerr != nil {
		log.Fatal("Wrong Handler.", cerr)
	}

	c.RegisterProcessor(newServiceInstance, msgTypes.msgCreationQueryDC)
	c.RegisterProcessor(newServiceInstance, msgTypes.msgResultsQueryDC)
	return newServiceInstance
}

// Process implements the processor interface and is used to recognize messages broadcasted between servers
func (s *Service) Process(msg *network.Envelope) {
	if msg.MsgType.Equal(msgTypes.msgCreationQueryDC) {
		tmp := (msg.Msg).(*CreationQueryDC)
		s.HandleCreationQueryDC(tmp)
	} else if msg.MsgType.Equal(msgTypes.msgResultsQueryDC) {
		tmp := (msg.Msg).(*ResultsQueryDC)
		s.HandleResultsQueryDC(tmp)
	}
}

// Query Handlers
//______________________________________________________________________________________________________________________

// HandleSurveyCreationQuery handles the reception of a survey creation query by instantiating the corresponding survey.
func (s *Service) HandleCreationQueryDC(recq *CreationQueryDC) (network.Message, onet.ClientError) {
	log.Lvl1(s.ServerIdentity().String(), " receives a Query Creation Request")

	// if this server is the one receiving the query from the client
	if recq.QueryID == "" {
		newID := QueryID(uuid.NewV4().String())
		recq.QueryID = newID
		//TODO: add checks on input to avoid SQL injections (regex)

		log.Lvl1(s.ServerIdentity().String(), " sends back confirmation to the client for query with ID: ", recq.QueryID)

	}

	// save the input query in the current service
	s.Query = *recq
	// initialize results containers
	s.AggregatedResults = make([]lib.FilteredResponse, 0, 0)
	s.KeySwitchedAggregatedResults = make([]lib.FilteredResponse, 0, 0)
	s.Groups = make([]string, 0, 0)

	return &ServiceState{s.Query.QueryID}, nil
}

// HandleSurveyResultsQuery handles the survey result query by the surveyor.
func (s *Service) HandleResultsQueryDC(resq *ResultsQueryDC) (network.Message, onet.ClientError) {

	log.Lvl1(s.ServerIdentity(), " receives an execution request for query with ID: ", resq.QueryID)

	s.Query.ClientPubKey = resq.ClientPublic

	s.StartService(resq.QueryID, true, false)

	log.Lvl1(s.ServerIdentity(), " sends result back to the client")

	return &ServiceResult{Results: &s.KeySwitchedAggregatedResults, Groups: &s.Groups}, nil

}

// Protocol Handlers
//______________________________________________________________________________________________________________________

// NewProtocol creates a protocol instance executed by all nodes
func (s *Service) NewProtocol(tn *onet.TreeNodeInstance, conf *onet.GenericConfig) (onet.ProtocolInstance, error) {
	tn.SetConfig(conf)

	var pi onet.ProtocolInstance
	var err error

	switch tn.ProtocolName() {
	case protocols.KeySwitchingProtocolName:
		pi, err = protocols.NewKeySwitchingProtocol(tn)
		if err != nil {
			return nil, err
		}

		keySwitch := pi.(*protocols.KeySwitchingProtocol)
		if tn.IsRoot() {

			keySwitch.TargetOfSwitch = &s.AggregatedResults
			keySwitch.TargetPublicKey = &s.Query.ClientPubKey
		}
	default:
		return nil, errors.New("Service attempts to start an unknown protocol: " + tn.ProtocolName() + ".")
	}

	return pi, nil
}

// StartProtocol starts a specific protocol (Pipeline, Shuffling, etc.)
func (s *Service) StartProtocol(name string) (onet.ProtocolInstance, error) {

	tree := s.Query.Roster.GenerateNaryTreeWithRoot(2, s.ServerIdentity())

	var tn *onet.TreeNodeInstance
	tn = s.NewTreeNodeInstance(tree, tree.Root, name)

	conf := onet.GenericConfig{Data: []byte(string(s.Query.QueryID))}

	pi, err := s.NewProtocol(tn, &conf)
	if err != nil {
		log.Fatal("Error running" + name)
	}

	s.RegisterProtocolInstance(pi)
	go pi.Dispatch()
	go pi.Start()

	return pi, err
}

// Service Phases
//______________________________________________________________________________________________________________________

// StartService starts the service (with all its different steps/protocols)
func (s *Service) StartService(targetQuery QueryID, root bool, exactPath bool) error {

	log.Lvl1(s.ServerIdentity(), " starts  Protocol for query ", targetQuery)

	//get DB configuration
	if _, err := toml.DecodeFile("db.toml", &dbConfig); err != nil {
		log.Fatal("Error: The database configuration is not valid")
	}

	//prepare SQL query statement
	queryStmt := s.PrepareQueryStatement(exactPath)

	//execute query to DB along with aggregation
	start0 := time.Now()
	resultSet, counts := s.ExecuteSqlQuery(&queryStmt)
	log.LLvl1("SQL Query Time: ", time.Since(start0))

	//perform aggregation
	start1 := time.Now()
	aggregatedResultSet := s.AggregateResultSet(resultSet, counts)
	log.LLvl1("Aggregation Time: ", time.Since(start1))

	//TODO: add obfuscation for differential privacy

	//copy aggregatedResultSet groups in list of string and aggregated counts in a CipherVector
	s.AggregatedResults = append(s.AggregatedResults, lib.NewFilteredResponse(0, 0))

	for key := range *aggregatedResultSet {
		s.Groups = append(s.Groups, key)
		s.AggregatedResults[0].AggregatingAttributes = append(s.AggregatedResults[0].AggregatingAttributes, *((*aggregatedResultSet)[key]))

	}

	//log.Lvl1(s.ServerIdentity(), " tests database ", *aggregatedResultSet)

	// Key Switch Phase
	start2 := time.Now()
	if root == true {
		start := lib.StartTimer(s.ServerIdentity().String() + "_KeySwitchingPhase")

		s.KeySwitchingPhase()

		lib.EndTimer(start)
	}
	log.LLvl1("Re-encryption Time: ", time.Since(start2))

	return nil
}

// KeySwitchingPhase performs the switch to the querier's key on the currently aggregated data.
func (s *Service) KeySwitchingPhase() error {
	pi, err := s.StartProtocol(protocols.KeySwitchingProtocolName)
	if err != nil {
		return err
	}

	s.KeySwitchedAggregatedResults = <-pi.(*protocols.KeySwitchingProtocol).FeedbackChannel

	return err
}

// Query and DB management
//______________________________________________________________________________________________________________________
func (s *Service) ExecuteSqlQuery(query *string) (*map[string][]string, *lib.CipherVector) {

	// open connection to DB
	db, err := sql.Open("postgres", "user="+dbConfig.Username+" password="+dbConfig.Password+" dbname="+dbConfig.DbName+" sslmode=disable")
	if err != nil {
		log.Fatal("Error: The data source arguments are not valid")
	}
	defer db.Close()

	//ping database to check the connection
	log.Lvl1(s.ServerIdentity(), " tests database connection")
	err = db.Ping()
	if err != nil {
		log.Fatal("Error: could not establish a connection with the database")
	}

	//local variable to store query results
	var loc, yr, cpt, count string
	var cipherText *lib.CipherText
	//var toEncryptInt int64

	log.Lvl1(s.ServerIdentity(), " runs query: ", *query)
	//execute query and check for potential errors
	rows, err := db.Query(*query)
	if err == sql.ErrNoRows {
		log.Fatal("No Results Found")
	}
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	//read each line of the result set and store the location, year and concept in a resultSet map and the totalnum
	//in a separate list
	log.Lvl1(s.ServerIdentity(), " reads query results and create resultSet")
	//instantiate resultSet map
	resultSet := make(map[string][]string)
	var counts lib.CipherVector

	//read rows obtained after SQL query
	for rows.Next() {
		err = rows.Scan(&loc, &yr, &cpt, &count)
		if err != nil {
			log.Fatal(err)
		}
		resultSet["location_cd"] = append(resultSet["location_cd"], loc)
		resultSet["year"] = append(resultSet["year"], yr)
		resultSet["concept_path"] = append(resultSet["concept_path"], cpt)

		//uncomment for deployed
		//-------------------------------------------------------------------
		cipherText = lib.NewCipherTextFromBase64(count)
		//-------------------------------------------------------------------

		//uncomment for test
		//-------------------------------------------------------------------
		//toEncryptInt, err = strconv.ParseInt(count, 10, 64)
		//if err != nil {
		//	log.Fatal(err)
		//}
		//cipherText = lib.EncryptInt(s.Query.Roster.Aggregate, toEncryptInt)
		//-------------------------------------------------------------------
		counts = append(counts, *cipherText)

	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return &resultSet, &counts

}

func (s *Service) AggregateResultSet(resultSet *map[string][]string, counts *lib.CipherVector) *map[string]*lib.CipherText {

	log.Lvl1(s.ServerIdentity(), " performs result aggregation of the resultSet")
	aggregatedResultSet := make(map[string]*lib.CipherText)

	//from the resultSet map create a new map where keys are group identifiers (specified in the initial query)
	//and values are the summations of counts in the same group
	for i := 0; i < len(*counts); i++ {

		key := ""
		if len(s.Query.GroupBy) > 0 {
			for j, gr := range s.Query.GroupBy {
				key += (*resultSet)[gr][i]
				if j < len(s.Query.GroupBy)-1 {
					key += ","
				}

			}
		} else {
			key += "total"
		}

		if _, ok := aggregatedResultSet[key]; !ok {
			aggregatedResultSet[key] = &((*counts)[i])
		} else {
			aggregatedResultSet[key].Add(*(aggregatedResultSet[key]), (*counts)[i])
		}

	}

	//TODO: implement parallelized version
	/*
		wg := lib.StartParallelize(len(*counts))
		mutexD := sync.Mutex{}

		for i := 0; i < len(*counts); i++ {
			if lib.PARALLELIZE {
				go func(i int) {
					defer wg.Done()
					key := ""
					if len(s.Query.GroupBy) > 0 {
						for _, gr := range s.Query.GroupBy {
							key += (*resultSet)[gr][i]
							key += ","
						}
					} else {
						key += "total,"
					}

					if _, ok := aggregatedResultSet[key]; !ok {
						mutexD.Lock()
						aggregatedResultSet[key] = &((*counts)[i])
						mutexD.Unlock()
					} else {
						mutexD.Lock()
						aggregatedResultSet[key].Add(*(aggregatedResultSet[key]), (*counts)[i])
						mutexD.Unlock()
					}
				}(i)
			}*/
	return &aggregatedResultSet
}

func (s *Service) PrepareQueryStatement(exactPath bool) string {

	// select statement
	selectStmt := "SELECT * "

	// from statement
	fromStmt := "FROM " + dbConfig.Table

	// where statement
	whereStmt := " WHERE "

	conceptPaths := ""
	if len(s.Query.Concepts) > 0 {
		conceptPaths += "("
		for i, path := range s.Query.Concepts {
			if(exactPath){
				conceptPaths += "concept_path="
				conceptPaths += "'" + doubleBackSlashes(path) + "'"
			}else{
				conceptPaths += "concept_path LIKE"
				conceptPaths += "'%" + doubleBackSlashes(path) + "%'"
			}
			
			if i < len(s.Query.Concepts)-1 {
				conceptPaths += " OR "
			}
		}
		conceptPaths += ")"
	}

	times := ""
	if len(s.Query.Times) > 0 {
		if len(s.Query.Concepts) > 0 {
			times += " AND ("
		} else {
			times += "("
		}
		for i, tm := range s.Query.Times {
			times += "time LIKE"
			times += " '" + tm + "%'"
			if i < len(s.Query.Times)-1 {
				times += " OR "
			}
		}
		times += ")"
	}

	locationCodes := ""
	if len(s.Query.Locations) > 0 {
		if len(s.Query.Concepts) > 0 || len(s.Query.Times) > 0 {
			locationCodes += " AND ( "
		} else {
			locationCodes += "("
		}
		for i, loc := range s.Query.Locations {
			locationCodes += "location_cd="
			locationCodes += "'" + loc + "'"
			if i < len(s.Query.Locations)-1 {
				locationCodes += " OR "
			}
		}
		locationCodes += ")"
	}

	whereStmt += conceptPaths + times + locationCodes

	//order by statement (optional)
	orderByStmt := " ORDER BY location_cd ASC;"

	// query statement
	queryStmt := selectStmt + fromStmt + whereStmt + orderByStmt

	return queryStmt
}

func doubleBackSlashes(text string) string{
	temp := strings.Replace(text,"\\\\i2b2_DIAG","",1)
	return strings.Replace(temp,"\\","\\\\",-1)
}