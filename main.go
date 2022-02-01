package main

import (
	"fmt"
	"log"
	"math"
	"os"

	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/go-cerebrum/activations"
	"github.com/perm-ai/go-cerebrum/dataset"
	"github.com/perm-ai/go-cerebrum/key"
	"github.com/perm-ai/go-cerebrum/layers"
	"github.com/perm-ai/go-cerebrum/logger"
	"github.com/perm-ai/go-cerebrum/losses"
	"github.com/perm-ai/go-cerebrum/management"
	"github.com/perm-ai/go-cerebrum/models"
	"github.com/perm-ai/go-cerebrum/utility"
)

func main() {

	fmt.Println("Starting Neural Network Banpu Training")

	BATCH_SIZE := 40
	LEARNING_RATE := 0.3
	EPOCH := 100

	importedData, err := management.LoadJsonData("/usr/local/go/src/github.com/perm-ai/go-cerebrum/importer/test-data/coal_neural_network.json")
	if err != nil {
		log.Fatal(err)
	}
	xCipherTexts := make(map[string]*ckks.Ciphertext)
	yCipherTexts := make([]*ckks.Ciphertext, 1)

	for i := range importedData.Encrypted {
		if importedData.Encrypted[i].ColumnName != "price" {
			xCipherTexts[importedData.Encrypted[i].ColumnName] = importedData.Encrypted[i].Data
		} else {
			yCipherTexts[0] = importedData.Encrypted[i].Data
		}
	}

	order := []string{"TM_AR", "TS_AR", "M_AD", "ASH_AD", "ASH_AR", "Sulfate_SO3", "Silica_SiO2", "Calcium_CaO", "Iron_Fe2O3"}

	keyPair := key.LoadKeys("/usr/local/go/src/github.com/perm-ai/go-cerebrum/keychain", 0, true, true, false, false)
	keychain := key.GenerateKeysFromKeyPair(0, keyPair.SecretKey, keyPair.PublicKey, true, true)

	utils := utility.NewUtils(keychain, math.Pow(2, 35), 0, true)

	batchSize := 40
	learningRate := 0.3
	epoch := 100
	var relu activations.Activation
	relu = activations.Relu{}

	fmt.Println("Dense 1 generating")

	dense1 := layers.NewDense(utils, 9, 50, &relu, true, batchSize, learningRate, 5)

	fmt.Println("Dense 2 generating")

	dense2 := layers.NewDense(utils, dense1.GetOutputSize(), 1, nil, true, batchSize, learningRate, 3)

	dense1.SetBootstrapOutput(true, "forward")
	dense1.SetBootstrapOutput(true, "backward")
	// dense2.SetBootstrapOutput(true, "backward")
	// dense2.SetBootstrapActivation(true, "forward")

	model := models.NewModel(utils, []layers.Layer1D{
		&dense1, &dense2,
		}, []layers.Layer2D{}, losses.MSE{U: utils}, false)

	loader := dataset.NewStandardLoader(xCipherTexts, order, yCipherTexts)

	timer := logger.StartTimer("Neural Network Training")

	model.Train1D(loader, learningRate, batchSize, epoch)

	timer.LogTimeTakenSecond()

	os.Mkdir("test_model_banpu_1", 0777)

	model.ExportModel1D("test_model_banpu_1")
	

}
