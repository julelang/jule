// Copyright 2021 The X Programming Language.
// Use of this source code is governed by a BSD 3-Clause
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/the-xlang/xxc/documenter"
	"github.com/the-xlang/xxc/parser"
	"github.com/the-xlang/xxc/pkg/x"
	"github.com/the-xlang/xxc/pkg/xio"
	"github.com/the-xlang/xxc/pkg/xlog"
	"github.com/the-xlang/xxc/pkg/xset"
)

type Parser = parser.Parser

func help(cmd string) {
	if cmd != "" {
		println("This module can only be used as single!")
		return
	}
	helpmap := [][]string{
		{"help", "Show help."},
		{"version", "Show version."},
		{"init", "Initialize new project here."},
		{"doc", "Documentize X source code."},
	}
	max := len(helpmap[0][0])
	for _, key := range helpmap {
		len := len(key[0])
		if len > max {
			max = len
		}
	}
	var sb strings.Builder
	const space = 5 // Space of between command name and description.
	for _, part := range helpmap {
		sb.WriteString(part[0])
		sb.WriteString(strings.Repeat(" ", (max-len(part[0]))+space))
		sb.WriteString(part[1])
		sb.WriteByte('\n')
	}
	println(sb.String()[:sb.Len()-1])
}

func version(cmd string) {
	if cmd != "" {
		println("This module can only be used as single!")
		return
	}
	println("The X Programming Language\n" + x.Version)
}

func initProject(cmd string) {
	if cmd != "" {
		println("This module can only be used as single!")
		return
	}
	bytes, err := json.MarshalIndent(*xset.Default, "", "  ")
	if err != nil {
		println(err)
		os.Exit(0)
	}
	err = ioutil.WriteFile(x.SettingsFile, bytes, 0666)
	if err != nil {
		println(err.Error())
		os.Exit(0)
	}
	println("Initialized project.")
}

func doc(cmd string) {
	cmd = strings.TrimSpace(cmd)
	paths := strings.SplitN(cmd, " ", -1)
	for _, path := range paths {
		path = strings.TrimSpace(path)
		p := compile(path, false, true)
		if p == nil {
			continue
		}
		if printlogs(p) {
			fmt.Println(x.GetErr("doc_couldnt_generated", path))
			continue
		}
		docjson, err := documenter.Documentize(p)
		if err != nil {
			fmt.Println(x.GetErr("error", err.Error()))
			continue
		}
		path = filepath.Join(x.Set.CxxOutDir, path+x.DocExt)
		writeOutput(path, docjson)
	}
}

func processCommand(namespace, cmd string) bool {
	switch namespace {
	case "help":
		help(cmd)
	case "version":
		version(cmd)
	case "init":
		initProject(cmd)
	case "doc":
		doc(cmd)
	default:
		return false
	}
	return true
}

func init() {
	execp, err := os.Executable()
	if err != nil {
		println(err.Error())
		os.Exit(0)
	}
	execp = filepath.Dir(execp)
	x.ExecPath = execp
	x.StdlibPath = filepath.Join(x.ExecPath, x.Stdlib)
	x.LangsPath = filepath.Join(x.ExecPath, x.Langs)

	// Not started with arguments.
	// Here is "2" but "os.Args" always have one element for store working directory.
	if len(os.Args) < 2 {
		os.Exit(0)
	}
	var sb strings.Builder
	for _, arg := range os.Args[1:] {
		sb.WriteString(" " + arg)
	}
	os.Args[0] = sb.String()[1:]
	arg := os.Args[0]
	i := strings.Index(arg, " ")
	if i == -1 {
		i = len(arg)
	}
	if processCommand(arg[:i], arg[i:]) {
		os.Exit(0)
	}
}

func loadLangWarns(path string, infos []fs.FileInfo) {
	i := -1
	for j, f := range infos {
		if f.IsDir() || f.Name() != "warns.json" {
			continue
		}
		i = j
		path = filepath.Join(path, f.Name())
		break
	}
	if i == -1 {
		return
	}
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		println("Language's warnings couldn't loaded (uses default);")
		println(err.Error())
		return
	}
	err = json.Unmarshal(bytes, &x.Warns)
	if err != nil {
		println("Language's warnings couldn't loaded (uses default);")
		println(err.Error())
		return
	}
}

