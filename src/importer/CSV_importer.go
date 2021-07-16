package importer

import (
	"bufio"
	"encoding/csv"

	// "encoding/json"
	// "fmt"
	"io"
	"log"
	"os"
	"strconv"
)

type csvData struct {
	FirstData  []float64
	SecondData []float64
}

func GetCSV(filepath string, colNum1 int, colNum2 int) csvData {

	// path, column number1, column number2

	csvFile, _ := os.Open(filepath)
	reader := csv.NewReader(bufio.NewReader(csvFile))

	var data csvData

	for {
		line, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			log.Fatal(error)
		}
		firstData, _ := strconv.ParseFloat(line[colNum1], 64)
		secondData, _ := strconv.ParseFloat(line[colNum2], 64)

		data.FirstData = append(data.FirstData, firstData)
		data.SecondData = append(data.SecondData, secondData)
	}

	return data
}
