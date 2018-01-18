package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type PipeBackendStruct struct {
	Domain  string         `json:"domain"`
	Type    string         `json:"type"`
	Num     int            `json:"num"`
	Records []RecordStruct `json:"record"`
}

type RecordStruct struct {
	IP     string `json:"ip"`
	Weight int    `json:"weight"`
	TTL    int    `json:"ttl"`
}

func main() {
	rand.Seed(time.Now().UnixNano())

	backend := PipeBackendStruct{}

	scanner := bufio.NewScanner(os.Stdin)
	//output("output", "Start Backend Program")

	scanner.Scan()
	s := strings.TrimRight(scanner.Text(), "\n")
	if s != "HELO\t1" {
		fmt.Println("FAIL")
		os.Exit(0)
	}
	fmt.Println("OK Sample Backend firing up\t")

	for scanner.Scan() {
		strs := strings.TrimRight(scanner.Text(), "\n")
		str := strings.Split(strs, "\t")
		if len(str) < 6 {
			fmt.Println("LOG\tPowerDNS sent unparseable line")
			continue
		}
		bytes, err := ioutil.ReadFile("/home/vagrant/backend/config.json")
		if err != nil {
			outputError(err)
			return
		}
		err = json.Unmarshal(bytes, &backend)
		if err != nil {
			outputError(err)
			return
		}
		switch str[3] {
		case "A", "ANY":
			if str[1] == backend.Domain {
				ips, ttl := selectIPs(backend)
				for _, i := range ips {
					if i == "" {
						break
					}
					fmt.Println("DATA\t" + str[1] + "\t" + str[2] + "\tA\t" + strconv.Itoa(ttl) + "\t" + "1\t" + i)
					//output("DATA\t" + str[1] + "\t" + str[2] + "\tA\t" + strconv.Itoa(ttl) + "\t" + "-1\t" + i)
				}
				go outputTTL(ttl)

			} else {
				output(str)
			}
		case "SOA":
			if str[1] == "sai.test" {
				fmt.Println("DATA\t" + str[1] + "\t" + str[2] + "\tSOA\t60\t-1\tsai.test. root.sai.test 2017112601 60 60 60 60")
				//output("DATA\t" + str[1] + "\t" + str[2] + "\tSOA\t60\t-1\tsai.test. root.sai.test 2017112601 60 60 60 60")
			}
		default:
			//output(str)
		}
		fmt.Println("END")
	}
}

func selectIPs(backend PipeBackendStruct) ([]string, int) {

	ttl := 10000000
	ips := make([]string, backend.Num)

	if len(backend.Records) == 1 {
		ips[0] = backend.Records[0].IP
		return ips, backend.Records[0].TTL
	}
	// すべての重みが0のときの処理
	if checkWeight(backend.Records) {
		for i := 0; i < backend.Num; i++ {
			ips[i] = backend.Records[i].IP
			if ttl > backend.Records[i].TTL {
				ttl = backend.Records[i].TTL
			}
		}
		return ips, ttl
	}

	//records := make([]RecordStruct, len(backend.Records))
	records := copyStruct(backend.Records)

	for k := 0; k < backend.Num && len(records) > 0; k++ {
		weight := make([]int, len(records))
		for i, v := range records {
			weight[i] = v.Weight
		}

		boundaries := make([]int, len(records)+1)
		for i := 1; i < len(boundaries); i++ {
			boundaries[i] = boundaries[i-1] + weight[i-1]
		}

		boundaryLast := int(boundaries[len(boundaries)-1])
		x := rand.Intn(boundaryLast) + 1
		idx := sort.SearchInts(boundaries, int(x)) - 1

		ips[k] = records[idx].IP
		if ttl > records[idx].TTL {
			ttl = records[idx].TTL
		}
		records = pop(records, idx)
	}
	return ips, ttl
}

func copyStruct(slice []RecordStruct) []RecordStruct {

	records := []RecordStruct{}

	for _, v := range slice {
		if v.Weight != 0 {
			records = append(records, v)
		}
	}
	return records

}

func pop(slice []RecordStruct, index int) []RecordStruct {

	result := []RecordStruct{}
	for i, v := range slice {
		if i != index {
			result = append(result, v)
		}
	}
	return result
}

func checkWeight(slice []RecordStruct) bool {

	for _, v := range slice {
		if v.Weight != 0 {
			return false
		}
	}

	return true

}

func outputError(err error) {

	f, err := os.OpenFile("/home/vagrant/backend/error.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	defer f.Close()
	if err != nil {
		return
	}
	log.SetOutput(f)
	log.Println(err)
}

func output(str interface{}) {

	f, err := os.OpenFile("/home/vagrant/backend/output.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	defer f.Close()
	if err != nil {
		return
	}
	log.SetOutput(f)
	log.Print(fmt.Sprintln(str))
}

func outputTTL(str interface{}) {

	now := time.Now().Format("15:04:05")

	f, err := os.OpenFile("/home/vagrant/backend/ttl.csv", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	defer f.Close()
	if err != nil {
		outputError(err)
		return
	}

	_, err = f.WriteString(now + "," + fmt.Sprintln(str))
	if err != nil {
		outputError(err)
	}
	return
}
