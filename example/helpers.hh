namespace GoCxx {

  template<typename FUNC>
  void* FuncToVoidPtr(FUNC f) {
    union Cnv_t {
      Cnv_t(FUNC ff): m_func(ff) {}
      FUNC m_func;
      void* m_ptr;
    } u(f);
    return u.m_ptr;
  }

  template <typename FUNC>
  FUNC VoidPtrToFunc(void* p) {
    union Cnv_t {
      Cnv_t(void* pp): m_ptr(pp) {}
      FUNC m_func;
      void* m_ptr;
    } u(p);
    return u.m_func;
  }

} //> namespace GoCxx
