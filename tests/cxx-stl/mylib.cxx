#include "mylib.hh"

Class::Class()
{}

Class::~Class()
{}

const std::vector<int>&
Class::ints() const
{
  return m_ints;
}

const std::vector<double>&
Class::doubles() const
{
  return m_doubles;
}

void
Class::add(int i)
{
  m_ints.push_back(i);
}

void
Class::add(double d)
{
  m_doubles.push_back(d);
}

Named::Named(const std::string& name) :
  m_name(name)
{}

Named::~Named()
{}

const std::string&
Named::name() const
{
  return m_name;
}

void
Named::setName(const std::string& name)
{
  m_name = name;
}
