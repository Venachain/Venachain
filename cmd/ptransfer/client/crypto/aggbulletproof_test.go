package crypto

import (
	"fmt"
	bn256 "github.com/PlatONEnetwork/PlatONE-Go/crypto/bn256/cloudflare"
	"math/big"
	"testing"
)

func TestVectorCommitment_Commit(t *testing.T) {

}

func TestVectorDecompose(t *testing.T) {
	v := make([]*big.Int, 3)
	v[0] = big.NewInt(14)
	v[1] = big.NewInt(5)
	v[2] = big.NewInt(3)
	res := VectorDecompose(v, 2, 8, 3)
	t.Log(res)
}

func TestLPolyCoeff(t *testing.T) {
	var i, n, m int64
	n = 3
	m = 2
	aL := make([]*big.Int, n*m)
	for i = 1; i <= 6; i++ {
		aL[i-1] = big.NewInt(i + 1)
	}
	z := big.NewInt(3)
	l0, _ := LPolyCoeff(aL, z, n*m)
	t.Log(l0) //[21888242871839275222246405745257275088548364400416034343698204186575808495616 0 1 2 3 4]
}

func TestRpolyCoeff(t *testing.T) {
	var i, n, m int64
	n = 3
	m = 2
	aR := make([]*big.Int, n*m)
	sR := make([]*big.Int, n*m)
	for i = 1; i <= 6; i++ {
		aR[i-1] = big.NewInt(i)
		sR[i-1] = big.NewInt(i * i)
	}
	y := big.NewInt(2)
	z := big.NewInt(3)
	r0, r1, _ := RpolyCoeff(aR, sR, y, z, n, m)
	t.Log(r0) //[13 28 60 83 182 396]
	t.Log(r1) //[1 8 36 128 400 1152]
}

func TestComputet0(t *testing.T) {
	var i, n, m int64
	n = 3
	m = 2
	aL := make([]*big.Int, n*m)
	aR := make([]*big.Int, n*m)
	sR := make([]*big.Int, n*m)
	for i = 1; i <= 6; i++ {
		aL[i-1] = big.NewInt(i + 1)
		aR[i-1] = big.NewInt(i)
		sR[i-1] = big.NewInt(i * i)
	}
	y := big.NewInt(2)
	z := big.NewInt(3)
	l0, _ := LPolyCoeff(aL, z, n*m)
	r0, _, _ := RpolyCoeff(aR, sR, y, z, n, m)
	res, _ := VectorInnerProduct(l0, r0)
	t0, _ := Computet0(aL, aR, sR, y, z, n, m)
	t.Log(t0) //2343
	if t0.Cmp(res) == 0 {
		t.Log(true)
	}
}

func TestComputet1(t *testing.T) {
	var i, n, m int64
	n = 3
	m = 2
	aL := make([]*big.Int, n*m)
	aR := make([]*big.Int, n*m)
	sL := make([]*big.Int, n*m)
	sR := make([]*big.Int, n*m)
	for i = 1; i <= 6; i++ {
		aL[i-1] = big.NewInt(i + 1)
		aR[i-1] = big.NewInt(i)
		sL[i-1] = big.NewInt(i)
		sR[i-1] = big.NewInt(i * i)
	}
	y := big.NewInt(2)
	z := big.NewInt(3)
	l0, _ := LPolyCoeff(aL, z, n*m)
	r0, r1, _ := RpolyCoeff(aR, sR, y, z, n, m)
	left, _ := VectorInnerProduct(l0, r1)
	right, _ := VectorInnerProduct(sL, r0)
	res := new(big.Int).Add(left, right)
	t1, _ := Computet1(aL, aR, sL, sR, y, z, n, m)
	t.Log(t1) //9966
	if t1.Cmp(res) == 0 {
		t.Log(true)
	}
}

func TestComputet2(t *testing.T) {
	var i, n, m int64
	n = 3
	m = 2
	aR := make([]*big.Int, n*m)
	sL := make([]*big.Int, n*m)
	sR := make([]*big.Int, n*m)
	for i = 1; i <= 6; i++ {
		aR[i-1] = big.NewInt(i)
		sL[i-1] = big.NewInt(i)
		sR[i-1] = big.NewInt(i * i)
	}
	y := big.NewInt(2)
	z := big.NewInt(3)
	_, r1, _ := RpolyCoeff(aR, sR, y, z, n, m)

	res, _ := VectorInnerProduct(sL, r1)
	t2, _ := Computet2(sL, aR, sR, y, z, n, m)
	t.Log(t2) //9549
	if t2.Cmp(res) == 0 {
		t.Log(true)
	}
}

func TestComputeAggLx(t *testing.T) {
	var i, n, m int64
	n = 3
	m = 2
	aL := make([]*big.Int, n*m)
	sL := make([]*big.Int, n*m)
	for i = 1; i <= 6; i++ {
		aL[i-1] = big.NewInt(i + 1)
		sL[i-1] = big.NewInt(i)
	}
	x := big.NewInt(2)
	z := big.NewInt(3)
	lx, _ := ComputeAggLx(aL, sL, x, z, n*m)
	t.Log(lx) //[1 4 7 10 13 16]
}

