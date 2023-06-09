package utility

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/tuneinsight/lattigo/v4/ckks"
	"github.com/tuneinsight/lattigo/v4/rlwe"
	"github.com/perm-ai/go-cerebrum/key"
	"github.com/perm-ai/go-cerebrum/logger"
)

var keyChain = key.GenerateKeys(0, true, true)
var utils = NewUtils(keyChain, math.Pow(2, 40), 100, true)
var log = logger.NewLogger(true)

type TestCase struct {
	data1       rlwe.Ciphertext
	data2       rlwe.Ciphertext
	rawData1    []float64
	rawData2    []float64
	addExpected []float64
	subExpected []float64
	mulExpected []float64
	dotExpected []float64
}

func GenerateTestCases(u Utils) [4]TestCase {

	rand.Seed(time.Now().UnixNano())
	log.Log("Generating Test Cases")
	random1 := make([]float64, utils.Params.Slots())
	random2 := make([]float64, utils.Params.Slots())
	randomAdd := make([]float64, utils.Params.Slots())
	randomSub := make([]float64, utils.Params.Slots())
	randomMul := make([]float64, utils.Params.Slots())
	randomDot := make([]float64, utils.Params.Slots())

	for i := 0; i < int(utils.Params.Slots()); i++ {
		random1[i] = rand.Float64() * 100
		random2[i] = rand.Float64() * 100
		randomAdd[i] = random1[i] + random2[i]
		randomSub[i] = random1[i] - random2[i]
		randomMul[i] = random1[i] * random2[i]
		randomDot[0] += randomMul[i]
	}

	// Normal ct (same scale, same level)
	t1 := TestCase{u.Encrypt(random1), u.Encrypt(random2), random1, random2, randomAdd, randomSub, randomMul, randomDot}

	// Ct with different scale, same level
	t2d1encd := u.EncodeToScale(random1, math.Pow(2, 30))
	t2d2encd := u.EncodeToScale(random2, math.Pow(2, 60))
	t2d1enct := u.Encryptor.EncryptNew(&t2d1encd)
	t2d2enct := u.Encryptor.EncryptNew(&t2d2encd)
	t2 := TestCase{*t2d1enct, *t2d2enct, random1, random2, randomAdd, randomSub, randomMul, randomDot}

	// Ct with different level, same scale
	t3d1enct := u.Encrypt(random1)
	t3d2enct := u.Encrypt(random2)
	u.Evaluator.DropLevel(&t3d2enct, 3)
	t3 := TestCase{t3d1enct, t3d2enct, random1, random2, randomAdd, randomSub, randomMul, randomDot}

	// Ct with different scale, different level
	t4d1encd := u.EncodeToScale(random1, math.Pow(2, 30))
	t4d2encd := u.EncodeToScale(random2, math.Pow(2, 60))
	t4d1enct := u.Encryptor.EncryptNew(&t4d1encd)
	t4d2enct := u.Encryptor.EncryptNew(&t4d2encd)
	u.Evaluator.DropLevel(t4d2enct, 3)
	t4 := TestCase{*t4d1enct, *t4d2enct, random1, random2, randomAdd, randomSub, randomMul, randomDot}

	return [4]TestCase{t1, t2, t3, t4}

}

func TestComplexToFloat(t *testing.T) {

	data := make([]complex128, utils.Params.Slots())
	data[0] = complex(324.4, 0)
	data[122] = complex(75916.3, 0)
	data[300] = complex(2334578347, 0)

	float := utils.Complex128ToFloat64(data)

	if !(float[0] == 324.4 && float[122] == 75916.3 && float[300] == 2334578347) {
		t.Error("Complex array wasn't correctly converted to float")
	}

}

func TestFloatToComplex(t *testing.T) {

	data := make([]float64, utils.Params.Slots())
	data[0] = 324.4
	data[122] = 75916.3
	data[300] = 2334578347

	float := utils.Float64ToComplex128(data)

	if !(float[0] == complex(324.4, 0) && float[122] == complex(75916.3, 0) && float[300] == complex(2334578347, 0)) {
		t.Error("Complex array wasn't correctly converted to float")
	}

}

