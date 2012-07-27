#ifndef MYLIB_HH
#define MYLIB_HH 1

#define Sc int

/*
class IAlg
{
protected:
  IAlg(){}
  IAlg(const IAlg&); // not implemented
  IAlg& operator=(const IAlg&); // not implemented
public:
  virtual ~IAlg();

  virtual const char* name() const =0;

  virtual Sc initialize() =0;
  virtual Sc execute() =0;
  virtual Sc finalize() =0;
};
*/

class Alg /*: public IAlg */
{
  char *m_name;

private:
  //Alg(); // not implemented
  Alg& operator=(const Alg& other); // not implemented
  Alg(const Alg& o); // not implemented

public:
  Alg(const char* name);
  ~Alg();

  const char * name() const { return m_name; }
  Sc initialize();
  Sc execute();
  Sc finalize();
};

class App
{
  Alg **m_algs;
  int m_nalgs;

  App(const App&); // not implemented
  App& operator=(const App&); // not implemented
public:
  App();
  ~App();

  Sc addAlg(Alg *alg);
  Sc run();
};
#endif /* MYLIB_HH */
