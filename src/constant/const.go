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

// Returns new constant value instance from 64-bit floating-point.
func New_f64(x float64) *Const { return &Const{data: x} }

// Returns new constant value instance with nil.
func New_nil() *Const { return &Const{data: nil} }


//
// Types
//

// Reports whether data is 64-bit signed integer.
func (c *Const) Is_i64() bool {
	switch c.data.(type) {
	case int64:
		return true

	default:
		return false
	}
}

// Reports whether data is 64-bit unsigned integer.
func (c *Const) Is_u64() bool {
	switch c.data.(type) {
	case uint64:
		return true

	default:
		return false
	}
}

// Reports whether data is boolean.
func (c *Const) Is_bool() bool {
	switch c.data.(type) {
	case bool:
		return true

	default:
		return false
	}
}

// Reports whether data is string.
func (c *Const) Is_str() bool {
	switch c.data.(type) {
	case string:
		return true

	default:
		return false
	}
}

// Reports whether data is 64-bit floating-point.
func (c *Const) Is_f64() bool {
	switch c.data.(type) {
	case float64:
		return true

	default:
		return false
	}
}

// Reports whether data is nil.
func (c *Const) Is_nil() bool { return c.data == nil }

// Reports whether c and x has same type.
func (c *Const) Are_same_types(x Const) bool {
	switch {
	case c.Is_i64() == x.Is_i64():
		return true

	case c.Is_u64() == x.Is_u64():
		return true

	case c.Is_f64() == x.Is_f64():
		return true

	case c.Is_bool() == x.Is_bool():
		return true

	case c.Is_str() == x.Is_str():
		return true

	case c.Is_nil() == x.Is_nil():
		return true

	default:
		return false
	}
}