func TestEncodingDecoding(t *testing.T) {

	data := utils.GenerateFilledArray(0.0)
	data[0] = 324.4
	data[122] = 75916.3
	data[300] = 2334556.3

	encoded := utils.Encode(data)
	decoded := utils.Decode(&encoded)

	fmt.Println(decoded[0])

	if !ValidateResult(decoded, data, false, 1, log) {
		t.Error("Data wasn't correctly encoded")
	}

}

func TestEncodingToScale(t *testing.T) {

	data := utils.GenerateFilledArray(0.0)
	data[0] = 324.4
	data[122] = 75916.3
	data[300] = 2334556.3

	encoded := utils.EncodeToScale(data, math.Pow(2.0, 20.0))
	decoded := utils.Decode(&encoded)

	if !ValidateResult(decoded, data, false, 1, log) {
		t.Error("Data wasn't correctly encoded to scale (2^80)")
	}

	data = utils.GenerateFilledArray(0.0)
	data[0] = 20
	data[122] = 30
	data[300] = 50

	encoded = utils.EncodeToScale(data, math.Pow(2.0, 60))
	decoded = utils.Decode(&encoded)

	if !ValidateResult(decoded, data, false, 1, log) {
		t.Error("Data wasn't correctly encoded to scale (2^60)")
	}

	data = utils.GenerateFilledArray(0.0)
	data[0] = 20
	data[122] = 30
	data[300] = 50

	encoded = utils.EncodeToScale(data, math.Pow(2.0, 80))
	decoded = utils.Decode(&encoded)

	if !ValidateResult(decoded, data, false, 1, log) {
		t.Error("Data wasn't correctly encoded to scale (2^80)")
	}

}

func TestEncryptionDecryption(t *testing.T) {

	data := utils.GenerateFilledArray(0.0)
	data[0] = 324.4
	data[5] = 2334556.3
	data[122] = 75916.3

	ct := utils.Encrypt(data)
	dt := utils.Decrypt(&ct)

	if !(math.Abs(dt[0]-324.4) < 0.1 && math.Abs(dt[122]-75916.3) < 0.1 && math.Abs(dt[5]-2334556.3) < 0.1) {
		t.Error("Data wasn't correctly Encrypted and Decrypted")
	}

}

func TestAddition(t *testing.T) {

	testCases := GenerateTestCases(utils)

	for i := range testCases {

		log.Log("Testing addition (" + strconv.Itoa(i+1) + "/4)")
		ct1 := testCases[i].data1
		ct2 := testCases[i].data2

		sum := utils.AddNew(&ct1, &ct2)
		addNewD := utils.Decrypt(sum)

		if !ValidateResult(addNewD, testCases[i].addExpected, false, 1, log) {
			t.Error("Data wasn't correctly added (AddNew)")
		}

		timer := logger.StartTimer("Add")
		utils.Add(&ct1, &ct2, sum)
		timer.LogTimeTaken()

		addD := utils.Decrypt(sum)

		if !ValidateResult(addD, testCases[i].addExpected, false, 1, log) {
			t.Error("Data wasn't correctly added (Add)")
		}

	}

}

func TestSubtraction(t *testing.T) {

	testCases := GenerateTestCases(utils)

	for i := range testCases {

		log.Log("Testing subtraction (" + strconv.Itoa(i+1) + "/4)")
		ct1 := testCases[i].data1
		ct2 := testCases[i].data2

		subNew := utils.SubNew(&ct1, &ct2)
		subNewD := utils.Decrypt(subNew)

		if !ValidateResult(subNewD, testCases[i].subExpected, false, 1, log) {
			t.Error("Data wasn't correctly subtracted (SubNew)")
		}

		utils.Sub(&ct1, &ct2, &ct1)
		subD := utils.Decrypt(&ct1)

		if !ValidateResult(subD, testCases[i].subExpected, false, 1, log) {
			t.Error("Data wasn't correctly subtracted (Sub)")
		}

	}

}

