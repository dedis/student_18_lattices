package polynomial

import (
	"github.com/dedis/student_18_lattices/bigint"
)

// NTT performs the number theoretic transform on polynomial p's coefficients
// The implementation is based on the description from https://eprint.iacr.org/2016/504.pdf,
// while the underlying algorithm originates from
// https://www.usenix.org/system/files/conference/usenixsecurity16/sec16_paper_alkim.pdf
func (p *Poly) NTT() (*Poly, error) {
	var j1, j2 uint32
	var U, V, T bigint.Int
	var S *bigint.Int
	t := p.n
	for m := uint32(1); m < p.n; m <<= 1 {
		t >>= 1
		for i := uint32(0); i < m; i++ {
			j1 = 2 * i * t
			j2 = j1 + t - 1
			S = &p.nttParams.PsiReverse[m+i]
			for j := j1; j <= j2; j++ {
				// TODO: implement fast reduction algorithms
				U.SetBigInt(&p.coeffs[j])
				V.Mul(&p.coeffs[j+t], S)
				V.Mod(&V, &p.q)
				T.Add(&U, &V)
				p.coeffs[j].Mod(&T, &p.q)
				T.Sub(&U, &V)
				p.coeffs[j+t].Mod(&T, &p.q)
			}
		}
	}
	return p, nil
}

func (p *Poly) FastNTT() (*Poly, error) {
	var j1, j2 uint32
	var U, V, T bigint.Int
	var S *bigint.Int
	t := p.n
	for m := uint32(1); m < p.n; m <<= 1 {
		t >>= 1
		for i := uint32(0); i < m; i++ {
			j1 = 2 * i * t
			j2 = j1 + t - 1
			S = &p.nttParams.PsiReverseMontgomery[m+i]
			for j := j1; j <= j2; j++ {
				U.SetBigInt(&p.coeffs[j])
				V.Mul(&p.coeffs[j+t], S)
				MontgomeryReduce(&V, &p.nttParams.q, &p.nttParams.qInv, p.nttParams.bitLen)
				T.Add(&U, &V)
				p.coeffs[j].Mod(&T, &p.q)
				T.Sub(&U, &V)
				p.coeffs[j+t].Mod(&T, &p.q)
			}
		}
	}
	return p, nil
}

// InverseNTT performs the inverse number theoretic transform on polynomial p's coefficients
func (p *Poly) InverseNTT() (*Poly, error) {
	var j1, j2, h uint32
	var U, V, T bigint.Int
	var S *bigint.Int
	t := uint32(1)
	for m := p.n; m > 1; m >>= 1 {
		j1 = 0
		h = m >> 1
		for i := uint32(0); i < h; i++ {
			j2 = j1 + t - 1
			S = &p.nttParams.PsiInvReverse[h+i]
			for j := j1; j <= j2; j++ {
				U.SetBigInt(&p.coeffs[j])
				V.SetBigInt(&p.coeffs[j+t])
				T.Add(&U, &V)
				p.coeffs[j].Mod(&T, &p.q)
				T.Sub(&U, &V)
				T.Mul(&T, S)
				p.coeffs[j+t].Mod(&T, &p.q)
			}
			j1 = j1 + (t << 1)
		}
		t <<= 1
	}
	var n_reverse bigint.Int
	n_reverse.Inv(bigint.NewInt(int64(p.n)), &p.q)
	for j := uint32(0); j < p.n; j++ {
		p.coeffs[j].Mod(p.coeffs[j].Mul(&p.coeffs[j], &n_reverse), &p.q)
	}
	return p, nil
}

func DebugNTT(coeffs, psiReverse []int64, q, n int64) []int64 {
	var j1, j2, level int64
	var U, V int64
	var S int64
	t := n
	for m := int64(1); m < n; m <<= 1 {
		level++
		t >>= 1
		for i := int64(0); i < m; i++ {
			j1 = 2 * i * t
			j2 = j1 + t - 1
			S = psiReverse[m+i]
			for j := j1; j <= j2; j++ {
				// TODO: implement fast reduction algorithms
				U = coeffs[j]
				V = montgomeryReduce(coeffs[j+t] * S)
				coeffs[j+t] = barrettReduce(U - V)
				if level&1 == 1 {
					coeffs[j] = barrettReduce(U + V)
				} else {
					coeffs[j] = U + V
				}
			}
		}
	}
	return coeffs
}

func DebugNTT2(coeffs, psiReverse []int64, q, n int64) []int64 {
	var j1, j2 int64
	var U, V int64
	var S int64
	t := n
	for m := int64(1); m < n; m <<= 1 {
		t >>= 1
		for i := int64(0); i < m; i++ {
			j1 = 2 * i * t
			j2 = j1 + t - 1
			S = psiReverse[m+i]
			for j := j1; j <= j2; j++ {
				// TODO: implement fast reduction algorithms
				U = coeffs[j]
				V = (coeffs[j+t] * S) % q
				coeffs[j+t] = (U - V) % q
				coeffs[j] = (U + V) % q
			}
		}
	}
	return coeffs
}

func DebugInverseNTT(coeffs, psiInvReverse []int64, q, n int64) []int64 {
	var j1, j2, h int64
	var U, V, T int64
	var S int64
	t := int64(1)
	for m := n; m > 1; m >>= 1 {
		j1 = 0
		h = m >> 1
		for i := int64(0); i < h; i++ {
			j2 = j1 + t - 1
			S = psiInvReverse[h+i]
			for j := j1; j <= j2; j++ {
				U = coeffs[j]
				V = coeffs[j+t]
				coeffs[j] = barrettReduce(U + V)
				T = barrettReduce(U - V)
				coeffs[j+t] = montgomeryReduce(T * S)
			}
			j1 = j1 + (t << 1)
		}
		t <<= 1
	}
	var temp bigint.Int
	temp.Inv(bigint.NewInt(int64(n)), bigint.NewInt(int64(q)))
	n_reverse := (temp.Int64() << 18) % 7681
	for j := int64(0); j < n; j++ {
		coeffs[j] = montgomeryReduce(coeffs[j] * n_reverse)
	}
	return coeffs
}

func MontgomeryReduce(x, q, qInv *bigint.Int, bitLen uint32) *bigint.Int{
	u := new(bigint.Int).Mul(x, qInv)
	montgomeryMod := new(bigint.Int).Lsh(bigint.NewInt(1), bitLen)
	montgomeryMod.Sub(montgomeryMod, bigint.NewInt(1))
	u.And(u, montgomeryMod)
	u.Mul(u, q)
	x.Add(x, u)
	x.Rsh(x, bitLen)
	if x.Compare(q) != -1.0 {
		return x.Sub(x, q)
	}
	return x
}

func montgomeryReduce(a int64) int64 {
	u := a * 7679
	u &= (1 << 18) - 1
	u *= 7681
	a += u
	a = int64(a >> 18)
	if a >= 7681 {
		return a - 7681
	}
	return a
}

func barrettReduce(a int64) int64 {
	u := int64(a >> 13) // ((uint32_t) a * sinv) >> 16
	u *= 7681
	a -= int64(u)
	return a
}
