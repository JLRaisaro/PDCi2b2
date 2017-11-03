package main

import(
	"net/http"
	"io/ioutil"
	"gopkg.in/dedis/onet.v1/log"
	"github.com/gorilla/mux"
	"time"
	"fmt"
	"encoding/json"
	"gopkg.in/urfave/cli.v1"
	"strconv"
	"github.com/JLRgithub/PrivateDCi2b2/services"
	"gopkg.in/dedis/onet.v1"
)

type State struct{
	address string
	group string
}

func webServer(c *cli.Context){
	r := mux.NewRouter()
	address := ""
	if a := c.String("address"); a != "" {
		address = a
	}
	group := ""
	if g := c.String("file"); g != "" {
		group = g
	}

	s := State{address,group}
	
	r.HandleFunc("/totalnum", s.totalNumHandler)
	r.HandleFunc("/totalnums", s.totalNumsHandler)

	srv := &http.Server{
		Handler:      r,
		Addr:         address,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

type Result struct{
	ConceptPath string `json:"conceptpath"`
	TotalNum string `json:"totalnum"`
}
type ResultGroup struct{
	Group string `json:"group"`
	TotalNum string `json:"totalnum"`
}
type Response struct{
	Concepts []Result `json:"concepts"`
}
type ResponseGroup struct{
	Groups []ResultGroup `json:"groups"`
}
/*
This handler finds the total num for each of the concept paths given as input, no group by location nor time is performed.
The results are aggregated to return 1 encrypted total num per concept path.
*/
func (state State) totalNumHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	body, err1 := ioutil.ReadAll(r.Body)
	if(err1!=nil){
		log.Fatal("Could not read request body.", err1)
	}
	var pathsObj map[string]interface{}
	json.Unmarshal(body,&pathsObj)
	paths := pathsObj["conceptpaths"].([]interface{})
	keyString := pathsObj["clientpublickey"].(string)
	fmt.Println("totalNum received public key : ",keyString)
	if(len(paths)>0){

		el, err := openGroupToml(state.group)
		if err != nil {
			log.Fatal("Could not open group toml file.",err)
		}

		client := serviceI2B2dc.NewClientFromKey(el.List[0], strconv.Itoa(0), keyString, false)

		var results []Result
		for _,path := range paths{
			results = append(results, queryAggr(path.(string),client,el)...)
		}
		
		res := &Response{results}
		resJson,_ := json.Marshal(res)
		w.Write(resJson)
	}
}
/*
This handler finds the total nums for a given concept path according to location and time. It returns 2 lists of encrypted
total nums, 1 where the results are grouped by location, the other by time.
*/
func (state State) totalNumsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	body, err1 := ioutil.ReadAll(r.Body)
	if(err1!=nil){
		log.Fatal("Could not read request body.", err1)
	}
	var pathObj map[string]interface{}
	json.Unmarshal(body,&pathObj)
	path := pathObj["conceptpath"].(string)
	keyString := pathObj["clientpublickey"].(string)
	fmt.Println("totalNums received path : ",path)
	fmt.Println("and public key : ",keyString)

	el, err := openGroupToml(state.group)
	if err != nil {
		log.Fatal("Could not open group toml file.",err)
	}
	client := serviceI2B2dc.NewClientFromKey(el.List[0], strconv.Itoa(0), keyString, false)

	results := queryGroupBy(path,client,el)
	res := &ResponseGroup{results}
	resJson,_ := json.Marshal(res)
	w.Write(resJson)
}

func queryAggr(path string, client  *serviceI2B2dc.APIremote, el *onet.Roster) []Result{
	queryID, err := client.SendQuery(el, serviceI2B2dc.QueryID(""), nil, []string{}, []string{}, []string{path}, []string{})
	if err != nil {
		log.Fatal("Service did not start.", err)
	}
	grps, aggr, err := client.ExecuteQuery(*queryID)
	if err != nil {
		log.Fatal("Query could not be executed.", err)
	}
	var results []Result
	if(grps!=nil && aggr!=nil){
		for i := 0; i < len(*grps); i++ {
			r := Result{path, (*aggr)[i]}
			fmt.Println(r)
			results = append(results, r)
		}
	}else{
		results = append(results, Result{path, (*aggr)[0]})
	}
	return results
}

func queryGroupBy(path string, client  *serviceI2B2dc.APIremote, el *onet.Roster) []ResultGroup{
	var groupBy []string
	groupBy = append(groupBy, "location_cd")
	groupBy = append(groupBy, "year")
	queryID, err := client.SendQuery(el, serviceI2B2dc.QueryID(""), nil, []string{}, []string{}, []string{path}, groupBy)
	if err != nil {
		log.Fatal("Service did not start.", err)
	}

	grps, aggr, err := client.ExecuteQuery(*queryID)
	if err != nil {
		log.Fatal("Query could not be executed.", err)
	}
	var results []ResultGroup
	if(grps!=nil && aggr!=nil){
		for i := 0; i < len(*grps); i++ {
			r := ResultGroup{(*grps)[i], (*aggr)[i]}
			fmt.Println(r)
			results = append(results, r)
		}
	}
	return results
}