package polynomial

import (
	"errors"
	"github.com/dedis/student_18_lattices/bigint"
	"github.com/LoCCS/bliss/poly"
)

type Poly struct {
	coeffs []bigint.Int
	n      uint32
	q      bigint.Int
	psiReverse []uint32
	psiInvReverse []uint32
}

// NewPolynomial creates a new polynomial with a given degree N and module Q
func NewPolynomial(N uint32, Q bigint.Int) (*Poly, error) {
	if (N & (N - 1)) != 0 { // judge if N is power of 2
		return nil, errors.New("polynomial degree N has to be power of 2")
	}
	nttparams, _ := GenerateNTTParameters(N, Q)
	psiReverse := nttparams.GetPsiReverseUint32()
	psiInvReverse := nttparams.GetPsiInvReverseUint32()
	p := &Poly{make([]bigint.Int, N), N, Q, psiReverse, psiInvReverse}
	for i := range p.coeffs {
		p.coeffs[i].SetInt(0)
	}
	return p, nil
}

// SetCoefficients sets the coefficient of target polynomial p to coeffs
func (p *Poly) SetCoefficients(coeffs []bigint.Int) error {
	if uint32(len(coeffs)) != p.n {
		return errors.New("provided coeffs has different length with target polynomial")
	}
	for i, c := range coeffs {
		p.coeffs[i].SetBigInt(&c)
	}
	return nil
}

// GetCoefficients returns the coefficients of target polynomial p
func (p *Poly) GetCoefficients() []bigint.Int {
	return p.coeffs
}

// GetCoefficientsInt64 returns the low 64 bits of coefficients of target polynomial p as int64
func (p *Poly) GetCoefficientsInt64() []int64 {
	coeffs := make([]int64, p.n)
	for i := range p.coeffs {
		coeffs[i] = p.coeffs[i].Int64()
	}
	return coeffs
}

// AddMod adds then mod the coefficients of p1 and p2
func (p *Poly) AddMod(p1, p2 *Poly) (*Poly, error) {
	if p.n != p1.n || !p.q.EqualTo(&p1.q) ||
		p.n != p2.n || !p.q.EqualTo(&p2.q) ||
		p1.n != p2.n || !p1.q.EqualTo(&p2.q) {
		return nil, errors.New("unmatched degree or module")
	}
	for i := range p.coeffs {
		p.coeffs[i].Add(&p1.coeffs[i], &p2.coeffs[i])
		p.coeffs[i].Mod(&p.coeffs[i], &p.q)
	}
	return p, nil
}

// SubMod subtracts then mod the coefficients of p1 and p2
func (p *Poly) SubMod(p1, p2 *Poly) (*Poly, error) {
	if p.n != p1.n || !p.q.EqualTo(&p1.q) ||
		p.n != p2.n || !p.q.EqualTo(&p2.q) ||
		p1.n != p2.n || !p1.q.EqualTo(&p2.q) {
		return nil, errors.New("unmatched degree or module")
	}
	for i := range p.coeffs {
		p.coeffs[i].Sub(&p1.coeffs[i], &p2.coeffs[i])
		p.coeffs[i].Mod(&p.coeffs[i], &p.q)
	}
	return p, nil
}

// Neg sets the coefficients of polynomial p to the negative of p1'coefficients
func (p *Poly) Neg(p1 *Poly) (*Poly, error) {
	if p.n != p1.n || !p.q.EqualTo(&p1.q) {
		return nil, errors.New("unmatched degree or module")
	}
	for i := range p.coeffs {
		p.coeffs[i].Neg(&p1.coeffs[i], &p.q)
	}
	return p, nil
}

// InnerProduct multiplies polynomials p1 and p2 in coefficient-wise
func (p *Poly) MulCoeffs(p1, p2 *Poly) (*Poly, error) {
	if p.n != p1.n || !p.q.EqualTo(&p1.q) ||
		p.n != p2.n || !p.q.EqualTo(&p2.q) ||
		p1.n != p2.n || !p1.q.EqualTo(&p2.q) {
		return nil, errors.New("unmatched degree or module")
	}
	for i := range p.coeffs {
		p.coeffs[i].Mul(&p1.coeffs[i], &p2.coeffs[i])
		p.coeffs[i].Mod(&p.coeffs[i], &p.q)
	}
	return p, nil
}

