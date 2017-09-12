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
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/onet.v1/log"
	"gopkg.in/urfave/cli.v1"
)

func encryptCsvFileFromApp(c *cli.Context) error {

	//cli arguments
	csvFileInPath := c.String("csvIn")
	csvFileOutPath := c.String("csvOut")
	attributeToEncrypt := c.String("attribute")
	keyFilePath := c.String("key")
	serversFilePath := c.String("file")

	var encryptionKey abstract.Point

	if serversFilePath != "" && keyFilePath == "" {
		el, err := openGroupToml(serversFilePath)
		if err != nil {
			return cli.NewExitError(err, 4)
		}
		encryptionKey = el.Aggregate
	} else if serversFilePath == "" && keyFilePath != "" {
		b, err := ioutil.ReadFile(keyFilePath)
		encryptionKey, err = lib.DeserializePoint(string(b))
		if err != nil {
			log.Error(err)
			return cli.NewExitError(err, 4)
		}
	} else {
		err := errors.New("Key Error")
		log.Error(err)
		return cli.NewExitError(err, 4)
	}

	//check that the number of arguments is 0
	if c.NArg() != 0 {
		err := errors.New("Wrong number of arguments (no arguments are allowed, except for the flags)")
		log.Error(err)
		return cli.NewExitError(err, 3)
	}

	start := time.Now()
	err := encryptCsvFile(&csvFileInPath, &csvFileOutPath, &attributeToEncrypt, &encryptionKey)
	if err != nil {
		log.Error(err)
		return cli.NewExitError(err, 3)
	}
	log.LLvl1("Encryption time: ", time.Since(start))

	return nil
}

func encryptCsvFile(inPath, outPath, attribute *string, pubKey *abstract.Point) error {
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
		// encrypt record's fields corresponding to the input attributes
		listAttributes := strings.Split(*attribute, ",")
		for i := 0; i < len(listAttributes); i++ {
			toEncryptInt, err := strconv.ParseInt(rec[headerMap[listAttributes[i]]], 10, 64)
			if err != nil {
				log.Error(err)
				return cli.NewExitError(err, 4)
			}
			encryptedInt := lib.EncryptInt(*pubKey, toEncryptInt)
			rec[headerMap[listAttributes[i]]] = (*encryptedInt).Serialize()
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

func convertSliceToMap(s *[]string) map[string]int {
	m := make(map[string]int)
	for i := 0; i < len(*s); i++ {
		m[(*s)[i]] = i
	}

	return m
}
