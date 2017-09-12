package main

import (
	"encoding/csv"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/JLRgithub/PrivateI2B2DQ/lib"
	"gopkg.in/dedis/onet.v1/log"
	"gopkg.in/urfave/cli.v1"
)

func decryptCsvFileFromApp(c *cli.Context) error {

	//cli arguments
	csvFileInPath := c.String("csvIn")
	csvFileOutPath := c.String("csvOut")
	attributeToEncrypt := c.String("attribute")
	keyFilePath := c.String("key")

	//check that the number of arguments is 0
	if c.NArg() != 0 {
		err := errors.New("Wrong number of arguments (no arguments are allowed, except for the flags)")
		log.Error(err)
		return cli.NewExitError(err, 3)
	}

	start := time.Now()
	err := decryptCsvFile(&csvFileInPath, &csvFileOutPath, &attributeToEncrypt, &keyFilePath)
	if err != nil {
		log.Error(err)
		return cli.NewExitError(err, 3)
	}
	log.LLvl1("Encryption time: ", time.Since(start))

	return nil
}

func decryptCsvFile(inPath, outPath, attribute, key *string) error {
	//setup reader
	csvIn, err := os.Open(*inPath)
	if err != nil {
		log.Fatal(err)
		return cli.NewExitError(err, 4)
	}
	r := csv.NewReader(csvIn)
	defer csvIn.Close()

	//setup writer
	csvOut, err := os.Create(*outPath)
	if err != nil {
		log.Fatal(err)
		return cli.NewExitError(err, 4)
	}
	w := csv.NewWriter(csvOut)
	defer csvOut.Close()

	//setup public key for encryption
	b, err := ioutil.ReadFile(*key)
	secKey, err := lib.DeserializeScalar(string(b))
	if err != nil {
		log.Error(err)
		return cli.NewExitError(err, 4)
	}

	//read and write header
	rec, err := r.Read()
	if err != nil {
		log.Fatal(err)
		return cli.NewExitError(err, 4)
	}

	headerMap := convertSliceToMap(&rec)

	if err = w.Write(rec); err != nil {
		log.Fatal(err)
		return cli.NewExitError(err, 4)
	}

	//loop over records
	for {
		// read record
		rec, err = r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
			return cli.NewExitError(err, 4)
		}
		// decrypt record's fields corresponding to the input attributes
		listAttributes := strings.Split(*attribute, ",")
		for i := 0; i < len(listAttributes); i++ {
			toDecrypt := lib.NewCipherTextFromBase64(rec[headerMap[listAttributes[i]]])
			decVal := lib.DecryptInt(secKey, *toDecrypt)
			rec[headerMap[listAttributes[i]]] = strconv.FormatInt(decVal, 10)
		}
		// write record
		if err = w.Write(rec); err != nil {
			log.Fatal(err)
			return cli.NewExitError(err, 4)
		}
		w.Flush()
	}

	return nil

}
