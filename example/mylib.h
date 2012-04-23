// -*- c++ -*-

#if 1
#include <string>
#include <vector>
#include <utility> // for std::pair
#include <stdint.h>

#include <iostream>
class Foo 
{
public:
  static std::string GetName() { return "Foo"; }

  Foo();
  ~Foo();

  void setDouble(double d) { m_d = d;}
  void setDouble(double d, float x) { m_d = d+x;}
  //void setFoo(const std::string &s) {std::cout << s << std::endl;}
  void setFoo(const std::string &s="f", double x=1) {std::cout << s << x << std::endl;}
  void setInt(int i) { std::cout << i << std::endl;}
  void setInt(long i, float x=2) {std::cout << i+x << std::endl;}
  double getDouble() { return m_d; }

  void setUint64(uint64_t v) { m_uint64 = v; }
  uint64_t getUint64() { return m_uint64; }

  std::vector<Foo> getCollection() { return std::vector<Foo>(1, *this); }
  std::vector<std::vector<std::pair<int, int> > > getMatrix() { return std::vector<std::vector<std::pair<int,int> > >(); }

  std::vector< int>::iterator getIterator() { return std::vector<int>().begin(); }

  static void frobnicate() { std::cout << "frobnicate"; }
  //void frobnicate() { std::cout << "frobnicate -- non static"; }

  Foo operator+(const Foo& a) { return Foo(); }
  Foo& operator+=(const Foo& a) { 
    this->m_d += a.m_d; 
    this->m_uint64 += a.m_uint64;
    return *this;
  }

  Foo& operator=(const Foo& rhs) {
    this->m_d = rhs.m_d;
    this->m_uint64 = rhs.m_uint64;
    return *this;
  }

  void save(std::ostream &out) { out << m_d << m_uint64; }
  
  double getme(double d) { return d; }
  int    getme(int i) { return i; }
  void   getme() const { return ; }
  void   getme()       { return ; }

private:
  double m_d;
  uint64_t m_uint64;
};

Foo::Foo()
{}

Foo::~Foo()
{}

class IFoo 
{
public:

  virtual
  void virtual_meth() const {}

  virtual
  void pure_virtual_meth() const = 0;

};

//template class std::vector<Foo>;
namespace Math {
  void do_hello() {
    std::cout << "helllo" << std::endl;
  }
  double do_add(double i, double j=2) { return i+j; }
  double do_add(int i, int j=2, int k=3, int l=4) { return i+j+k+l; }
  double do_add() { return 42; }

  const std::string& say_hello() { static std::string hi="hi"; return hi;}

  double adder(double i, double j) { return i+j; }
}

namespace Math2 {
  // std::string do_hello() {
  //   return "hi";
  // }
  std::string do_hello(const std::string& name = "you") {
    return std::string("hello ") + name;
  }
  std::string do_hello(const char* name) {
    return std::string("hello -- ") + std::string(name);
  }

  std::string do_hello_c(const char* name) {
    return std::string("hello -- ") + std::string(name);
  }
}

namespace NS {
  class Bar {
  public:
    Bar() {}
    void syHello() {
      std::cout << "hello" << std::endl;
    }
  };

  template<class F> F tmpl_fct() { return F(); }

  template<> int tmpl_fct<int>() { return 42; }

  template<int N> int tmpl_fct_n() { return N; }
  template<> int tmpl_fct_n<42>() { return 42; }
  template<> int tmpl_fct_n<-1>() { return -1; }
}

namespace TT {
  typedef int    foo_t;
  typedef foo_t* bar_t;
  typedef const bar_t baz_t;
  //'int *'    'foo *'    'bar'
}

#endif

#define BIT(n)       (1ULL << (n))
enum MyEnum {
  kValue = BIT(2)
};

typedef int            Ssiz_t;      //String size (int)
struct LongStr_t
{
  Ssiz_t    fCap;    // Max string length (including null)
  Ssiz_t    fSize;   // String length (excluding null)
  char     *fData;   // Long string data
};

enum Enum0 { 
  kMinCap = (sizeof(LongStr_t) - 1)/sizeof(char) <= 2 ?
            2 : (sizeof(LongStr_t) - 1)/sizeof(char),
  kMin1 = 1 >= 2 ? 1 : 2
};

typedef void (*Func_t)();

class TVersionCheck {
public:
   TVersionCheck(int versionCode);  // implemented in TSystem.cxx
};

//static TVersionCheck gVersionCheck(335360);
//#define ROOT_VERSION_CODE 335360
//static TVersionCheck gVersionCheck = ROOT_VERSION_CODE;

class EnumNs
{
public:
   enum ESTLtype { kSTL       = 300 /* TVirtualStreamerInfo::kSTL */, 
                   kSTLbitset =  8
   };
   // TStreamerElement status bits
   enum {
      kHasRange     = BIT(6),
      kCache        = BIT(9),
      kRepeat       = BIT(10),
      kRead         = BIT(11),
      kWrite        = BIT(12),
      kDoNotDelete  = BIT(13)
   };
  
};


class Base
{
public:
  virtual void initialize() = 0;
};

class Base2
{
public:
  virtual void execute() = 0;
};

class Alg : public Base, virtual public Base2
{
public:
  virtual void initialize() { std::cout << "::initialize();\n"; }
  virtual void execute() { std::cout << "::execute();\n"; }
};


class WithPrivateBase: public Base, private Base2
{
public:
  virtual void initialize() { std::cout << "::initialize();\n"; }
  virtual void execute() { std::cout << "::execute();\n"; }

  WithPrivateBase() {}
  WithPrivateBase(int /*i*/) {}

  Enum0 myenum;

  operator double() { return myenum; }

  enum Enum1 { 
    kMinCap = (sizeof(LongStr_t) - 1)/sizeof(char) <= 2 ?
              2 : (sizeof(LongStr_t) - 1)/sizeof(char),
    kMin1 = 1 >= 2 ? 1 : 2
  };

private:
  void some_private_method() { std::cout << "::private()\n"; }
};

// #include <cblas.h>