func loadLangErrs(path string, infos []fs.FileInfo) {
	i := -1
	for j, f := range infos {
		if f.IsDir() || f.Name() != "errs.json" {
			continue
		}
		i = j
		path = filepath.Join(path, f.Name())
		break
	}
	if i == -1 {
		return
	}
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		println("Language's errors couldn't loaded (uses default);")
		println(err.Error())
		return
	}
	err = json.Unmarshal(bytes, &x.Errs)
	if err != nil {
		println("Language's errors couldn't loaded (uses default);")
		println(err.Error())
		return
	}
}

func loadLang() {
	lang := strings.TrimSpace(x.Set.Language)
	if lang == "" || lang == "default" {
		return
	}
	path := filepath.Join(x.LangsPath, lang)
	infos, err := ioutil.ReadDir(path)
	if err != nil {
		println("Language couldn't loaded (uses default);")
		println(err.Error())
		return
	}
	loadLangWarns(path, infos)
	loadLangErrs(path, infos)
}

func checkMode() {
	lower := strings.ToLower(x.Set.Mode)
	if lower != xset.ModeTranspile &&
		lower != xset.ModeCompile {
		key, _ := reflect.TypeOf(x.Set).Elem().FieldByName("Mode")
		tag := string(key.Tag)
		// 6 for skip "json:
		tag = tag[6 : len(tag)-1]
		println(x.GetErr("invalid_value_for_key", x.Set.Mode, tag))
		os.Exit(0)
	}
	x.Set.Mode = lower
}

func loadXSet() {
	// File check.
	info, err := os.Stat(x.SettingsFile)
	if err != nil || info.IsDir() {
		println(`X settings file ("` + x.SettingsFile + `") is not found!`)
		os.Exit(0)
	}
	bytes, err := os.ReadFile(x.SettingsFile)
	if err != nil {
		println(err.Error())
		os.Exit(0)
	}
	x.Set, err = xset.Load(bytes)
	if err != nil {
		println("X settings has errors;")
		println(err.Error())
		os.Exit(0)
	}
	loadLang()
	checkMode()
}

// printlogs prints logs and returns true
// if logs has error, false if not.
func printlogs(p *Parser) bool {
	var str strings.Builder
	for _, log := range p.Warns {
		switch log.Type {
		case xlog.FlatWarn:
			str.WriteString("WARNING: ")
			str.WriteString(log.Msg)
		case xlog.Warn:
			str.WriteString("WARNING: ")
			str.WriteString(log.Path)
			str.WriteByte(':')
			str.WriteString(fmt.Sprint(log.Row))
			str.WriteByte(':')
			str.WriteString(fmt.Sprint(log.Column))
			str.WriteByte(' ')
			str.WriteString(log.Msg)
		}
		str.WriteByte('\n')
	}
	for _, log := range p.Errs {
		switch log.Type {
		case xlog.FlatErr:
			str.WriteString("ERROR: ")
			str.WriteString(log.Msg)
		case xlog.Err:
			str.WriteString("ERROR: ")
			str.WriteString(log.Path)
			str.WriteByte(':')
			str.WriteString(fmt.Sprint(log.Row))
			str.WriteByte(':')
			str.WriteString(fmt.Sprint(log.Column))
			str.WriteByte(' ')
			str.WriteString(log.Msg)
		}
		str.WriteByte('\n')
	}
	print(str.String())
	return len(p.Errs) > 0
}