func TestMultiplication(t *testing.T) {

	testCases := GenerateTestCases(utils)

	for i := range testCases {

		log.Log("Testing multiplication (" + strconv.Itoa(i+1) + "/4)")
		ct1 := testCases[i].data1
		ct2 := testCases[i].data2

		mulNew := utils.MultiplyNew(&ct1, &ct2, false, true)
		mulNewD := utils.Decrypt(mulNew)

		if !ValidateResult(mulNewD, testCases[i].mulExpected, false, 1, log) {
			t.Error("Data wasn't correctly multiplied (MultiplyNew)")
		}

		newCiphertext1 := ckks.NewCiphertext(utils.Params, 1, utils.Params.MaxLevel())
		utils.Multiply(&ct1, &ct2, newCiphertext1, false, true)
		mulD := utils.Decrypt(newCiphertext1)

		if !ValidateResult(mulD, testCases[i].mulExpected, false, 1, log) {
			t.Error("Data wasn't correctly multiplied (Multiply)")
		}

		mulNewRes := utils.MultiplyNew(&ct1, &ct2, true, true)
		mulNewResD := utils.Decrypt(mulNewRes)

		if !ValidateResult(mulNewResD, testCases[i].mulExpected, false, 1, log) && mulNewRes.Scale.Float64() != ct1.Scale.Float64()*ct2.Scale.Float64() {
			t.Error("Data wasn't correctly multiplied (MultiplyRescaleNew)")
		}

		newCiphertext2 := ckks.NewCiphertext(utils.Params, 1, utils.Params.MaxLevel())
		utils.Multiply(&ct1, &ct2, newCiphertext2, true, true)
		mulResD := utils.Decrypt(newCiphertext2)

		if !ValidateResult(mulResD, testCases[i].mulExpected, false, 1, log) && newCiphertext2.Scale.Float64() != ct1.Scale.Float64()*ct2.Scale.Float64() {
			t.Error("Data wasn't correctly multiplied (MultiplyRescale)")
		}

	}

}

func TestExponential(t *testing.T) {

	random := utils.GenerateRandomFloatArray(100, -1, 3)
	expRandom := make([]float64, len(random))

	for i := range random {

		expRandom[i] = math.Exp(random[i])

	}

	ct := utils.EncryptToScale(random, math.Pow(2, 40))

	expCt := utils.ExpNew(&ct, 100)

	fmt.Println(ct.Level(), expCt.Level())

	if !ValidateResult(utils.Decrypt(expCt), expRandom, false, -0.3, log) {
		t.Error("Exponential function wasn't correctly evaluated")
	}

}

func TestInverse(t *testing.T) {

	randomArr := utils.GenerateRandomFloatArray(100, 20, 40)
	expected := make([]float64, len(randomArr))
	mulExpected := make([]float64, len(randomArr))

	for i := 0; i < 100; i++ {
		expected[i] = 1 / randomArr[i]
		mulExpected[i] = 5 * expected[i]
	}

	ct := utils.Encrypt(randomArr)

	inverse := utils.InverseNew(&ct, (float64(1) / float64(50)), 100)

	fmt.Printf("Cost: %d Levels\n", ct.Level()-inverse.Level())

	if !ValidateResult(utils.Decrypt(inverse)[0:100], expected[0:100], false, 1, log) {
		t.Error("Inverse wasn't correctly evaluated")
	}

	mulResult := utils.MultiplyConstArrayNew(*inverse, utils.GenerateFilledArraySize(5, 100), true, false)

	if !ValidateResult(utils.Decrypt(mulResult)[0:100], mulExpected[0:100], false, 1, log) {
		t.Error("Inversed ciphertext wasn't correctly multiplied with plaintext")
	}

	ct2 := utils.EncryptToScale(randomArr, 2475880078665336141973028864.0)

	inverseApprox := utils.InverseApproxNew(&ct2, (float64(1) / float64(50)), 100)
	fmt.Printf("Consumed %d levels\n", ct.Level()-inverseApprox.Level())

	utils.MultiplyConstArray(inverseApprox, utils.GenerateFilledArray((float64(1) / float64(50))), inverseApprox, true, false)
	if !ValidateResult(utils.Decrypt(inverseApprox)[0:100], expected[0:100], false, 1, log) {
		t.Error("Inversed Approx ciphertext wasn't correctly calculated")
	}

}

