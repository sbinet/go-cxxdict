#include <iostream>
#include <string.h>
#include <stdlib.h>

#include "mylib.hh"

#define SC_SUCCESS 0
#define SC_ERROR 1

/*
IAlg::~IAlg()
{}
*/

Alg::Alg(const char* name) :
  //IAlg(),
  m_name(NULL)
{
  m_name = strdup(name);
}

Alg::~Alg()
{
  free(m_name);
}

Sc
Alg::initialize()
{
  std::cerr << "[" << m_name << "]::initialize...\n";
  return SC_SUCCESS;
}

Sc
Alg::execute()
{
  std::cerr << "[" << m_name << "]::execute...\n";
  return SC_SUCCESS;
}

Sc
Alg::finalize()
{
  std::cerr << "[" << m_name << "]::finalize...\n";
  return SC_SUCCESS;
}

App::App()
{
  m_algs = (Alg**)malloc(10*sizeof(Alg*));
  m_nalgs = 0;
}

App::~App()
{
  for (int i=0; i<m_nalgs; i++) {
    std::cerr << "[app]::delete alg=" << m_algs[i]->name() << "\n";
    delete m_algs[i]; m_algs[i] = NULL;
  }
  free(m_algs);
}

Sc
App::addAlg(Alg *alg)
{
  std::cerr << "[app]::addAlg(" << alg->name() << ")...\n";
  if (m_nalgs < 10) {
    m_algs[m_nalgs] = alg;
    m_nalgs++;
  } else {
    std::cerr << "[app]::addAlg: already maxed out!\n";
    return SC_ERROR;
  }

  return SC_SUCCESS;
}

Sc
App::run()
{
  std::cerr << "[app]::run...\n";
  for (int i=0; i<m_nalgs; i++) {
    std::cerr << "alg[" << i << "].execute()...\n";
    m_algs[i]->execute();
    std::cerr << "alg[" << i << "].execute()...[done]\n";

  }
  return SC_SUCCESS;
}