func appendStandard(code *string) {
	year, month, day := time.Now().Date()
	hour, min, _ := time.Now().Clock()
	timeStr := fmt.Sprintf("%d/%d/%d %d.%d (DD/MM/YYYY) (HH.MM)",
		day, month, year, hour, min)
	*code = `// Auto generated by XXC compiler.
// X compiler version: ` + x.Version + `
// Date:               ` + timeStr + `

#if defined(WIN32) || defined(_WIN32) || defined(__WIN32__) || defined(__NT__)
#define _WINDOWS
#endif

// region X_STANDARD_IMPORTS
#include <iostream>
#include <string>
#include <string.h>
#include <functional>
#include <vector>
#include <map>
#ifdef _WINDOWS
#include <windows.h>
#endif
// endregion X_STANDARD_IMPORTS

// region X_CXX_API
// region X_BUILTIN_VALUES
#define nil nullptr
// endregion X_BUILTIN_VALUES

// region X_BUILTIN_TYPES
typedef int8_t        i8;
typedef int16_t       i16;
typedef int32_t       i32;
typedef int64_t       i64;
typedef uint8_t       u8;
typedef uint16_t      u16;
typedef uint32_t      u32;
typedef uint64_t      u64;
typedef std::size_t   size;
typedef float         f32;
typedef double        f64;
typedef void          *voidptr;
#define func          std::function

// region X_STRUCTURES
template<typename _Item_t>
class array {
public:
  std::vector<_Item_t> _buffer{};

  array<_Item_t>(void) noexcept                       {}
  array<_Item_t>(const std::nullptr_t) noexcept       {}
  array<_Item_t>(const array<_Item_t>& _Src) noexcept { this->_buffer = _Src._buffer; }

  array<_Item_t>(const std::initializer_list<_Item_t> &_Src) noexcept
  { this->_buffer = std::vector<_Item_t>(_Src.begin(), _Src.end()); }

  ~array<_Item_t>(void) noexcept { this->_buffer.clear(); }

  typedef _Item_t       *iterator;
  typedef const _Item_t *const_iterator;
  iterator begin(void) noexcept             { return &this->_buffer[0]; }
  const_iterator begin(void) const noexcept { return &this->_buffer[0]; }
  iterator end(void) noexcept               { return &this->_buffer[this->_buffer.size()]; }
  const_iterator end(void) const noexcept   { return &this->_buffer[this->_buffer.size()]; }

  inline void clear(void) noexcept { this->_buffer.clear(); }

  _Item_t *find(const _Item_t &_Item) noexcept {
    iterator _it{this->begin()};
    iterator _end{this->end()};
    for (; _it < _end; ++_it)
    { if (_Item == *_it) { return _it; } }
    return nil;
  }

  _Item_t *find_last(const _Item_t &_Item) noexcept {
    iterator _it{this->end()};
    iterator _begin{this->begin()};
    for (; _it >= _begin; --_it)
    { if (_Item == *_it) { return _it; } }
    return nil;
  }

  void erase(const _Item_t &_Item) noexcept {
    auto _it{this->_buffer.begin()};
    auto _end{this->_buffer.end()};
    for (; _it < _end; ++_it) {
      if (_Item == *_it) {
        this->_buffer.erase(_it);
        return;
      }
    }
  }

  void erase_all(const _Item_t &_Item) noexcept {
    auto _it{this->_buffer.begin()};
    auto _end{this->_buffer.end()};
    for (; _it < _end; ++_it)
    { if (_Item == *_it) { this->_buffer.erase(_it); } }
  }

  void append(const array<_Item_t> &_Items) noexcept {
    for (const _Item_t _item: _Items) { this->_buffer.push_back(_item); }
  }

  bool insert(const size &_Start, const array<_Item_t> &_Items) noexcept {
    auto _it{this->_buffer.begin()+_Start};
    if (_it >= this->_buffer.end()) { return false; }
    this->_buffer.insert(_it, _Items.begin(), _Items.end());
    return true;
  }

  bool operator==(const array<_Item_t> &_Src) const noexcept {
    const size _length = this->_buffer.size();
    const size _Src_length = _Src._buffer.size();
    if (_length != _Src_length) { return false; }
    for (size _index = 0; _index < _length; ++_index)
    { if (this->_buffer[_index] != _Src._buffer[_index]) { return false; } }
    return true;
  }

  bool operator==(const std::nullptr_t) const noexcept       { return this->_buffer.empty(); }
  bool operator!=(const array<_Item_t> &_Src) const noexcept { return !(*this == _Src); }
  bool operator!=(const std::nullptr_t) const noexcept       { return !this->_buffer.empty(); }
  _Item_t& operator[](const size _Index)                     { return this->_buffer[_Index]; }

  friend std::ostream& operator<<(std::ostream &_Stream,
                                  const array<_Item_t> &_Src) {
    _Stream << '[';
    const size _length = _Src._buffer.size();
    for (size _index = 0; _index < _length;) {
      _Stream << _Src._buffer[_index++];
      if (_index < _length) { _Stream << u8", "; }
    }
    _Stream << ']';
    return _Stream;
  }
};

template<typename _Key_t, typename _Value_t>
class map: public std::map<_Key_t, _Value_t> {
public:
  map<_Key_t, _Value_t>(void) noexcept                 {}
  map<_Key_t, _Value_t>(const std::nullptr_t) noexcept {}
  map<_Key_t, _Value_t>(const std::initializer_list<std::pair<_Key_t, _Value_t>> _Src)
  { for (const auto _data: _Src) { this->insert(_data); } }

  array<_Key_t> keys(void) const noexcept {
    array<_Key_t> _keys{};
    for (const auto _pair: *this)
    { _keys._buffer.push_back(_pair.first); }
    return _keys;
  }

  array<_Value_t> values(void) const noexcept {
    array<_Value_t> _values{};
    for (const auto _pair: *this)
    { _values._buffer.push_back(_pair.second); }
    return _values;
  }

  inline bool has(const _Key_t _Key) const noexcept { return this->find(_Key) != this->end(); }
  inline void del(const _Key_t _Key) noexcept { this->erase(_Key); }

  bool operator==(const std::nullptr_t) const noexcept { return this->empty(); }
  bool operator!=(const std::nullptr_t) const noexcept { return !this->empty(); }

  friend std::ostream& operator<<(std::ostream &_Stream,
                                  const map<_Key_t, _Value_t> &_Src) {
    _Stream << '{';
    size _length = _Src.size();
    for (const auto _pair: _Src) {
      _Stream << _pair.first;
      _Stream << ':';
      _Stream << _pair.second;
      if (--_length > 0) { _Stream << u8", "; }
    }
    _Stream << '}';
    return _Stream;
  }
};
// endregion X_STRUCTURES

// region UTF8_ENCODING
constexpr u8 xx   {0xF1};
constexpr u8 as   {0xF0};
constexpr u8 s1   {0x02};
constexpr u8 s2   {0x13};
constexpr u8 s3   {0x03};
constexpr u8 s4   {0x23};
constexpr u8 s5   {0x34};
constexpr u8 s6   {0x04};
constexpr u8 s7   {0x44};
constexpr u8 locb {0b10000000};
constexpr u8 hicb {0b10111111};
constexpr u8 maskx{0b00111111};
constexpr u8 mask2{0b00011111};
constexpr u8 mask3{0b00001111};
constexpr u8 mask4{0b00000111};

u8 first[256] {
  as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, as,
  as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, as,
  as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, as,
  as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, as,
  as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, as,
  as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, as,
  as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, as,
  as, as, as, as, as, as, as, as, as, as, as, as, as, as, as, as,
  xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx,
  xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx,
  xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx,
  xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx,
  xx, xx, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1,
  s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1,
  s2, s3, s3, s3, s3, s3, s3, s3, s3, s3, s3, s3, s3, s4, s3, s3,
  s5, s6, s6, s6, s7, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx, xx,
};

struct acceptRange {
public:
  u8 lo;
  u8 hi;
};

acceptRange acceptRanges[16] {
  {locb, hicb},
  {0xA0, hicb},
  {locb, 0x9F},
  {0x90, hicb},
  {locb, 0x8F},
};

const size runelen(const char *_Src) noexcept {
  const size n{strlen(_Src)};
  if (n < 1) { return 0; }
  const u8 s0{(u8)(*_Src)};
  const u8 x{first[s0]};
  if (x >= as) { return 1; }
  const i32 sz{x&7};
  const acceptRange accept{acceptRanges[x>>4]};
  if (n < sz) { return 1; }
  const u8 s1{(u8)(*(_Src+1))};
  if (s1 < accept.lo || accept.hi < s1) { return 1; }
  if (sz <= 2) { return 2; }
  const u8 s2{(u8)(*(_Src+2))};
  if (s2 < locb || hicb < s2) { return 1; }
  if (sz <= 3) { return 3; }
  const u8 s3{(u8)(*(_Src+3))};
  if (s3 < locb || hicb < s3) { return 1; }
  return 4;
}
// endregion UTF8_ENCODING

struct rune {
public:
  std::vector<u8> _bytes{};

  rune(void) noexcept {}

  rune(const char *_Src) noexcept
  { for (; *_Src; ++_Src) { this->_bytes.push_back((u8)(*_Src)); }; }

  rune(const u8 _Src) noexcept { this->_bytes.push_back(_Src); }

  bool operator==(const rune &_Rune) const noexcept { 
    if (this->_bytes.size() != _Rune._bytes.size()) { return false; }
    for (size _index{0}; _index < this->_bytes.size(); ++_index)
    { if (this->_bytes[_index] != _Rune._bytes[_index]) { return false; } }
    return true;
  }

  bool operator!=(const rune &_Rune) const noexcept { return !(*this == _Rune); }

  friend std::ostream& operator<<(std::ostream &_Stream,
                                  const rune &_Src) {
    return (_Stream << std::string(_Src._bytes.begin(), _Src._bytes.end()));
  }
};

class str {
public:
  std::vector<rune> _buffer{};

  str(void) noexcept {}

  str(const char *_Src) noexcept {
    while (*_Src) {
      size _len{runelen(_Src)};
      rune _rune{};
      while (_len-- > 0) { _rune._bytes.push_back((u8)(*_Src++)); }
      this->_buffer.push_back(_rune);
    }
  }

  str(const str &_Src) noexcept
  { this->_buffer = _Src._buffer; }
  
  str(const array<rune> _Src) noexcept
  { this->_buffer = _Src._buffer; }

  str(const array<u8> _Src) noexcept: str(std::string(_Src.begin(), _Src.end()).c_str())  {}

  typedef rune       *iterator;
  typedef const rune *const_iterator;
  iterator begin(void) noexcept             { return &this->_buffer[0]; }
  const_iterator begin(void) const noexcept { return &this->_buffer[0]; }
  iterator end(void) noexcept               { return &this->_buffer[this->_buffer.size()]; }
  const_iterator end(void) const noexcept   { return &this->_buffer[this->_buffer.size()]; }

  operator array<rune>(void) const noexcept {
    array<rune> _array{};
    _array._buffer = std::vector<rune>{this->begin(), this->end()};
    return _array;
  }

  operator array<u8>(void) const noexcept {
    array<u8> _array{};
    for (const rune &_rune: *this) {
      for (const u8 &_byte: _rune._bytes)
      { _array._buffer.push_back(_byte); }
    }
    return _array;
  }

  rune &operator[](size _Index) { return this->_buffer[_Index]; }

  bool operator==(const str &_Str) const noexcept { 
    if (this->_buffer.size() != _Str._buffer.size()) { return false; }
    for (size _index{0}; _index < this->_buffer.size(); ++_index)
    { if (this->_buffer[_index] != _Str._buffer[_index]) { return false; } }
    return true;
  }

  void operator+=(const str _Str) noexcept {
    for (const rune _rune: _Str)
    { this->_buffer.push_back(_rune); }
  }

  str operator+(const str _Str) const noexcept {
    str _str{};
    _str._buffer = this->_buffer;
    for (const rune _rune: _Str)
    { _str._buffer.push_back(_rune); }
    return _str;
  }

  bool operator!=(const str &_Str) const noexcept { return !(*this == _Str); }

  friend std::ostream& operator<<(std::ostream &_Stream, const str &_Src) {
    for (const rune &_rune: _Src._buffer) { _Stream << _rune; }
    return _Stream;
  }
};
// endregion X_BUILTIN_TYPES

// region X_MISC
class exception: public std::exception {
private:
  std::basic_string<char> _buffer;
public:
  exception(const char *_Str)      { this->_buffer = _Str; }
  const char *what() const throw() { return this->_buffer.c_str(); }
};

template<typename _Alloc_t>
static inline _Alloc_t *xalloc() { return new(std::nothrow) _Alloc_t; }

template <typename _Enum_t, typename _Index_t, typename _Item_t>
static inline void foreach(const _Enum_t _Enum,
                           const func<void(_Index_t, _Item_t)> _Body) {
  _Index_t _index{0};
  for (auto _item: _Enum) { _Body(_index++, _item); }
}

template <typename _Enum_t, typename _Index_t>
static inline void foreach(const _Enum_t _Enum,
                           const func<void(_Index_t)> _Body) {
  _Index_t _index{0};
  for (auto begin = _Enum.begin(), end = _Enum.end(); begin < end; ++begin)
  { _Body(_index++); }
}

template <typename _Key_t, typename _Value_t>
static inline void foreach(const map<_Key_t, _Value_t> _Map,
                           const func<void(_Key_t)> _Body) {
  for (const auto _pair: _Map) { _Body(_pair.first); }
}

template <typename _Key_t, typename _Value_t>
static inline void foreach(const map<_Key_t, _Value_t> _Map,
                           const func<void(_Key_t, _Value_t)> _Body) {
  for (const auto _pair: _Map) { _Body(_pair.first, _pair.second); }
}

template<typename _Function_t, typename _Tuple_t, size_t ... _I_t>
inline auto tuple_as_args(const _Function_t _Function,
                          const _Tuple_t _Tuple,
                          const std::index_sequence<_I_t ...>)
{ return _Function(std::get<_I_t>(_Tuple) ...); }

template<typename _Function_t, typename _Tuple_t>
inline auto tuple_as_args(const _Function_t _Function, const _Tuple_t _Tuple) {
  static constexpr auto _size = std::tuple_size<_Tuple_t>::value;
  return tuple_as_args(_Function, _Tuple, std::make_index_sequence<_size>{});
}

struct defer {
  typedef func<void(void)> _Function_t;
  template<class Callable>
  defer(Callable &&_function): _function(std::forward<Callable>(_function)) {}
  defer(defer &&_Src): _function(std::move(_Src._function))                 { _Src._function = nullptr; }
  ~defer() noexcept                                                         { if (this->_function) { this->_function(); } }
  defer(const defer &)          = delete;
  void operator=(const defer &) = delete;
  _Function_t _function;
};

std::ostream& operator<<(std::ostream &_Stream, const i8 &_Src)
{ return _Stream << (i32)(_Src); }

std::ostream& operator<<(std::ostream &_Stream, const u8 &_Src)
{ return _Stream << (i32)(_Src); }

#define XTHROW(_Msg) throw exception(_Msg)
#define _CONCAT(_A, _B) _A ## _B
#define CONCAT(_A, _B) _CONCAT(_A, _B)
#define DEFER(_Expr) defer CONCAT(XXDEFER_, __LINE__){[&](void) mutable -> void { _Expr; }}
#define XID(_Identifier) CONCAT(_, _Identifier)
// endregion X_MISC

// region X_BUILTIN_FUNCTIONS
template <typename _Obj_t>
static inline void XID(out)(const _Obj_t _Obj) noexcept { std::cout << _Obj; }

template <typename _Obj_t>
static inline void XID(outln)(const _Obj_t _Obj) noexcept {
  XID(out)<_Obj_t>(_Obj);
  std::cout << std::endl;
}
// endregion X_BUILTIN_FUNCTIONS
// endregion X_CXX_API

// region TRANSPILED_X_CODE
` + *code + `
// endregion TRANSPILED_X_CODE

// region X_ENTRY_POINT
int main() {
  std::setlocale(LC_ALL, "");
#ifdef _WINDOWS
  std::wcin.imbue(std::locale::global(std::locale()));
#endif
  _main();
  return EXIT_SUCCESS;
}
// endregion X_ENTRY_POINT`
}

