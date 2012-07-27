go-cxxdict
==========

`GoCxx` is a (still experimental) tool to automatically generate `Go`
bindings from any `C/C++` library using GCC-XML informations.

Eventually, its plugin system should allow to support the generation
of `Go` bindings from various sources (`CLang`, `SWIG-XML`, ...)
Its plugin system should also allow to generate bindings for other
languages than just `Go`.

For the moment, only the generation of `Go` code for the `gc` compiler
is supported (`gccgo` is planned though.)


Status
------

What works:

- simple functions
- overloaded functions
- simple classes
- classes with simple inheritance
- classes with overloaded methods
- handling of C strings (ie: `const char*`)


Tests
-----

Go under `go-cxxdict/tests` and run:

``
$ ./run.sh
:: tests: basics cxx-oop
:: running test [basics]...
:: running test [basics]...[ok]
:: running test [cxx-oop]...
:: running test [cxx-oop]...[ok]
``
