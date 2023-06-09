package regression

import (
	"strconv"

	"github.com/tuneinsight/lattigo/v4/rlwe"
	"github.com/perm-ai/go-cerebrum/activations"
	"github.com/perm-ai/go-cerebrum/logger"
	"github.com/perm-ai/go-cerebrum/utility"
)

type LogisticRegression struct {
	utils  utility.Utils
	Weight []*rlwe.Ciphertext
	Bias   *rlwe.Ciphertext
}

type LogisticRegressionGradient struct {
	dw []*rlwe.Ciphertext
	db *rlwe.Ciphertext
}

type Data struct {
	x          []*rlwe.Ciphertext
	target     *rlwe.Ciphertext
	datalength int
}

func NewLogisticRegression(u utility.Utils, numOfFeatures int) LogisticRegression {

	value := u.GenerateFilledArray(0.0)
	b := u.EncryptToPointer(value)
	w := make([]*rlwe.Ciphertext, numOfFeatures)
	for i := 0; i < numOfFeatures; i++ {
		w[i] = u.EncryptToPointer(value)
	}

	return LogisticRegression{u, w, b}

}

func (model LogisticRegression) Forward(data Data) *rlwe.Ciphertext {

	//prediction(yhat) = sigmoid(w1*x1+w2*x2+...+b)
	result := model.utils.EncryptToPointer(model.utils.GenerateFilledArray(0.0))
	sigmoid := activations.Sigmoid{U: model.utils}
	//w[i]*x[i]
	for i := range data.x {
		Weight := model.utils.MultiplyNew(model.Weight[i].CopyNew(), data.x[i].CopyNew(), true, false)
		model.utils.Add(Weight, result, result)

	}
	model.utils.Add(model.Bias, result, result)
	if result.Level() < 6 {
		model.utils.BootstrapInPlace(result)
	}
	Arrresult := make([]*rlwe.Ciphertext, 1)
	Arrresult[0] = result
	return sigmoid.Forward(Arrresult, data.datalength)[0]

}

func (model LogisticRegression) Backward(data Data, predict *rlwe.Ciphertext, lr float64) LogisticRegressionGradient {

	dw := make([]*rlwe.Ciphertext, len(model.Weight))
	err := model.utils.SubNew(predict, data.target)
	multiplier := model.utils.EncodePlaintextFromArray(model.utils.GenerateFilledArraySize((2.0/float64(data.datalength))*lr, data.datalength))

	for i := range model.Weight {

		dw[i] = model.utils.MultiplyNew(data.x[i].CopyNew(), err.CopyNew(), true, false)
		model.utils.SumElementsInPlace(dw[i])
		model.utils.MultiplyPlain(dw[i].CopyNew(), multiplier, dw[i], true, false)

	}

	db := model.utils.SumElementsNew(*err)
	model.utils.MultiplyPlain(db.CopyNew(), multiplier, db, true, false)

	return LogisticRegressionGradient{dw, db}

}

func (model *LogisticRegression) UpdateGradient(grad LogisticRegressionGradient) {
	for i := range grad.dw {
		model.utils.Sub(model.Weight[i], grad.dw[i], model.Weight[i])
	}
	model.utils.Sub(model.Bias, grad.db, model.Bias)

}

func (model *LogisticRegression) Train(x []*rlwe.Ciphertext, target *rlwe.Ciphertext, datalength int, learningRate float64, epoch int) {
	data := Data{x, target, datalength}
	log := logger.NewLogger(true)
	log.Log("Starting Logistic Regression Training on encrypted data")

	for i := 0; i < epoch; i++ {

		log.Log("Forward propagating " + strconv.Itoa(i+1) + "/" + strconv.Itoa(epoch))
		fwd := model.Forward(data)

		if fwd.Level() < 5 {
			model.utils.BootstrapInPlace(fwd)
		}

		log.Log("Backward propagating " + strconv.Itoa(i+1) + "/" + strconv.Itoa(epoch))
		grad := model.Backward(data, fwd, learningRate)

		log.Log("Updating gradient " + strconv.Itoa(i+1) + "/" + strconv.Itoa(epoch) + "\n")
		model.UpdateGradient(grad)

		if i != epoch-1 {
			if model.Weight[0].Level() < 8 {
				for i := range model.Weight {
					model.utils.BootstrapInPlace(model.Weight[i])
				}
			}

			if model.Bias.Level() < 5 {
				model.utils.BootstrapInPlace(model.Bias)
			}
		}

	}
	log.Log("Trainning complete")
}
