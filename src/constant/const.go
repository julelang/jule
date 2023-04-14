package constant

// Constant data.
// Use New_nil function istead of &Const{} for nil literal.
type Const struct {
	data any
}

// Returns new constant value instance from 64-bit signed integer.
func New_i64(x int64) *Const { return &Const{data: x} }

// Returns new constant value instance from 64-bit unsigned integer.
func New_u64(x uint64) *Const { return &Const{data: x} }

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

// Reads data as 64-bit signed integer.
// Returns 0 if data is string, bool or which is not numeric.
func (c *Const) As_i64() int64 {
	switch c.data.(type) {
	case int64:
		return c.data.(int64)

	case uint64:
		return int64(c.data.(uint64))

	case float64:
		return int64(c.data.(float64))

	default:
		return 0
	}
}

// Reads data as 64-bit unsigned integer.
// Returns 0 if data is string, bool or which is not numeric.
func (c *Const) As_u64() uint64 {
	switch c.data.(type) {
	case uint64:
		return c.data.(uint64)

	case int64:
		return uint64(c.data.(int64))

	case float64:
		return uint64(c.data.(float64))

	default:
		return 0
	}
}

// Reads data as 64-bit floating-point.
// Returns 0 if data is string, bool or which is not numeric.
func (c *Const) As_f64() float64 {
	switch c.data.(type) {
	case float64:
		return c.data.(float64)

	case int64:
		return float64(c.data.(int64))

	case uint64:
		return float64(c.data.(uint64))

	default:
		return 0
	}
}

//
// Set
//

// Sets constant value from 64-bit signed integer.
func (c *Const) Set_i64(x int64) { c.data = x }

// Sets constant value from 64-bit unsigned integer.
func (c *Const) Set_u64(x uint64) { c.data = x }

// Sets constant value from boolean.
func (c *Const) Set_bool(x bool) { c.data = x }

// Sets constant value from string.
func (c *Const) Set_str(x string) { c.data = x }

// Sets constant value from 64-bit floating-point.
func (c *Const) Set_f64(x float64) { c.data = x }

// Sets constant value to nil.
func (c *Const) Set_nil() { c.data = nil }

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

//
// Comparison
//

// Reports whether c and x are equals.
// Returns false if type is not supported.
func (c *Const) Eqs(x Const) bool {
	switch {
	case c.Is_nil():
		return x.Is_nil()

	case c.Is_bool():
		return x.Is_bool() && c.Read_bool() == x.Read_bool()

	case c.Is_str():
		return x.Is_str() && c.Read_str() == x.Read_str()

	case c.Is_i64():
		return c.Read_i64() == x.As_i64()

	case c.Is_u64():
		return c.Read_u64() == x.As_u64()

	case c.Is_f64():
		return c.Read_f64() == x.As_f64()

	default:
		return false
	}
}

// Reports whether c less than x.
// Returns false if type is unsupported by operation.
//
// Supported types are:
//  - 64-bit signed integer
//  - 64-bit unsigned integer
//  - 64-bit floating-point
func (c *Const) Lt(x Const) bool {
	switch {
	case c.Is_i64():
		return c.Read_i64() < x.As_i64()

	case c.Is_u64():
		return c.Read_u64() < x.As_u64()

	case c.Is_f64():
		return c.Read_f64() < x.As_f64()

	default:
		return false
	}
}

// Reports whether c greater than x.
// Returns false if type is unsupported by operation.
//
// Supported types are:
//  - 64-bit signed integer
//  - 64-bit unsigned integer
//  - 64-bit floating-point
func (c *Const) Gt(x Const) bool {
	switch {
	case c.Is_i64():
		return c.Read_i64() > x.As_i64()

	case c.Is_u64():
		return c.Read_u64() > x.As_u64()

	case c.Is_f64():
		return c.Read_f64() > x.As_f64()

	default:
		return false
	}
}

//
// Ops
//

// Adds x's value to c's value.
// Reports whether operation is success.
func (c *Const) Add(x Const) bool {
	switch {
	case c.Is_str():
		if x.Is_str() {
			return false
		}
		c.Set_str(c.Read_str() + x.Read_str())

	case c.Is_f64():
		c.Set_f64(c.Read_f64() + x.As_f64())

	case c.Is_i64():
		c.Set_i64(c.Read_i64() + x.As_i64())

	case c.Is_u64():
		c.Set_u64(c.Read_u64() + x.As_u64())

	default:
		return false
	}
	return true
}

// Subs x's value to c's value.
// Reports whether operation is success.
func (c *Const) Sub(x Const) bool {
	switch {
	case c.Is_f64():
		c.Set_f64(c.Read_f64() - x.As_f64())

	case c.Is_i64():
		c.Set_i64(c.Read_i64() - x.As_i64())

	case c.Is_u64():
		c.Set_u64(c.Read_u64() - x.As_u64())

	default:
		return false
	}
	return true
}

// Multiplies x's value to c's value.
// Reports whether operation is success.
func (c *Const) Mul(x Const) bool {
	switch {
	case c.Is_f64():
		c.Set_f64(c.Read_f64() * x.As_f64())

	case c.Is_i64():
		c.Set_i64(c.Read_i64() * x.As_i64())

	case c.Is_u64():
		c.Set_u64(c.Read_u64() * x.As_u64())

	default:
		return false
	}
	return true
}

// Divides x's value to c's value.
// Reports whether operation is success.
// Reports false if divided-by-zero.
func (c *Const) Div(x Const) bool {
	switch {
	case c.Is_f64():
		l := x.As_f64()
		if l == 0 {
			return false
		}
		c.Set_f64(c.Read_f64() / l)

	case c.Is_i64():
		l := x.As_i64()
		if l == 0 {
			return false
		}
		c.Set_i64(c.Read_i64() / l)

	case c.Is_u64():
		l := x.As_u64()
		if l == 0 {
			return false
		}
		c.Set_u64(c.Read_u64() / l)

	default:
		return false
	}
	return true
}