// MulScalar multiplies each coefficients of p with scalar
func (p *Poly) MulScalar(p1 *Poly, scalar bigint.Int) (*Poly, error) {
	if p.n != p1.n || !p.q.EqualTo(&p1.q) {
		return nil, errors.New("unmatched degree or module")
	}
	for i := range p.coeffs {
		p.coeffs[i].Mul(&p1.coeffs[i], &scalar)
		//p.coeffs[i].Mod(&p1.coeffs[i], &p.q)
	}
	return p, nil
}

// MulPoly multiplies p1 and p2 in polynomial style
func (p *Poly) MulPoly(p1, p2 *Poly) (*Poly, error) {
	if p.n != p1.n || !p.q.EqualTo(&p1.q) {
		return nil, errors.New("unmatched degree or module")
	}
	p1.NTT()
	p2.NTT()
	p.MulCoeffs(p1, p2)
	p.InverseNTT()
	if p != p1 {
		p1.InverseNTT()
	}
	if p != p2 {
		p2.InverseNTT()
	}
	return p, nil
}

func (p *Poly) DebugMulPoly(p1, p2 *Poly) (*Poly, error) {
	if p.n != p1.n || !p.q.EqualTo(&p1.q) {
		return nil, errors.New("unmatched degree or module")
	}
	// copy the coefficients of our poly to bliss.poly
	coeffs1 := p1.GetCoefficientsInt64()
	coeffs2 := p2.GetCoefficientsInt64()
	_coeffs1 := make([]int32, p.n)
	_coeffs2 := make([]int32, p.n)
	for i := range coeffs1 {
		_coeffs1[i] = int32(coeffs1[i])
		_coeffs2[i] = int32(coeffs2[i])
	}

	_p1, _ := poly.New(0)
	_p1.SetData(_coeffs1)
	_p2, _ := poly.New(0)
	_p2.SetData(_coeffs2)

	// multiply ntt
	tmp1, _ := _p1.NTT()
	tmp2, _ := _p2.MultiplyNTT(tmp1)

	// copy back the coefficients after poly multiplication
	_coeffs := tmp2.GetData()
	coeffs := make([]bigint.Int, p.n)
	for i := range coeffs {
		coeffs[i].SetInt(int64(_coeffs[i]))
	}
	p.SetCoefficients(coeffs)
	return p, nil
}

func (p *Poly) Div(p1 *Poly, scalar bigint.Int) (*Poly, error) {
	if p.n != p1.n || !p.q.EqualTo(&p1.q) {
		return nil, errors.New("unmatched degree or module")
	}
	if scalar.EqualTo(bigint.NewInt(int64(0))) {
		return nil, errors.New("divisor cannot be zero")
	}
	for i := range p.coeffs {
		p.coeffs[i].Div(&p1.coeffs[i], &scalar)
	}
	return p, nil
}

func (p *Poly) DivRound(p1 *Poly, scalar bigint.Int) (*Poly, error) {
	if p.n != p1.n || !p.q.EqualTo(&p1.q) {
		return nil, errors.New("unmatched degree or module")
	}
	if scalar.EqualTo(bigint.NewInt(int64(0))) {
		return nil, errors.New("divisor cannot be zero")
	}
	for i := range p.coeffs {
		p.coeffs[i].DivRound(&p1.coeffs[i], &scalar)
	}
	return p, nil
}

func (p *Poly) Mod(p1 *Poly, m bigint.Int) (*Poly, error) {
	if p.n != p1.n || !p.q.EqualTo(&p1.q) {
		return nil, errors.New("unmatched degree or module")
	}
	for i := range p.coeffs {
		p.coeffs[i].Mod(&p1.coeffs[i], &m)
	}
	return p, nil
}