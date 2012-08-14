// -*- c++ -*-
#ifndef MYLIB_HH
#define MYLIB_HH 1

#include <vector>
#include <string>

class Class
{
  int m_n_ints;
  std::vector<int> m_ints;
  std::vector<double> m_doubles;

private:
  Class& operator=(const Class&); // not implemented
public:
  Class();
  virtual ~Class();

  const std::vector<int>& ints() const;
  const std::vector<double>& doubles() const;

  int nbr_ints() const { return m_ints.size(); }
  int nbr_doubles() const { return m_doubles.size(); }

  int* n_ints();

  void add(int i);
  void add(double d);
};

class Named
{
  std::string m_name;
private:
  //Named& operator=(const Named&); // not implemented
public:
  Named(const std::string& name);
  virtual ~Named();

  const std::string& name() const;
  void setName(const std::string& name);
};

#endif /* !MYLIB_HH */
