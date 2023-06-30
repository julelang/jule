#ifndef __JULE_STD_JULE_PARSER
#define __JULE_STD_JULE_PARSER

#include "../../api/jule.hpp"

template<typename Vector, typename Item>
jule::Slice<Item> __jule_parser_vector_as_slice(const Vector &vec) noexcept;

template<typename Vector, typename Item>
jule::Slice<Item> __jule_parser_vector_as_slice(const Vector &vec) noexcept {
	jule::Slice<Item> slice;
	slice._len = vec._method_len();
	slice.data.alloc = vec._field_heap;
	return slice;
}

#endif // __JULE_STD_JULE_PARSER
