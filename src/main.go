package main

import (

	//"math"
	// "fmt"

	"fmt"
	"math"

	//"github.com/perm-ai/GO-HEML-prototype/src/importer"
	//"github.com/perm-ai/GO-HEML-prototype/src/utility"
	// "fmt"

	"github.com/perm-ai/GO-HEML-prototype/src/ml"
	"github.com/perm-ai/GO-HEML-prototype/src/utility"
)

func main() {

	// lrData := importer.GetTitanicData("./test-data/titanic1.json")
	// x := lrData.Age
	// y := lrData.Pclass
	// target := lrData.Target
	// ml.Normalize_Data(x)
	// ml.Normalize_Data(y)
	// logisticRegression := ml.NewLogisticRegression()
	// ml.Train(logisticRegression, x, y, target, 0.1, 20)

	// Acc := 0.0
	// for i := 0; i < 10; i++ {
	// 	x, y, target := utility.GenerateLinearData(300)
	// 	logisticRegression := ml.NewLogisticRegression()
	// 	Acc += ml.Train(logisticRegression, x, y, target, 0.1, 20)
	// }
	// fmt.Printf("Average Accuracy : %f", Acc/10)
	datasetSize := 300
	utils := utility.NewUtils(math.Pow(2, 35), 0, true, true)
	data1, data2, data3 := utility.GenerateLinearData(datasetSize)
	logisticRegression := ml.NewLogisticRegression(utils)

	x := utils.Encrypt(data1)
	y := utils.Encrypt(data2)
	target := utils.Encrypt(data3)

	logisticRegression.TrainLR(x, y, target, 0.1, datasetSize, 1)

	accuracy := logisticRegression.AccuracyTest(data1, data2, data3, datasetSize)
	fmt.Printf("The accuracy of this logistic regression model is %f percent", accuracy)
}
