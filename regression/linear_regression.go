package regression

import (
	"fmt"
	"math"

	"github.com/tuneinsight/lattigo/v4/rlwe"
	"github.com/perm-ai/go-cerebrum/logger"
	"github.com/perm-ai/go-cerebrum/utility"
)

type LinearRegression struct {
	utils  utility.Utils
	Weight []*rlwe.Ciphertext
	Bias   *rlwe.Ciphertext
}

type LinearRegressionGradient struct {
	DM []*rlwe.Ciphertext
	DB *rlwe.Ciphertext
}

// need to pass in number of independent features
func NewLinearRegression(u utility.Utils, numOfFeatures int) LinearRegression {

	zeros := u.GenerateFilledArray(0.0)
	m := make([]*rlwe.Ciphertext, numOfFeatures)
	for i := 0; i < numOfFeatures; i++ {
		m[i] = u.EncryptToPointer(zeros)
	}
	b := u.EncryptToPointer(zeros)

	return LinearRegression{u, m, b}

}

func (l LinearRegression) Forward(input []*rlwe.Ciphertext) *rlwe.Ciphertext {
	
	result := l.utils.InterDotProduct(input, l.Weight, true, true, nil)

	l.utils.Add(result, l.Bias, result)

	return result

}

func (l LinearRegression) Backward(input []*rlwe.Ciphertext, output *rlwe.Ciphertext, y *rlwe.Ciphertext, size int, learningRate float64) LinearRegressionGradient {

	// Calculate backward gradient using the following equation
	// dM = (-2/n) * sum(input * (label - prediction)) * learning_rate
	// dB = (-2/n) * sum(label - prediction) * learning_rate

	err := l.utils.SubNew(y, output)

	dM := make([]*rlwe.Ciphertext, len(input))
	multiplier := l.utils.EncodePlaintextFromArray(l.utils.GenerateFilledArraySize((-2.0/float64(size))*learningRate, size))

	// for i := range input {
	// 	dM[i] = l.utils.MultiplyNew(*input[i], *err.CopyNew(), true, false)
	// 	l.utils.SumElementsInPlace(&dM[i])
	// 	l.utils.MultiplyPlain(&dM[i], &multiplier, &dM[i], true, false)
	// }

	channels := make([]chan *rlwe.Ciphertext, len(input))

	for i := range input {
		channels[i] = make(chan *rlwe.Ciphertext)
		go func(index int, utils utility.Utils, channel chan *rlwe.Ciphertext) {
			product := utils.MultiplyNew(input[index], err.CopyNew(), true, false)
			utils.SumElementsInPlace(product)
			result := utils.MultiplyPlainNew(product, multiplier, true, false)

			channel <- result
		}(i, l.utils.CopyWithClonedEval(), channels[i])
	}

	for c := range channels {
		dM[c] = <-channels[c]
	}

	dB := l.utils.SumElementsNew(*err)
	l.utils.MultiplyPlain(dB, multiplier, dB, true, false)

	return LinearRegressionGradient{dM, dB}

}

func (l *LinearRegression) UpdateGradient(gradient LinearRegressionGradient) {

	for i := range gradient.DM {
		l.utils.Sub(l.Weight[i], gradient.DM[i], l.Weight[i])
	}

	l.utils.Sub(l.Bias, gradient.DB, l.Bias)

}

// pack data in array of ciphertexts
func (model *LinearRegression) Train(x []*rlwe.Ciphertext, y *rlwe.Ciphertext, learningRate float64, size int, epoch int) {

	log := logger.NewLogger(true)

	log.Log("Starting Linear Regression Training on encrypted data")

	for i := 0; i < epoch; i++ {

		log.Log(fmt.Sprintf("Forward propagating %d/%d (current lvl: %d)", i+1, epoch, x[0].Level()))
		fwd := model.Forward(utility.Clone1dCiphertext(x))

		log.Log(fmt.Sprintf("Backward propagating %d/%d(current lvl: %d)", i+1, epoch, fwd.Level()))
		grad := model.Backward(utility.Clone1dCiphertext(x), fwd, y.CopyNew(), size, learningRate)

		log.Log(fmt.Sprintf("Updating gradient %d/%d(current lvl: %d)\n", i+1, epoch, grad.DM[0].Level()))
		model.UpdateGradient(grad)

		if model.Weight[0].Level() < 4 || model.Bias.Level() < 4 {

			for i := range x {
				model.utils.BootstrapInPlace(model.Weight[i])
			}

			if model.Bias.Scale.Float64() < math.Pow(2,50){
				model.utils.Evaluator.ScaleUp(model.Bias, rlwe.NewScale(math.Pow(2, 60)/model.Bias.Scale.Float64()), model.Bias)
			}

			model.utils.BootstrapInPlace(model.Bias)
			
		}

	}

}
