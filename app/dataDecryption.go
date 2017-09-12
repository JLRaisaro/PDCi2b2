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

func decryptIntFromApp(c *cli.Context) error {

	// cli arguments
	keyFilePath := c.String("key")
	b1, err := ioutil.ReadFile(keyFilePath)
	secKey, err := lib.DeserializeScalar(string(b1))
	if err != nil {
		log.Error(err)
		return cli.NewExitError(err, 4)
	}

	if c.NArg() != 1 {
		err := errors.New("Wrong number of arguments (only 1 allowed, except for the flags)")
		log.Error(err)
		return cli.NewExitError(err, 3)
	}

	// value to decrypt from file
	/*toDecryptFilePath := c.Args().Get(0)
	b2, err := ioutil.ReadFile(toDecryptFilePath)
	toDecrypt := lib.NewCipherTextFromBase64(string(b2))*/

	// value to decrypt form command line
	toDecryptSerialized := c.Args().Get(0)
	toDecrypt := lib.NewCipherTextFromBase64(toDecryptSerialized)

	// decryption
	decVal := lib.DecryptInt(secKey, *toDecrypt)

	// output on stdout
	resultString := strconv.FormatInt(decVal, 10) + "\n"
	_, err = io.WriteString(os.Stdout, resultString)
	if err != nil {
		log.Error("Error while writing result.", err)
		return cli.NewExitError(err, 4)
	}

	return nil
}