func TestComputeP(t *testing.T) {
	var i, j, n, m int64
	n = 3
	m = 2
	gVector := make([]*bn256.G1, n*m)
	hPrime := make([]*bn256.G1, n*m)
	for i = 1; i <= 3; i++ {
		gVector[i-1] = new(bn256.G1).ScalarBaseMult(big.NewInt(2))
		hPrime[i-1] = new(bn256.G1).ScalarBaseMult(big.NewInt(3))
	}
	for j = 4; j <= 6; j++ {
		gVector[j-1] = new(bn256.G1).ScalarBaseMult(big.NewInt(3))
		hPrime[j-1] = new(bn256.G1).ScalarBaseMult(big.NewInt(2))
	}
	A := new(bn256.G1).ScalarBaseMult(big.NewInt(7))
	S := new(bn256.G1).ScalarBaseMult(big.NewInt(5))
	x := big.NewInt(3)
	y := big.NewInt(2)
	z := big.NewInt(2)
	h := new(bn256.G1).ScalarBaseMult(big.NewInt(2))
	mu := big.NewInt(11)
	res, _ := UpdateAggP(A, S, h, gVector, hPrime, x, y, z, mu, n, m)
	t.Log(res)
	t.Log(new(bn256.G1).ScalarBaseMult(big.NewInt(432)))
}

func TestAggBpProve(t *testing.T) {
	aggBp := new(AggBulletProof)
	t.Log(aggBp)
}

func TestAggBulletProof_AggBpVerify(t *testing.T) {
	AggBp()
	v := make([]*big.Int, 2)
	v[0] = big.NewInt(3)
	v[1] = big.NewInt(7)
	values := AggBpWitness{v}
	param := aggbpparam
	proof, _ := AggBpProve(&param, &values)
	fmt.Println("proof:", proof)
	res, _ := AggBpVerify(proof, &param)
	fmt.Println("res:", res)
}
func TestAggBulletProof_AggBpVerify_s(t *testing.T) {
	AggBp()
	v := make([]*big.Int, 2)
	v[0] = big.NewInt(3)
	v[1] = big.NewInt(16)
	//values := AggBpWitness{v}
	param := aggbpparam
	//proof, _ := AggBpProve_s(&param, v)
	//fmt.Println("proof:", proof)
	proof := "0xf904bbb8800a34f1a213d254de0eba0b69d8e7b56b37d894ac09a07f4f5525da5a938a04e7237eec3b2ba873f123183690dc6e47771219843dfa7a6f935b89699ac62cce7806d1640ba69600dd350f71c4e79ceb7d30f38b22fb8fe01c8f6512e2215624410bd4e253c794bf614e652b47efc79baaf73ed1dbea455f3c4d785e5f9fe771c9b840220f75a94d7b3a2bbed79b6f0ac89035229a756e13f55fafcf80b5c6d3697de408877646ec2d0090fb888ade54591a6d2e3bbee164fd2090d96ac3251476f7deb84022ca8bdd94e5b1fd78ece1c6f717dccfb54bbbf207f9de6cea06894be37b70521d43ec746c118fd82a1266c13e22a3d9ba6507eafd8769ccde58844bd2862ffdb84026e6d023af66c2a95cd4f2c284388e0bbcaf19beff9e46277c78a7eda6214ad628a4cec887763731ee7e2da7267d93d6ba8790a9af2221bbd335ae579a23f6ceb8400e6920b42110f76016cfc3c5701d626dcbf342e304d43174b45e619e3c4dd71d1452a1e32df810ee157b7acf8e3291e636f94f5ccdd6e5ea93eb99f4bc6ad60da02dc3d3af9e817a2754612dbafec846caaa75c03342d6709865f0af5b11c2e44ea025e54488f44ae6a0e798a0d405cd5e788a029be47e9bc5355283c7c5beadc8e4a03051775421defd0ab2a136ac34d5b7d033685c59d2c23a9202fb3a2a85ed98f2b902cbf902c8a0075418c2087faae3ceb69017d76735066a61c823f01f688a2d4ce2de982637e2a008c8cd29f91ef7ab7e90dac4e35200a7f395535bfa232db4ffb14be46d5635dcb90140002446a3f4867c6d3b2e27750190f64b2f55c84e69d40a73f2d6fb1723ad52fc12b60c82a8af72e31eef8d56b0a8b79301c32e033e490daf592330b85c65276f2a678742fb5f2202616287f04d1c89fd3c678615c3b40c701500205b4b60208008703238de3fbe6386db40695c039b6c2267583677c795cbbd06b0612e83c89f2102cc43d1c5e7e1db689354c93f216339f8e561d500bcea73a092f8d8db45be19c83bb4a2805a47bea16bb87b4f8d02a5f941597fa0acd15c6c8eb0aefbf5ba2bff8b30aa6d2d90eb11d95e91610ed1eaccd6c42b0633649f8ffa407f038fc418af3087e03717a170592094abfb054c71a35543a03a48550eb5fc0b2ebcb56d2ce137faaf3169360d6fc0fb4f4754ae28ced8cd04671cdd4d6e58353a22bae718afa93e25bbfe76e3a00c17d7d9158f3564ca22da29d59a8a2e14677171750fb9014016ff771ed3710b0ec7598e620c153f51089de86269d3c2f920b85fc20dbd8ece260db4a6c9186214cb95bbb4c0d6dfaf48379b47cf65338593e488124ceda4e32fc2dec5513bc409956dc930db830e78cb3f7cbe0ff94632231be38596964dc81bdbb4e172bff811d8bb60ddbe5815182b847d3d71328c3833bf6fc848d83689002968dcf05274122404af7ee30ce1b5b6ddbcd5bcd10e6cb3c3977657bd3d9e1cafba49198e16381abd29776f8176f80610ed4bc35f85a4256b6d8fa9f500b629f2bf3606aa7f6004de44aba453114d097cceeb29f6f3527be21aaa6807680e227fe509c71a2b323edb24cc6bceed35de6ff9d9bd958ace9434099d3d7942d327d71888320049c9c6b155ec239cae29c55341d0cdc362a93c86c718f742cd2a1da8596bf00c8edce7b52c0551513a8c070f3a1a76f67785cc37233d458fe9b5"
	res, _ := AggBpVerify_s(proof, &param)

	fmt.Println("res:", res)
}