func TestSquareRoot(t *testing.T) {

	testCase := utils.GenerateRandomFloatArray(100, 0.2, 2)
	expect := make([]float64, 100)

	for i := range expect {
		expect[i] = math.Sqrt(testCase[i])
	}

	timer := logger.StartTimer("Square root")

	ct := utils.Encrypt(testCase)
	utils.Evaluator.DropLevel(&ct, ct.Level()-9)
	result := utils.Sqrt(&ct, 3, 100)

	timer.LogTimeTaken()

	decryptedResult := utils.Decrypt(result)[0:100]

	fmt.Printf("Before: %d, After: %d (took %d levels)", ct.Level(), result.Level(), (ct.Level() - result.Level()))

	if !ValidateResult(decryptedResult, expect, false, 1, log) {
		t.Error("Square root wasn't calculated correctly")
	}

}

func TestSquareRootPoly(t *testing.T) {

	testCase := utils.GenerateRandomFloatArray(100, 0.3, 1.75)
	expect := make([]float64, 100)

	for i := range expect {
		expect[i] = math.Sqrt(testCase[i])
	}

	timer := logger.StartTimer("Square root")

	ct := utils.Encrypt(testCase)
	utils.Evaluator.DropLevel(&ct, ct.Level()-9)
	result := utils.SqrtApprox(&ct, 3, 1.75, true)

	timer.LogTimeTaken()

	decryptedResult := utils.Decrypt(result)[0:100]

	fmt.Printf("Before: %d, After: %d (took %d levels)", ct.Level(), result.Level(), (ct.Level() - result.Level()))

	if !ValidateResult(decryptedResult, expect, false, 1, log) {
		t.Error("Square root wasn't calculated correctly")
	}

}

func TestDotProduct(t *testing.T) {

	testCases := GenerateTestCases(utils)

	ct1 := testCases[0].data1
	ct2 := testCases[0].data2

	dotNew := utils.DotProductNew(&ct1, &ct2, true)
	dotNewD := utils.Decrypt(dotNew)

	if !ValidateResult(dotNewD, testCases[0].dotExpected, true, -0.69, log) {
		t.Error("Dot product wasn't correctly calculated (DotProductNew)")
	}

	utils.DotProduct(&ct1, &ct2, &ct1, true)
	dotD := utils.Decrypt(&ct1)

	if !ValidateResult(dotD, testCases[0].dotExpected, true, -0.69, log) {
		t.Error("Dot product wasn't correctly calculated (DotProduct)")
	}

}

