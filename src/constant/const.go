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
// Read
//

// Reads 64-bit signed integer data.
// Returns 0 if data is not 64-bit signed integer.
func (c *Const) Read_i64() int64 {
	if !c.Is_i64() {
		return 0
	}
	return c.data.(int64)
}

// Reads 64-bit unsigned integer data.
// Returns 0 if data is not 64-bit unsigned integer.
func (c *Const) Read_u64() uint64 {
	if !c.Is_u64() {
		return 0
	}
	return c.data.(uint64)
}

// Reads boolean data.
// Returns false if data is not boolean.
func (c *Const) Read_bool() bool {
	if !c.Is_bool() {
		return false
	}
	return c.data.(bool)
}

// Reads string data.
// Returns empty string if data is not string.
func (c *Const) Read_str() string {
	if !c.Is_str() {
		return ""
	}
	return c.data.(string)
}

// Reads 64-bit floating-point data.
// Returns 0 if data is not 64-bit floating-point.
func (c *Const) Read_f64() float64 {
	if !c.Is_f64() {
		return 0
	}
	return c.data.(float64)
}

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
