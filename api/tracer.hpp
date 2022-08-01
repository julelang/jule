// Copyright 2022 The X Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

#ifndef __XXC_TRACER_HPP
#define __XXC_TRACER_HPP

// Structure for traceback.
struct tracer;

struct tracer {
    static constexpr uint_xt _n{20};

    std::array<str_xt, _n> _traces;

    void push(const str_xt &_Src) {
        for (int_xt _index{_n-1}; _index > 0; _index--) {
            this->_traces[_index] = this->_traces[_index-1];
        }
        this->_traces[0] = _Src;
    }

    str_xt string(void) noexcept {
        str_xt _traces{};
        for (const str_xt &_trace: this->_traces) {
            if (_trace.empty()) { break; }
            _traces += _trace;
            _traces += "\n";
        }
        return _traces;
    }

    void ok(void) noexcept {
        for (int_xt _index{0}; _index < _n; _index++) {
            this->_traces[_index] = this->_traces[_index+1];
            if (this->_traces[_index+1].empty()) { break; }
        }
    }
};

#endif // #ifndef __XXC_TRACER_HPP
