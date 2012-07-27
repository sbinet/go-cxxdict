// -*- c++ -*-

class Foo 
{
  double m_d;
  Foo(const Foo&);// not implemented
  Foo operator=(const Foo&); // not implemented
public:
  Foo() : m_d(0.) {}
  void setDouble(double d) { m_d = d; }
  void setDouble(int d) { m_d = d; }
  double getDouble() { return m_d; }
};

// a free function
double add42(double x) { return x+42; }
double add42(   int x) { return x+42; }
