package main

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/JLRgithub/PrivateI2B2DQ/lib"
	"gopkg.in/dedis/onet.v1/log"
	"gopkg.in/urfave/cli.v1"
)

func encryptIntFromApp(c *cli.Context) error {

	//cli arguments
	keyFilePath := c.String("key")
	b, err := ioutil.ReadFile(keyFilePath)
	pubKey, err := lib.DeserializePoint(string(b))
	if err != nil {
		log.Error(err)
		return cli.NewExitError(err, 4)
	}

	//uncomment for using collective key
	/*// cli arguments
	groupFilePath := c.String("file")

	// generate el with group file
	el, err := openGroupToml(groupFilePath)
	if err != nil {
		log.Error("Error while opening group file", err)
		return cli.NewExitError(err, 1)
	}*/

	if c.NArg() != 1 {
		err := errors.New("Wrong number of arguments (only 1 allowed, except for the flags)")
		log.Error(err)
		return cli.NewExitError(err, 3)
	}

	toEncrypt := c.Args().Get(0)
	toEncryptInt, err := strconv.ParseInt(toEncrypt, 10, 64)
	if err != nil {
		log.Error(err)
		return cli.NewExitError(err, 4)
	}

	// encrypt
	encryptedInt := lib.EncryptInt(pubKey, toEncryptInt)

	// homomorphic operations
	/*encryptedInt.Add(*encryptedInt, *encryptedInt)
	encryptedInt.MulCipherTextbyScalar(*encryptedInt, suite.Scalar().SetInt64(2))*/

	// test encryption for ETL
	/*originalDatasetNbr := 250000
	start := time.Now()
	for i:=0; i< originalDatasetNbr; i++ {
		lib.EncryptInt(pubKey, toEncryptInt)
	}
	log.LLvl1("1x: ", time.Since(start))

	start = time.Now()
	for i:=0; i< originalDatasetNbr*2; i++ {
		lib.EncryptInt(el.Aggregate, toEncryptInt)
	}
	log.LLvl1("2x: ", time.Since(start))

	start = time.Now()
	for i:=0; i< originalDatasetNbr*3; i++ {
		lib.EncryptInt(el.Aggregate, toEncryptInt)
	}
	log.LLvl1("4x: ", time.Since(start))*/

	// output in xml format on stdout
	resultString := (*encryptedInt).Serialize()
	_, err = io.WriteString(os.Stdout, resultString)
	if err != nil {
		log.Error("Error while writing result.", err)
		return cli.NewExitError(err, 4)
	}

	return nil
}
