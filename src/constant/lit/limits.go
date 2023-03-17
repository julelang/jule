package lit

// Maximum positive value of 32-bit floating-points.
const MAX_F32 = 0x1p127 * (1 + (1 - 0x1p-23))
// Maximum negative value of 32-bit floating-points.
const MIN_F32 = -0x1p127 * (1 + (1 - 0x1p-23))

// Maximum positive value of 64-bit floating-points.
const MAX_F64 = 0x1p1023 * (1 + (1 - 0x1p-52))
// Maximum negative value of 64-bit floating-points.
const MIN_F64 = -0x1p1023 * (1 + (1 - 0x1p-52))

// Maximum positive value of 8-bit signed integers.
const MAX_I8 = 127
// Maximum negative value of 8-bit signed integers.
const MIN_I8 = -128
// Maximum positive value of 16-bit signed integers.
const MAX_I16 = 32767
// Maximum negative value of 16-bit signed integers.
const MIN_I16 = -32768
// Maximum positive value of 32-bit signed integers.
const MAX_I32 = 2147483647
// Maximum negative value of 32-bit signed integers.
const MIN_I32 = -2147483648
// Maximum positive value of 64-bit signed integers.
const MAX_I64 = 9223372036854775807
// Maximum negative value of 64-bit signed integers.
const MIN_I64 = -9223372036854775808

// Maximum value of 8-bit unsigned integers.
const MAX_U8 = 255
// Maximum value of 16-bit unsigned integers.
const MAX_U16 = 65535
// Maximum value of 32-bit unsigned integers.
const MAX_U32 = 4294967295
// Maximum value of 64-bit unsigned integers.
const MAX_U64 = 18446744073709551615
