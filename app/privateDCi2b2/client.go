package main

import (
	"os"

	"regexp"
	"strconv"
	"strings"

	"encoding/csv"

	"time"

	"github.com/JLRgithub/PrivateI2B2DQ/lib"
	"github.com/JLRgithub/PrivateI2B2DQ/services/i2b2dc"
	"github.com/btcsuite/goleveldb/leveldb/errors"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/app"
	"gopkg.in/dedis/onet.v1/log"
	"gopkg.in/urfave/cli.v1"
)

// BEGIN CLIENT: QUERIER ----------
func startQuery(servers *onet.Roster, locations, times, concepts, groupBy []string, out string) {

	start := time.Now()
	// create
	client := serviceI2B2dc.NewClient(servers.List[0], strconv.Itoa(0))
	queryID, err := client.SendQuery(servers, serviceI2B2dc.QueryID(""), nil, locations, times, concepts, groupBy)
	if err != nil {
		log.Fatal("Service did not start.", err)
	}

	// execute query
	grps, aggr, err := client.ExecuteQuery(*queryID)
	if err != nil {
		log.Fatal("Query could not be executed.", err)
	}
	end := time.Since(start)

	// print output
	log.Lvl1(client, "outputs query resuls: ", *grps, *aggr)

	// save output in Csv file
	// print output
	log.Lvl1(client, "writes results to: ", out)
	if out != "" {
		csvOut, err := os.Create(out)
		if err != nil {
			log.Fatal(err)
		}
		w := csv.NewWriter(csvOut)
		defer csvOut.Close()

		//writing header with elements in the groupBy statement + totalnum
		if len(groupBy) > 0 {
			groupBy = append(groupBy, "totalnum")
			err = w.Write(groupBy)
			if err != nil {
				log.Fatal("the output Csv file cannot be written", err)
			}
		}

		var record []string
		for i := 0; i < len(*grps); i++ {
			record = strings.Split((*grps)[i], ",")
			record = append(record, strconv.FormatInt((*aggr)[i], 10))
			err = w.Write(record)
			if err != nil {
				log.Fatal("the output Csv file cannot be written", err)
			}
			w.Flush()
		}
	}
	log.Lvl1("Total query response time:", end)
}

func runQuery(c *cli.Context) error {
	tomlFileName := c.String("file")

	// query parameters
	location := []string{}
	time := []string{}
	concept := []string{}
	groupBy := []string{}
	out := ""

	if loc := c.String("location"); loc != "" {
		location = strings.Split(loc, ",")
	}
	if tm := c.String("time"); tm != "" {
		time = strings.Split(tm, ",")
	}
	if cpt := c.String("concept"); cpt != "" {
		concept = strings.Split(cpt, ",")
	}
	if gb := c.String("groupBy"); gb != "" {
		groupBy = strings.Split(gb, ",")
	}
	out = c.String("csvOut")

	//check that the number of arguments is 0
	/*if c.NArg() != 0 {
		err := errors.New("Wrong number of arguments (no arguments are allowed, except for the flags)")
		log.Error(err)
		return cli.NewExitError(err, 3)
	}*/

	el, err := openGroupToml(tomlFileName)
	if err != nil {
		return err
	}

	startQuery(el, location, time, concept, groupBy, out)

	return nil
}

func openGroupToml(tomlFileName string) (*onet.Roster, error) {
	f, err := os.Open(tomlFileName)
	if err != nil {
		return nil, err
	}
	el, err := app.ReadGroupToml(f)
	if err != nil {
		return nil, err
	}

	if len(el.List) <= 0 {
		return nil, errors.New("Empty or invalid unlynx group file:" + tomlFileName)
	}

	return el, nil
}

func checkRegex(input, expression, errorMessage string) {
	var aux = regexp.MustCompile(expression)

	correct := aux.MatchString(input)

	if !correct {
		log.Fatal(errorMessage)
	}
}

func parseQuery(el *onet.Roster, sum string, count bool, where, predicate, groupBy string) ([]string, bool, []lib.WhereQueryAttribute, string, []string) {

	if sum == "" || (where != "" && predicate == "") || (where == "" && predicate != "") {
		log.Fatal("Wrong query! Please check the sum, where and the predicate parameters")
	}

	sumRegex := "{s[0-9]+(,\\s*s[0-9]+)*}"
	whereRegex := "{(w[0-9]+(,\\s*[0-9]+))*(,\\s*w[0-9]+(,\\s*[0-9]+))*}"
	groupByRegex := "{g[0-9]+(,\\s*g[0-9]+)*}"

	checkRegex(sum, sumRegex, "Error parsing the sum parameter(s)")
	sum = strings.Replace(sum, " ", "", -1)
	sum = strings.Replace(sum, "{", "", -1)
	sum = strings.Replace(sum, "}", "", -1)
	sumFinal := strings.Split(sum, ",")

	if count {
		check := false
		for _, el := range sumFinal {
			if el == "count" {
				check = true
			}
		}

		if !check {
			log.Fatal("No 'count' attribute in the sum variables")
		}
	}

	checkRegex(where, whereRegex, "Error parsing the where parameter(s)")
	where = strings.Replace(where, " ", "", -1)
	where = strings.Replace(where, "{", "", -1)
	where = strings.Replace(where, "}", "", -1)
	tmp := strings.Split(where, ",")

	whereFinal := make([]lib.WhereQueryAttribute, 0)

	var variable string
	for i := range tmp {
		// if is a variable (w1, w2...)
		if i%2 == 0 {
			variable = tmp[i]
		} else { // if it is a value
			value, err := strconv.Atoi(tmp[i])
			if err != nil {
				log.Fatal("Something wrong with the where value")
			}

			whereFinal = append(whereFinal, lib.WhereQueryAttribute{Name: variable, Value: *lib.EncryptInt(el.Aggregate, int64(value))})
		}
	}

	checkRegex(groupBy, groupByRegex, "Error parsing the groupBy parameter(s)")
	groupBy = strings.Replace(groupBy, " ", "", -1)
	groupBy = strings.Replace(groupBy, "{", "", -1)
	groupBy = strings.Replace(groupBy, "}", "", -1)
	groupByFinal := strings.Split(groupBy, ",")

	return sumFinal, count, whereFinal, predicate, groupByFinal
}

// CLIENT END: QUERIER ----------
