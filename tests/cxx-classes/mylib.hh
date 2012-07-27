// -*- c++ -*-
#ifndef MYLIB_HH
#define MYLIB_HH 1

class Base
{
private:
  Base& operator=(const Base&); // not implemented
public:
  Base();
  virtual ~Base();

  //FIXME
  //void do_hello(const char *who);

  virtual void do_virtual_hello(const char *who);
  virtual void pure_virtual_method(const char *who)=0;
  virtual const char* name() const =0;
};

class D1 : public Base
{
  char *m_name;
  //D1& operator=(const D1&); // not implemented
public:
  D1(const char* name);
  virtual ~D1();

  virtual void do_virtual_hello(const char *who);
  virtual void pure_virtual_method(const char *who);
  virtual const char* name() const { return m_name; }
};

#endif /* !MYLIB_HH */
