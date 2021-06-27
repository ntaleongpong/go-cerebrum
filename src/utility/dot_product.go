package utility

import "github.com/ldsec/lattigo/v2/ckks"

func (u Utils) rotateAndAdd(ct *ckks.Ciphertext, size float64) ckks.Ciphertext {

	midpoint := size / 2

	rotated := u.Evaluator.RotateNew(ct, uint64(midpoint), &u.GaloisKey)
	u.Add(*ct, *rotated, ct)

	if midpoint == 1 {
		return *ct
	} else {
		return u.rotateAndAdd(ct, midpoint)
	}

}

func (u Utils) SumElementsInPlace(ct *ckks.Ciphertext) {

	u.rotateAndAdd(ct, float64(u.Params.LogSlots()))

}

func (u Utils) SumElementsNew(ct ckks.Ciphertext) ckks.Ciphertext {

	return u.rotateAndAdd(&ct, float64(u.Params.LogSlots()))

}

func (u Utils) DotProduct(a ckks.Ciphertext, b ckks.Ciphertext, destination *ckks.Ciphertext) {

	u.MultiplyRescale(a, b, destination)
	u.SumElementsInPlace(destination)

}

func (u Utils) DotProductNew(a *ckks.Ciphertext, b *ckks.Ciphertext) ckks.Ciphertext {

	result := u.MultiplyRescaleNew(a, b)
	u.SumElementsInPlace(&result)

	return result

}