func TestBootstrapping(t *testing.T) {

	if !utils.bootstrapEnabled {
		t.Error("Bootstrapping isn't enabled")
	}

	ct := utils.Encrypt(utils.GenerateFilledArray(3.12))
	maxLevel := ct.Level()

	utils.Evaluator.DropLevel(&ct, ct.Level()-1)
	preBootstrap := ct.Level()

	utils.BootstrapIfNecessary(&ct)

	decrypted := utils.Decrypt(&ct)

	log.Log(fmt.Sprintf("Max Level: %d, Post Bootstrapping level: %d", maxLevel, ct.Level()))

	// Test if bootstrap increase level and correctly decrypt
	if ct.Level() <= preBootstrap || !ValidateResult(decrypted, utils.GenerateFilledArray(3.12), false, 1, log) {
		t.Error("Wasn't bootstrapped correctly")
	}

	encTwos := utils.Encrypt(utils.GenerateFilledArray(2))

	addResult := utils.AddNew(&ct, &encTwos)

	// Test if bootstrapped ciphertext can correctly evaluate addition
	if !ValidateResult(utils.Decrypt(addResult), utils.GenerateFilledArray(3.12+2), false, 1, log) {
		t.Error("Addition wasn't evaluated correctly after bootstrap")
	}

	productByConst := utils.MultiplyConstNew(&ct, 0.1, false, false)

	// Test if bootstrapped ciphertext can correctly evaluate addition
	if !ValidateResult(utils.Decrypt(productByConst), utils.GenerateFilledArray(3.12*0.1), false, 1, log) {
		t.Error("Multiplication by const wasn't evaluated correctly after bootstrap")
	}

	product := utils.MultiplyNew(&encTwos, &ct, true, true)

	// Test if bootstrapped ciphertext can correctly evaluate addition
	if !ValidateResult(utils.Decrypt(product), utils.GenerateFilledArray(3.12*2), false, 1, log) {
		t.Error("Multiplication wasn't evaluated correctly after bootstrap")
	}

}

func TestConcurrentBootstrapping(t *testing.T) {

	ct1 := utils.EncryptToPointer(utils.GenerateFilledArray(3.12))
	ct2 := utils.EncryptToPointer(utils.GenerateFilledArray(4.20))
	ct3 := utils.EncryptToPointer(utils.GenerateFilledArray(6.9))
	maxLevel := ct1.Level()

	utils.Evaluator.DropLevel(ct1, ct1.Level()-9)
	utils.Evaluator.DropLevel(ct2, ct2.Level()-9)
	utils.Evaluator.DropLevel(ct3, ct3.Level()-9)
	preBootstrap := ct1.Level()

	data := []*rlwe.Ciphertext{ct1, ct2}

	timer := logger.StartTimer("TestConcurrentBootstrapping")

	utils.Bootstrap1dInPlace(data, true)

	timer.LogTimeTaken()

	ct1 = data[0]
	decrypted1 := utils.Decrypt(ct1)

	ct2 = data[1]
	decrypted2 := utils.Decrypt(ct2)

	ct3 = data[2]
	decrypted3 := utils.Decrypt(ct3)

	log.Log(fmt.Sprintf("Max Level: %d, Post Bootstrapping level: %d", maxLevel, ct1.Level()))

	// Test if bootstrap increase level and correctly decrypt
	if ct1.Level() <= preBootstrap || !ValidateResult(decrypted1, utils.GenerateFilledArray(3.12), false, 1, log) || !ValidateResult(decrypted2, utils.GenerateFilledArray(4.20), false, 1, log) || !ValidateResult(decrypted3, utils.GenerateFilledArray(6.9), false, 1, log) {
		t.Error("Wasn't bootstrapped correctly")
	}

}

func TestTranspose(t *testing.T) {

	// Test case:										Expected result:
	// [[01, 02, 03, 04, 05, 06, 07, 08, 09, 10],		[[01, 11, 21],
	//  [11, 12, 13, 14, 15, 16, 17, 18, 19, 20],		 [02, 12, 22], ...
	//  [21, 22, 23, 24, 25, 26, 27, 28, 29, 30]]		 [10, 20, 30]]

	testCase := make([]rlwe.Ciphertext, 3)

	// Generate test case
	for i := range testCase {

		data := make([]float64, utils.Params.Slots())
		for j := 1; j <= 10; j++ {
			data[j-1] = float64((10 * i) + j)
		}

		testCase[i] = utils.Encrypt(data)

	}

	// Generate expected array to check correctness of the function
	expected := make([][]float64, 10)

	for i := range expected {
		row := make([]float64, utils.Params.Slots())
		for j := 0; j < 3; j++ {
			row[j] = float64(i + 1 + (10 * j))
		}
		expected[i] = row
	}

	// Compute transposed array
	transposedCt := utils.Transpose(testCase, 10)

	for i := range transposedCt {
		decryptedResult := utils.Decrypt(&transposedCt[i])

		if !ValidateResult(decryptedResult, expected[i], false, 1, log) {

			t.Error("Data was incorrectly transposed")

		}
	}

}