func writeOutput(path, content string) {
	err := os.MkdirAll(x.Set.CxxOutDir, 0777)
	if err != nil {
		println(err.Error())
		os.Exit(0)
	}
	bytes := []byte(content)
	err = ioutil.WriteFile(path, bytes, 0666)
	if err != nil {
		println(err.Error())
		os.Exit(0)
	}
}

func compile(path string, main, justDefs bool) *Parser {
	loadXSet()
	p := parser.New(nil)
	// Check standard library.
	inf, err := os.Stat(x.StdlibPath)
	if err != nil || !inf.IsDir() {
		p.Errs = append(p.Errs, xlog.CompilerLog{
			Type: xlog.FlatErr,
			Msg:  "standard library directory not found",
		})
		return p
	}

	f, err := xio.Openfx(path)
	if err != nil {
		println(err.Error())
		return nil
	}
	p.File = f
	p.Parsef(true, false)
	return p
}

func execPostCommands() {
	for _, cmd := range x.Set.PostCommands {
		fmt.Println(">", cmd)
		parts := strings.SplitN(cmd, " ", -1)
		err := exec.Command(parts[0], parts[1:]...).Run()
		if err != nil {
			println(err.Error())
		}
	}
}

func doSpell(path, cxx string) {
	defer execPostCommands()
	writeOutput(path, cxx)
	switch x.Set.Mode {
	case xset.ModeCompile:
		defer os.Remove(path)
		println("compilation is not supported yet")
	}
}

func main() {
	fpath := os.Args[0]
	p := compile(fpath, true, false)
	if p == nil {
		return
	}
	if printlogs(p) {
		os.Exit(0)
	}
	cxx := p.Cxx()
	appendStandard(&cxx)
	path := filepath.Join(x.Set.CxxOutDir, x.Set.CxxOutName)
	doSpell(path, cxx)
}
