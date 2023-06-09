package management

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/tuneinsight/lattigo/v4/rlwe"
)

// Raw struct with string encrypted data

type EncryptedDataRaw struct {
	Name      string      `json:"name"`
	Encrypted []ColumnRaw `json:"encryptedData"`
}

type ColumnRaw struct {
	ColumnName string     `json:"columnName"`
	Type       string     `json:"type"`
	Length     int        `json:"length"`
	Data       string     `json:"data"`
	Label      []LabelRaw `json:"label"`
}

type LabelRaw struct {
	Category string `json:"category"`
	Index    int    `json:"index"`
	Data     string `json:"data"`
}

// Struct with marshalled data

type EncryptedData struct {
	Name      string   `json:"name"`
	Encrypted []Column `json:"encryptedData"`
}

type Column struct {
	ColumnName string           `json:"columnName"`
	Type       string           `json:"type"`
	Length     int              `json:"length"`
	Data       *rlwe.Ciphertext `json:"data"`
	Label      []Label          `json:"label"`
}

type Label struct {
	Category string           `json:"category"`
	Index    int              `json:"index"`
	Data     *rlwe.Ciphertext `json:"data"`
}

func LoadJsonData(filePath string) (EncryptedData, error) {

	// Open  jsonFile
	jsonFile, err := os.Open(filePath)

	// if we os.Open returns an error then handle it
	if err != nil {
		return EncryptedData{}, err
	}

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)

	// if returns an error then handle it
	if err != nil {
		return EncryptedData{}, err
	}

	var rawData EncryptedDataRaw

	// we unmarshal our byteArray which contains our
	err = json.Unmarshal(byteValue, &rawData)

	// if returns an error then handle it
	if err != nil {
		return EncryptedData{}, err
	}

	// Generate encrypted data struct to store unmarshaled result
	data := EncryptedData{Name: rawData.Name, Encrypted: make([]Column, len(rawData.Encrypted))}

	for i := range rawData.Encrypted {

		data.Encrypted[i] = Column{
			ColumnName: rawData.Encrypted[i].ColumnName,
			Type:       rawData.Encrypted[i].Type,
			Length:     rawData.Encrypted[i].Length,
			Data:       &rlwe.Ciphertext{},
			Label:      make([]Label, len(rawData.Encrypted[i].Label)),
		}

		if rawData.Encrypted[i].Data != "" {

			byteCt, err := base64.StdEncoding.DecodeString(rawData.Encrypted[i].Data)

			if err != nil {
				return EncryptedData{}, err
			}

			err = data.Encrypted[i].Data.UnmarshalBinary(byteCt)

			if err != nil {
				return EncryptedData{}, err
			}

		}

		if len(rawData.Encrypted[i].Label) != 0 {

			for j := range rawData.Encrypted[i].Label {

				data.Encrypted[i].Label[j] = Label{
					Category: rawData.Encrypted[i].Label[j].Category,
					Index:    rawData.Encrypted[i].Label[j].Index,
					Data:     &rlwe.Ciphertext{},
				}

				byteCt, err := base64.StdEncoding.DecodeString(rawData.Encrypted[i].Label[j].Data)

				if err != nil {
					return EncryptedData{}, err
				}

				err = data.Encrypted[i].Label[j].Data.UnmarshalBinary(byteCt)

				if err != nil {
					return EncryptedData{}, err
				}

			}

		}

	}

	return data, nil

}

func SaveToJson(data EncryptedData, filePath string) error {

	// Generate encrypted data struct to store unmarshaled result
	rawData := EncryptedDataRaw{Name: data.Name, Encrypted: make([]ColumnRaw, len(data.Encrypted))}

	for i := range rawData.Encrypted {

		rawData.Encrypted[i] = ColumnRaw{
			ColumnName: rawData.Encrypted[i].ColumnName,
			Type:       rawData.Encrypted[i].Type,
			Length:     rawData.Encrypted[i].Length,
			Data:       "",
			Label:      make([]LabelRaw, len(data.Encrypted[i].Label)),
		}

		if data.Encrypted[i].Data != nil {

			byteCt, err := data.Encrypted[i].Data.MarshalBinary()

			if err != nil {
				return err
			}

			rawData.Encrypted[i].Data = base64.StdEncoding.EncodeToString(byteCt)

		}

		if len(data.Encrypted[i].Label) != 0 {

			for j := range data.Encrypted[i].Label {

				rawData.Encrypted[i].Label[j] = LabelRaw{
					Category: rawData.Encrypted[i].Label[j].Category,
					Index:    rawData.Encrypted[i].Label[j].Index,
					Data:     "",
				}

				byteCt, err := data.Encrypted[i].Label[j].Data.MarshalBinary()

				if err != nil {
					return err
				}

				rawData.Encrypted[i].Label[j].Data = base64.StdEncoding.EncodeToString(byteCt)
			}

		}

	}

	jsonData, err := json.Marshal(rawData)

	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filePath, jsonData, 0777)

	return err

}
