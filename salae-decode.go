
package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func usage() {
	fmt.Printf("\nUsage: %s <CSV file>\n\n", os.Args[0])
}

var opcodeMap = map[int]string {
	0x5a: "SFDP",
	0x6: "write enable",
	0xb7: "enable 4ba mode",
	0x4: "write disable",
	0x3: "slow read",
	0xb: "fast read",
}

func decode(opcode int) string {
	val, err := opcodeMap[opcode]
	if err != true {
		return ""
	} else {
		return val
	}
}

func getNextByte(r *csv.Reader) (int, error) {
	fields, err := r.Read()

	if err == io.EOF {
		return -1, nil
	}

	if err != nil {
		return -1, err
	}

	w := strings.SplitN(fields[2], ";", 2)
	mosi := strings.SplitN(w[0], " ", 2)

	if mosi[1] == "" {
		mosi[1] = ";"
	}

	if matched, _ := regexp.MatchString(`'`, mosi[1]); matched == true {
		if mosi[1] == "'" {
			return 0x27, nil
		}
		if mosi[1] == "' '" {
			return 32, nil
		}

		quotes := regexp.MustCompile(`'`)
		return strconv.Atoi(quotes.ReplaceAllString(mosi[1], ""))
	}

	return int(mosi[1][0]), nil
}

func getAddr(r *csv.Reader) int {
	var addr int

	addr = 0

	for i := 0; i < 4; i++ {
		val, err := getNextByte(r)
		if err != nil {
			return 0
		}
		addr = (addr << 8) | val
	}
	return addr
}

func validOpcode(opcode int) bool {
	if opcode == 0 || opcode == 0x7f || opcode == 0x80 || opcode == 0xff {
		return false
	}
	return true
}

func readFile(fileName string) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}

	defer f.Close()
	r := csv.NewReader(f)
	r.LazyQuotes = true

	for {
		opcode, err := getNextByte(r)
		if err != nil {
			return err
		}
		if opcode == -1 {
			break
		}
		if validOpcode(opcode) == true {
			if opcode == 0x3 || opcode == 0xb {
				addr := getAddr(r)
				fmt.Printf("0x%02x @ 0x%08x (%s)\n", opcode, addr, decode(opcode))
			} else {
				fmt.Printf("0x%02x (%s)\n", opcode, decode(opcode))
			}
		}
	}

	return nil
}

func main() {
	if len(os.Args) != 2 {
		usage()
		log.Fatal("Incorrect number of arguments")
	}

	CSVFile := os.Args[1]
	fmt.Printf("File is %s\n", CSVFile)

	if err := readFile(CSVFile); err != nil {
		log.Fatal(err)
	}
}