package importer

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/GO-HEML-prototype/src/utility"
)

type mnistDatas struct {
	Data 	[]mnistData
}

type mnistData struct {
	Image 	[]float64
	Label 	int
}

type MnistData struct {
	Image 	[]float64
	Label 	[]float64
}

type EncryptedMnistData struct {
	Image 	ckks.Ciphertext
	Label 	ckks.Ciphertext
}


func getMnistData(filepath string) mnistDatas {
	jsonFile, _ := os.Open(filepath)
	defer jsonFile.Close()
	file, _ := ioutil.ReadAll(jsonFile)

	var data mnistDatas
	json.Unmarshal(file, &data)

	return data
}

type MnistDataLoader struct {

	utils		utility.Utils
	RawData 	[]MnistData

}

func (m *MnistDataLoader) LoadData(filepath string) {
	
	for _, data := range getMnistData(filepath).Data{

		label := make([]float64, 10)
		label[data.Label] = 1

		m.RawData = append(m.RawData, MnistData{data.Image, label})

	}

}

func (m *MnistDataLoader) GetDataAsBatch(batch int, batchSize int) []EncryptedMnistData{

	start := batch * batchSize
	end := start + batchSize

	result := make([]EncryptedMnistData, batchSize)

	for i, data := range m.RawData[start:end]{

		result[i] = EncryptedMnistData{m.utils.Encrypt(data.Image), m.utils.Encrypt(data.Label)}

	}

	return result

}