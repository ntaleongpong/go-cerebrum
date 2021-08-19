package layers

import (
	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/go-cerebrum/activations"
	"github.com/perm-ai/go-cerebrum/array"
	"github.com/perm-ai/go-cerebrum/utility"
)

//=================================================
//		   CONVOLUTIONAL KERNEL (FILTER)
//=================================================

type conv2dKernel struct {
	Row		int
	Column	int
	Depth	int
	Data 	[][][]ckks.Ciphertext
}

func generateRandomNormal2dKernel(row int, col int, depth int, utils utility.Utils) conv2dKernel {

	randomNums := array.GenerateRandomNormalArray(row * col * depth)
	data := make([][][]ckks.Ciphertext, row)

	for r := 0; r < row; r++ {

		data[r] = make([][]ckks.Ciphertext, col)

		for c := 0; c < col; c++ {

			data[r][c] = make([]ckks.Ciphertext, depth)

			for d := 0; d < depth; d++{
				data[r][c][d] = utils.Encrypt(utils.GenerateFilledArray(randomNums[(r * col) + c]))
			}
		}
	}

	return conv2dKernel{row, col, depth, data}

}

//=================================================
//		   		CONVOLUTIONAL LAYER
//=================================================

type Conv2D struct {
	utils		utility.Utils
	Kernels		[]conv2dKernel
	Bias		[]ckks.Ciphertext
	Strides		[]int
	Padding		bool
	Activation	*activations.Activation
	InputSize	[]int
	batchSize	int
}

// Constructor for Convolutional layer struct
// filters is number of kernel in this layer
// kernelSize is the size of kernel with row in index 0 and column in index 1
// strides specify the stride in convolution along rows and columns
// padding specify the padding when evaluating convolution. Set to true to add padding of size 1 to each side and false for no padding
// activation specify the activation function
// useBias specify whether or not a bias will be used
// inputSize specify the size of input in the following order [row, column, channel] if there's channel, or [row, column] if there's no channel
// batchSize, when useBias is true, is the number of training example in a training ciphertexts with lowest number of training data
func NewConv2D(utils utility.Utils, filters int, kernelSize []int, strides []int, padding bool, activation *activations.Activation, useBias bool, inputSize []int, batchSize int) Conv2D{

	kernels := make([]conv2dKernel, filters)

	for i := range kernels{
		kernels[i] = generateRandomNormal2dKernel(kernelSize[0], kernelSize[1], inputSize[2], utils)
	}

	bias := []ckks.Ciphertext{}
	if useBias{
		bias = make([]ckks.Ciphertext, filters)
		for i := range bias {
			bias[i] = utils.Encrypt(utils.GenerateRandomNormalArray(batchSize))
		}
	}

	return Conv2D{utils: utils, Kernels: kernels, Bias: bias, Strides: strides, Padding: padding, Activation: activation, InputSize: inputSize, batchSize: batchSize}

}

// Evaluate forward pass of the convolutional 2d layer
// input must be packed according to section 3.1.1 in https://eprint.iacr.org/2018/1056.pdf
func (c Conv2D) Forward (input [][][]*ckks.Ciphertext) [][][]*ckks.Ciphertext {

	// Calculate the starting coordinate
	start := 0
	if c.Padding {
		start = -1
	}

	// (W1−F+2P)/S+1
	outputRowSize := int(float64(c.InputSize[0] - c.Kernels[0].Row + (2 * (start * (-1)))) / float64(c.Strides[0] + 1))

	// (H1−F+2P)/S+1 
	outputColumnSize := int(float64(c.InputSize[1] - c.Kernels[1].Row + (2 * (start * (-1)))) / float64(c.Strides[1] + 1))

	// Get kernel dimention
	kernelDim := [2]int{c.Kernels[0].Row, c.Kernels[0].Column}

	output := make([][][]*ckks.Ciphertext, outputRowSize)

	for row := start; row < (c.InputSize[0] - kernelDim[0]); row += c.Strides[0]{

		output[row] = make([][]*ckks.Ciphertext, outputColumnSize)

		for col := start; col < (c.InputSize[1] - kernelDim[1]); col += c.Strides[1]{

			output[row][col] = make([]*ckks.Ciphertext, len(c.Kernels))
			
			for k := range c.Kernels{

				var result *ckks.Ciphertext

				for krow := 0; krow < c.Kernels[k].Row; krow++ {
					for kcol := 0; kcol < c.Kernels[k].Column; kcol++{
						
						// Check if in padding
						if row + krow == -1 || col + kcol == -1 {
							continue
						}

						for kdep := 0; kdep < c.Kernels[k].Depth; kdep++{

							product := c.utils.MultiplyNew(c.Kernels[k].Data[krow][kcol][kdep], *input[row + krow][col + kcol][kdep], true, false)

							if result == nil {
								result = &product
							} else {
								c.utils.Add(product, *result, result)
							}

						}
					}
				}

				if len(c.Bias) != 0{
					c.utils.Add(c.Bias[k], *result, result)
				}

				if c.Activation != nil{
					activatedResult := (*c.Activation).Forward(*result, c.InputSize[1])
					result = &activatedResult
				}

				output[row][col][k] = result
				
			}

		}
	}

	return output

}