package main

import (
	//"gopkg.in/dedis/onet.v1/app"
	"os"

	"gopkg.in/dedis/onet.v1/app"
	"gopkg.in/dedis/onet.v1/log"
	"gopkg.in/urfave/cli.v1"
)

const (
	// BinaryName is the name of the unlynx app
	BinaryName = "i2b2dc"

	// Version of the binary
	Version = "1.00"

	// DefaultGroupFile is the name of the default file to lookup for group
	// definition
	DefaultGroupFile = "group.toml"

	optionConfig      = "config"
	optionConfigShort = "c"

	optionGroupFile      = "file"
	optionGroupFileShort = "f"

	// query flags

	optionLocation      = "location"
	optionLocationShort = "l"

	optionTime      = "time"
	optionTimeShort = "t"

	optionConceptCode      = "concept"
	optionConceptCodeShort = "c"

	optionGroupBy      = "groupBy"
	optionGroupByShort = "g"

	// decryption flags

	optionDecryptKey      = "key"
	optionDecryptKeyShort = "k"

	optionEncryptKey      = "key"
	optionEncryptKeyShort = "k"

	// csv manipulation flags

	optionCsvFileIn      = "csvIn"
	optionCsvFileInShort = "i"

	optionCsvFileOut      = "csvOut"
	optionCsvFileOutShort = "o"

	attributeToEncrypt      = "attribute"
	attributeToEncryptShort = "a"
)

