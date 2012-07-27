#include "mylib.hh"

#include <stdio.h>
#include <string.h>
#include <stdlib.h>

Base::Base()
{}

Base::~Base()
{
  printf("Base::~Base...\n");
  fflush(stdout);
}

/*
void
Base::do_hello(const char* who)
{
  printf("Base::do_hello(%s)\n", who);
  fflush(stdout);
}
*/

void
Base::do_virtual_hello(const char *who)
{
  printf("Base::do_virtual_hello(%s)\n", who);
  fflush(stdout);
}

D1::D1(const char *name) :
  Base(),
  m_name(NULL)
{
  m_name = strdup(name);
}

D1::~D1()
{
  printf("D1::~D1[%s]...\n", this->name());
  fflush(stdout);
  free(m_name);
}

void
D1::do_virtual_hello(const char *who)
{
  printf("D1[%s]::do_virtual_hello(%s)\n", this->m_name, who);
  fflush(stdout);
}

void
D1::pure_virtual_method(const char *who)
{
  printf("D1[%s]::pure_virtual_method(%s)\n", this->m_name, who);
  fflush(stdout);
}