func TestOuterProduct(t *testing.T) {

	// Test the correctness of outer product evaluation betweem two ciphertexts
	// Test case:			Expected:
	// A = E(3, 4)			[ E(6, 9, 15, 18),
	// B = E(2, 3, 5, 6)	  E(8, 12, 20, 24)]

	testCaseA := utils.Encrypt([]float64{3, 4})
	testCaseB := utils.Encrypt([]float64{2, 3, 5, 6})

	outerProduct := utils.Outer(&testCaseA, &testCaseB, 2, 4, 1)

	for i := range outerProduct {

		decryptedProduct := utils.Decrypt(outerProduct[i])
		expectedResult := make([]float64, utils.Params.Slots())

		if i == 0 {
			expectedResult[0] = 6
			expectedResult[1] = 9
			expectedResult[2] = 15
			expectedResult[3] = 18
		} else {
			expectedResult[0] = 8
			expectedResult[1] = 12
			expectedResult[2] = 20
			expectedResult[3] = 24
		}

		if !ValidateResult(decryptedProduct, expectedResult, false, 1, log) {

			t.Error("Outer was incorrectly calculated")

		}

	}

}

func TestKeyDumpAndLoad(t *testing.T) {

	data := utils.GenerateFilledArray(5)
	data[1] = 0
	ct := utils.Encrypt(data)

	filename := "/usr/local/go/src/github.com/perm-ai/go-cerebrum/keychain"

	log.Log("Dumping")
	keyChain.DumpKeys(filename, true, true, true, true, true)

	log.Log("Loading")
	loadedKeys := key.LoadKeys(filename, 0, true, true, true, true)

	log.Log("Generating new utils")
	newUtils := NewUtils(loadedKeys, math.Pow(2, 35), 0, true)

	log.Log("Testing")
	if !ValidateResult(newUtils.Decrypt(&ct), data, false, 1, log) {
		t.Error("Incorrect validation")
	}

	rotated := newUtils.RotateNew(&ct, 1)
	rotatedExpected := utils.GenerateFilledArray(5)
	rotatedExpected[0] = 0

	if !ValidateResult(newUtils.Decrypt(&rotated), rotatedExpected, false, 1, log) {
		t.Error("Incorrect rotation")
	}

	productExpected := newUtils.GenerateFilledArray(10)
	productExpected[1] = 0
	product := newUtils.MultiplyConstNew(&ct, 2, true, false)

	if !ValidateResult(newUtils.Decrypt(product), productExpected, false, 1, log) {
		t.Error("Incorrect multiplication")
	}

	utils.Evaluator.DropLevel(&ct, ct.Level()-2)
	newUtils.BootstrapInPlace(&ct)

	if !ValidateResult(newUtils.Decrypt(&ct), data, false, 1, log) {
		t.Error("Incorrect bootstrapping")
	}

}

func TestKeyPairDumpAndLoad(t *testing.T) {

	data := utils.GenerateFilledArray(5)
	data[1] = 0
	ct := utils.Encrypt(data)

	filename := "/usr/local/go/src/github.com/perm-ai/go-cerebrum/"
	filename += "keychain"

	log.Log("Dumping")
	keyChain.DumpKeys(filename, true, true, false, false, false)

	log.Log("Loading")
	loadedKeys := key.LoadKeys(filename, 0, true, true, false, false)
	loadedKeys = key.GenerateKeysFromKeyPair(0, loadedKeys.SecretKey, loadedKeys.PublicKey, false, true)

	log.Log("Generating new utils")
	newUtils := NewUtils(loadedKeys, math.Pow(2, 35), 0, true)

	log.Log("Testing")
	if !ValidateResult(newUtils.Decrypt(&ct), data, false, 1, log) {
		t.Error("Incorrect validation")
	}

	rotated := newUtils.RotateNew(&ct, 1)
	rotatedExpected := utils.GenerateFilledArray(5)
	rotatedExpected[0] = 0

	if !ValidateResult(newUtils.Decrypt(&rotated), rotatedExpected, false, 1, log) {
		t.Error("Incorrect rotation")
	}

	productExpected := newUtils.GenerateFilledArray(10)
	productExpected[1] = 0
	product := newUtils.MultiplyConstNew(&ct, 2, true, false)

	if !ValidateResult(newUtils.Decrypt(product), productExpected, false, 1, log) {
		t.Error("Incorrect multiplication")
	}

	utils.Evaluator.DropLevel(&ct, ct.Level()-2)
	newUtils.BootstrapInPlace(&ct)

	if !ValidateResult(newUtils.Decrypt(&ct), data, false, 1, log) {
		t.Error("Incorrect bootstrapping")
	}

}