func main() {
	cliApp := cli.NewApp()
	cliApp.Name = BinaryName
	cliApp.Usage = "Query i2b2 summary data stored on the cloud in the i2b2 data model under homomorphic encryption"
	cliApp.Version = Version

	binaryFlags := []cli.Flag{
		cli.IntFlag{
			Name:  "debug, d",
			Value: 0,
			Usage: "debug-level: 1 for terse, 5 for maximal",
		},
	}

	manipulateCsvFlags := []cli.Flag{
		cli.StringFlag{
			Name: optionGroupFile + ", " + optionGroupFileShort,
			//Value: DefaultGroupFile,
			Usage: "Servers' group definition `FILE`",
		},
		cli.StringFlag{
			Name:  optionCsvFileIn + ", " + optionCsvFileInShort,
			Usage: "input CSV `FILE` to be encrypted/decrypted",
		},

		cli.StringFlag{
			Name:  optionCsvFileOut + ", " + optionCsvFileOutShort,
			Usage: "output CSV `FILE` after encryption/decryption",
		},

		cli.StringFlag{
			Name:  optionEncryptKey + ", " + optionEncryptKeyShort,
			Usage: "`FILE` with base64-encoded public key",
		},

		cli.StringFlag{
			Name:  attributeToEncrypt + ", " + attributeToEncryptShort,
			Usage: "name of the header attribute in the CSV file to encrypt/decrypted",
		},
	}

	encryptFlags := []cli.Flag{
		cli.StringFlag{
			Name: optionGroupFile + ", " + optionGroupFileShort,
			//Value: DefaultGroupFile,
			Usage: "Servers' group definition `FILE`",
		},
		cli.StringFlag{
			Name:  optionEncryptKey + ", " + optionEncryptKeyShort,
			Usage: "`FILE` with base64-encoded public key",
		},
	}

	decryptFlags := []cli.Flag{
		cli.StringFlag{
			Name:  optionDecryptKey + ", " + optionDecryptKeyShort,
			Usage: "`FILE` with base64-encoded secret key",
		},
	}

	queryFlags := []cli.Flag{
		cli.StringFlag{
			Name:  optionGroupFile + ", " + optionGroupFileShort,
			Value: DefaultGroupFile,
			Usage: "Servers' definition `FILE`",
		},

		// query flags

		cli.StringFlag{
			Name:  optionLocation + ", " + optionLocationShort,
			Usage: "Specify location codes in the SQL-WHERE clause. E.g., hosp1 or hosp2 -> LOC:hosp1,LOC:hosp2",
		},
		cli.StringFlag{
			Name:  optionTime + ", " + optionTimeShort,
			Usage: "Specify time frames in the SQL-WHERE clause. E.g., 2015 or 2016 -> 2015,2016",
		},
		cli.StringFlag{
			Name:  optionConceptCode + ", " + optionConceptCodeShort,
			Usage: "Specify the concepts codes in the SQL-WHERE clause. E.g., ICD10:E08 or ICD10:E09 -> ICD10:E08, ICD10:E08",
		},
		cli.StringFlag{
			Name:  optionGroupBy + ", " + optionGroupByShort,
			Usage: "Specify the attributes in the SQL-GROUPBY clause. Possible values: 'lcoation_cd', 'concept_cd', 'year'",
		},
		cli.StringFlag{
			Name:  optionCsvFileOut + ", " + optionCsvFileOutShort,
			Usage: "Specify the output csv `FILE`",
		},
	}

	serverFlags := []cli.Flag{
		cli.StringFlag{
			Name:  optionConfig + ", " + optionConfigShort,
			Value: app.GetDefaultConfigFile(BinaryName),
			Usage: "Configuration file of the server",
		},
	}

	cliApp.Commands = []cli.Command{
		// BEGIN CLIENT: DATA ENCRYPTION ----------
		{
			Name:    "encrypt",
			Aliases: []string{"e"},
			Usage:   "Encrypt an integer with the public key of the collective authority",
			Action:  encryptIntFromApp,
			Flags:   encryptFlags,
		},
		// CLIENT END: DATA ENCRYPTION ------------

		// BEGIN CLIENT: DATA ENCRYPTION ----------
		{
			Name:    "encryptCsv",
			Aliases: []string{"ecsv"},
			Usage:   "Encrypt a CSV file attribute with an ElGamal public key",
			Action:  encryptCsvFileFromApp,
			Flags:   manipulateCsvFlags,
		},
		// CLIENT END: DATA ENCRYPTION ------------

		// BEGIN CLIENT: DATA ENCRYPTION ----------
		{
			Name:    "decryptCsv",
			Aliases: []string{"dcsv"},
			Usage:   "Decrypt CSV file attribute(s) with and ElGamal secret key",
			Action:  decryptCsvFileFromApp,
			Flags:   manipulateCsvFlags,
		},
		// CLIENT END: DATA ENCRYPTION ------------

		// BEGIN CLIENT: DATA DECRYPTION ----------
		{
			Name:    "decrypt",
			Aliases: []string{"d"},
			Usage:   "Decrypt an integer with the provided private key",
			Action:  decryptIntFromApp,
			Flags:   decryptFlags,
		},
		// CLIENT END: DATA DECRYPTION ------------

		// BEGIN CLIENT: KEY GENERATION ----------
		{
			Name:    "keygen",
			Aliases: []string{"k"},
			Usage:   "Generate a pair of public/private keys.",
			Action:  keyGenerationFromApp,
		},
		// CLIENT END: KEY GENERATION ------------

		// BEGIN CLIENT: QUERIER ----------
		{
			Name:    "run",
			Aliases: []string{"r"},
			Usage:   "Run i2b2DC querying service",
			Action:  runQuery,
			Flags:   queryFlags,
		},
		// CLIENT END: QUERIER ----------

		// BEGIN SERVER --------
		{
			Name:  "server",
			Usage: "Start i2b2dc server",
			Action: func(c *cli.Context) error {
				runServer(c)
				return nil
			},
			Flags: serverFlags,
			Subcommands: []cli.Command{
				{
					Name:    "setup",
					Aliases: []string{"s"},
					Usage:   "Setup server configuration (interactive)",
					Action: func(c *cli.Context) error {
						if c.String(optionConfig) != "" {
							log.Fatal("[-] Configuration file option cannot be used for the 'setup' command")
						}
						if c.GlobalIsSet("debug") {
							log.Fatal("[-] Debug option cannot be used for the 'setup' command")
						}
						app.InteractiveConfig(BinaryName)
						return nil
					},
				},
			},
		},
		// SERVER END ----------
	}

	cliApp.Flags = binaryFlags
	cliApp.Before = func(c *cli.Context) error {
		log.SetDebugVisible(c.GlobalInt("debug"))
		return nil
	}
	err := cliApp.Run(os.Args)
	log.ErrFatal(err)
}
