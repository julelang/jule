package constant

// Constant data.
// Use New_nil function istead of &Const{} for nil literal.
type Const struct {
	data any
}

// Returns new constant value instance from 64-bit signed integer.
func New_i64(x int64) *Const { return &Const{data: x} }

// Returns new constant value instance from 64-bit unsigned integer.
func New_u64(x uint32) *Const { return &Const{data: x} }

// Returns new constant value instance from boolean.
func New_bool(x bool) *Const { return &Const{data: x} }

// Returns new constant value instance from string.
func New_str(x string) *Const { return &Const{data: x} }

// Returns new constant value instance with nil.
func New_nil() *Const { return &Const{data: nil} }

// Returns new constant value instance from 64-bit floating point.
func New_f64(x float64) *Const { return &Const{data: x} }