func TestInterDotProduct(t *testing.T) {

	as := make([]*rlwe.Ciphertext, 784)
	bs := make([]*rlwe.Ciphertext, 784)
	a := utils.EncryptToPointer(utils.GenerateFilledArray(0.2))
	b := utils.EncryptToPointer(utils.GenerateFilledArray(0.8))

	for i := range as {
		as[i] = a.CopyNew()
		bs[i] = b.CopyNew()
	}

	timer := logger.StartTimer("inter dot product")
	result := utils.InterDotProduct(as, bs, true, true, nil)
	timer.LogTimeTaken()

	decrypted := utils.Decrypt(result)

	fmt.Println(decrypted[0:4])

	if !ValidateResult(decrypted, utils.GenerateFilledArray(125.44), false, 1, log) {
		t.Error("Incorrect bootstrapping")
	}

}

func TestFillArray(t *testing.T){

	plainTestCase := make([]float64, utils.Params.Slots())
	plainTestCase[0] = 1.23
	testCase := utils.EncryptToPointer(plainTestCase)

	utils.FillCiphertextInPlace(testCase, 2000)
	dec := utils.Decrypt(testCase)

	if !ValidateResult(dec, utils.GenerateFilledArraySize(1.23, 2000), false, 1, log) {
		t.Error("Incorrect")
	}

}

func TestInterOuterProduct(t *testing.T) {

	// Test the correctness of outer product evaluation betweem two ciphertexts
	// Test case:			Expected:
	// A = E(3, 4)			[ E(6, 9, 15, 18),
	// B = E(2, 3, 5, 6)	  E(8, 12, 20, 24)]

	testCaseA := []*rlwe.Ciphertext{utils.EncryptToPointer(utils.GenerateFilledArray(3)), utils.EncryptToPointer(utils.GenerateFilledArray(4))}
	testCaseB := []*rlwe.Ciphertext{utils.EncryptToPointer(utils.GenerateFilledArray(2)), utils.EncryptToPointer(utils.GenerateFilledArray(3)), utils.EncryptToPointer(utils.GenerateFilledArray(5)), utils.EncryptToPointer(utils.GenerateFilledArray(6))}

	outerProduct := utils.InterOuter(testCaseA, testCaseB, true)
	expectedResults := [][]float64{
		{6,9,15,18},
		{8,12,20,24},
	}

	for r := range outerProduct {

		for c := range outerProduct[r]{

			decryptedProduct := utils.Decrypt(outerProduct[r][c])
			expectedResult := utils.GenerateFilledArray(expectedResults[r][c])

			if !ValidateResult(decryptedProduct, expectedResult, false, 1, log) {

				t.Error("Outer was incorrectly calculated")

			}

		}

	}

}

func TestSumElementZero(t *testing.T){

	ct := utils.EncryptToPointer(utils.GenerateFilledArray(0))
	utils.SumElementsInPlace(ct)

	decrypted := utils.Decrypt(ct)

	if !ValidateResult(decrypted, utils.GenerateFilledArray(0), false, 1, log) {

		t.Error("Sum zero was incorrectly calculated")

	}

}