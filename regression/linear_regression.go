package regression

import (
	"fmt"
	"strconv"

	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/go-cerebrum/logger"
	"github.com/perm-ai/go-cerebrum/utility"
)

type LinearRegression struct {
	utils utility.Utils
	M     ckks.Ciphertext
	B     ckks.Ciphertext
}

type LinearRegressionGradient struct {
	DM ckks.Ciphertext
	DB ckks.Ciphertext
}

func NewLinearRegression(u utility.Utils) LinearRegression {

	zeros := u.GenerateFilledArray(0.0)
	m := u.Encrypt(zeros)
	b := u.Encrypt(zeros)

	return LinearRegression{u, m, b}

}

func (l LinearRegression) Forward(input *ckks.Ciphertext) ckks.Ciphertext {

	result := l.utils.MultiplyNew(*input, l.M, true, false)

	l.utils.Add(result, l.B, &result)

	return result

}

func (l LinearRegression) Backward(input *ckks.Ciphertext, output ckks.Ciphertext, y *ckks.Ciphertext, size int, learningRate float64) LinearRegressionGradient {

	// Calculate backward gradient using the following equation
	// dM = (-2/n) * sum(input * (label - prediction)) * learning_rate
	// dB = (-2/n) * sum(label - prediction) * learning_rate

	err := l.utils.SubNew(*y, output)

	dM := l.utils.MultiplyNew(*input, *err.CopyNew(), true, false)
	l.utils.SumElementsInPlace(&dM)
	l.utils.MultiplyConstArray(&dM, l.utils.GenerateFilledArraySize((-2/float64(size))*learningRate, size), &dM, true, false)

	dB := l.utils.SumElementsNew(err)
	l.utils.MultiplyConstArray(&dB, l.utils.GenerateFilledArraySize((-2/float64(size))*learningRate, size), &dB, true, false)

	return LinearRegressionGradient{dM, dB}

}

func (l *LinearRegression) UpdateGradient(gradient LinearRegressionGradient) {

	l.utils.Sub(l.M, gradient.DM, &l.M)
	l.utils.Sub(l.B, gradient.DB, &l.B)

}

func (model *LinearRegression) Train(x *ckks.Ciphertext, y *ckks.Ciphertext, learningRate float64, size int, epoch int) {

	log := logger.NewLogger(true)

	log.Log("Starting Linear Regression Training on encrypted data")

	for i := 0; i < epoch; i++ {

		log.Log("Forward propagating " + strconv.Itoa(i+1) + "/" + strconv.Itoa(epoch))
		fwd := model.Forward(x.CopyNew())

		log.Log("Backward propagating " + strconv.Itoa(i+1) + "/" + strconv.Itoa(epoch))
		grad := model.Backward(x.CopyNew(), fwd, y, size, learningRate)

		log.Log("Updating gradient " + strconv.Itoa(i+1) + "/" + strconv.Itoa(epoch) + "\n")
		model.UpdateGradient(grad)

		if model.M.Level() < 4 || model.B.Level() < 4 {
			fmt.Println("Bootstrapping gradient")
			if model.B.Level() != 1 {
				model.utils.Evaluator.DropLevel(&model.B, model.B.Level()-1)
			}
			model.utils.BootstrapInPlace(&model.M)
			model.utils.BootstrapInPlace(&model.B)
		}

	}

}