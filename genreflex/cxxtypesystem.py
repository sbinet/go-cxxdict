# -*- python -*-
# @purpose model the C/C++ type system, following CLang data model

__all__ = (
    'CxxType',
    )

### stdlib imports ------------------------------------------------------------
#import os
import string
import sys

### globals -------------------------------------------------------------------
_g_xref = {
    '_0': {'elem':'Unknown', 'attrs':{'id':'_0','name':''}, 'subelems':[]},
    }

_g_xid = {
    #'_0': CxxType.by_id('_0'),
    }

_g_name2id = {
    '': '_0',
    }

stldeftab = {}
stldeftab['deque']        = '=','std::allocator'
stldeftab['list']         = '=','std::allocator'
stldeftab['map']          = '=','=','std::less','std::allocator'
stldeftab['multimap']     = '=','=','std::less','std::allocator'
stldeftab['queue']        = '=','std::deque'
stldeftab['set']          = '=','std::less','std::allocator'
stldeftab['multiset']     = '=','std::less','std::allocator'
stldeftab['stack']        = '=','std::deque'
stldeftab['vector']       = '=','std::allocator'
stldeftab['basic_string'] = '=','std::char_traits','std::allocator'
#stldeftab['basic_ostream']= '=','std::char_traits'
#stldeftab['basic_istream']= '=','std::char_traits'
#stldeftab['basic_streambuf']= '=','std::char_traits'
if sys.platform == 'win32' :
  stldeftab['hash_set']      = '=', 'stdext::hash_compare', 'std::allocator'
  stldeftab['hash_multiset'] = '=', 'stdext::hash_compare', 'std::allocator'
  stldeftab['hash_map']      = '=', '=', 'stdext::hash_compare', 'std::allocator'
  stldeftab['hash_multimap'] = '=', '=', 'stdext::hash_compare', 'std::allocator'
else :
  stldeftab['hash_set']      = '=','__gnu_cxx::hash','std::equal_to','std::allocator'
  stldeftab['hash_multiset'] = '=','__gnu_cxx::hash','std::equal_to','std::allocator'
  stldeftab['hash_map']      = '=','=','__gnu_cxx::hash','std::equal_to','std::allocator'
  stldeftab['hash_multimap'] = '=','=','__gnu_cxx::hash','std::equal_to','std::allocator'  

#FIXME... 
#stldeftab['reverse_iterator'] = ('=', '__gnu_cxx::__normal_iterator')


### utils ---------------------------------------------------------------------
_g_normalize_class_cache = {}
def _normalize_class(name,alltempl,_useCache=True,_cache=None) :
    if _cache is None:
        _cache = _g_normalize_class_cache
    if _useCache:
        key = (name,alltempl)
        if _cache.has_key(key):
            return _cache[key]    
        else:
            ret = _normalize_class(name,alltempl,False)
            _cache[key] = ret
            return ret
    names, cnt = [], 0
    # Special cases:
    # a< (0 > 1) >::b
    # a< b::c >
    # a< b::c >::d< e::f >
    for s in string.split(name,'::') :
        if cnt == 0 : names.append(s)
        else        : names[-1] += '::' + s
        cnt += s.count('<')+s.count('(')-s.count('>')-s.count(')')
    normlist = [_normalize_fragment(frag,alltempl,_useCache) for frag in names]
    return string.join(normlist, '::')

def _normalize_fragment(name,alltempl=False,_useCache=True,_cache={}) :
    name = name.strip()
    if _useCache:
        key = (name,alltempl)
        if _cache.has_key(key):
            return _cache[key]    
        else:
            ret = _normalize_fragment(name,alltempl,False)
            _cache[key] = ret
            return ret
    if name.find('<') == -1  : 
        nor =  name
        if nor.find('int') == -1: return nor
        for e in [ ['long long unsigned int', 'unsigned long long'],
                 ['long long int',          'long long'],
                 ['unsigned short int',     'unsigned short'],
                 ['short unsigned int',     'unsigned short'],
                 ['short int',              'short'],
                 ['long unsigned int',      'unsigned long'],
                 ['unsigned long int',      'unsigned long'],
                 ['long int',               'long']] :
            nor = nor.replace(e[0], e[1])
        return nor
    else : clname = name[:name.find('<')]
    if name.rfind('>') < len(clname) : suffix = ''
    else                             : suffix = name[name.rfind('>')+1:]
    args = _get_template_args(name)
    sargs = [_normalize_class(a, alltempl, _useCache=_useCache) for a in args]

    if not alltempl :
        defargs = stldeftab.get (clname)
        if defargs and type(defargs) == type({}):
            args = [_normalize_class(a, True, _useCache=_useCache) for a in args]
            defargs_tup = None
            for i in range (1, len (args)):
                defargs_tup = defargs.get (tuple (args[:i]))
                if defargs_tup:
                    lastdiff = i-1
                    for j in range (i, len(args)):
                        if defargs_tup[j] != args[j]:
                            lastdiff = j
                    sargs = args[:lastdiff+1]
                    break
        elif defargs:
            # select only the template parameters different from default ones
            args = sargs
            sargs = []
            nargs = len(args)
            if len(defargs) < len(args): nargs = len(defargs)
            for i in range(nargs) :  
                if args[i].find(defargs[i]) == -1 : sargs.append(args[i])
        sargs = [_normalize_class(a, alltempl, _useCache=_useCache) for a in sargs]

    nor = clname + '<' + string.join(sargs,',')
    if nor[-1] == '>' : nor += ' >' + suffix
    else              : nor += '>' + suffix
    return nor

def _get_template_args(cls) :
    begin = cls.find('<')
    if begin == -1:
        return []
    end = cls.rfind('>')
    if end == -1:
        return []
    args, cnt = [], 0
    for s in string.split(cls[begin+1:end],','):
        if   cnt == 0 : args.append(s)
        else          : args[-1] += ','+ s
        cnt += s.count('<')+s.count('(')-s.count('>')-s.count(')')
    if len(args) and len(args[-1]) and args[-1][-1] == ' ' :
        args[-1] = args[-1][:-1]
    return args

def cxxtypes_itr():
    for xid in _g_xref.keys():
        yield CxxType.by_id(xid)

def builtin_types_itr():
    for t in cxxtypes_itr():
        if t.is_builtin_type():
            yield t
            
### classes -------------------------------------------------------------------

class CxxName:
    F = 1<<0  # FINAL
    Q = 1<<1  # QUALIFIED
    S = 1<<2  # SCOPED
    pass

class CxxType(object):

    def __init__(self, xid):
        global _g_xref
        self.cxx_id = xid
        elem  = _g_xref[xid]['elem']
        attrs = _g_xref[xid]['attrs']

        self.kind = elem
        self.size = None
        #self.members = []
        #self.bases = []
        
        if attrs.get('size'):
            self.size = int(attrs.get('size'))
            
        if self.is_function_type():
            self.return_type = None
            returns = 'returns' in attrs
            if returns:
                self.return_type = CxxType.by_id(attrs['returns'])
                pass

            self.args_type = []
            args = _g_xref[xid]['subelems']
            for a in args:
                self.args_type.append(CxxArgument(elems=a, parent=self))

        if self.is_structure_or_class_type() or self.is_namespace_type():
            self.members = []
            for m in attrs.get('members','').split():
                xref = self.xref[m]
                #print "+++",m,self.xref[m]
                m = CxxMember(mbr=xref, parent=self)
                if xref['elem'] in (
                    'GetNewDelFunctions',
                    'GetBasesTable',
                    ):
                    continue
                self.members.append(m)

            self.bases = []
            for b in attrs.get('bases','').split():
                access = 'public'
                if b[:10] == 'protected:':
                    b = b[10:]
                    access = 'protected'
                if b[:8]  == 'private:':
                    b = b[8:]
                    access = 'private'
                #bases.append( {'type': b, 'access': access, 'virtual': '-1' } )
                #return bases
                #self.bases.append(CxxType.by_id(b))
                self.bases.append(CxxBase(b, access=access, parent=self))

        if self.is_enum_type():
            self.members = []
            subelems = _g_xref[xid]['subelems']
            for m in subelems:
                #print "+++",m,self.xref[m]
                m = CxxEnumValue(data=m, parent=self)
                self.members.append(m)
                
        return

    @property
    def xref(self):
        global _g_xref
        return _g_xref

    @property
    def _attrs(self):
        return self.xref[self.cxx_id]['attrs']
    
    @staticmethod
    def by_id(xid):
        global _g_xid
        global _g_name2id
        
        if xid in _g_xid:
            return _g_xid[xid]
        _g_xid[xid] = t = CxxType(xid)

        n = t.spelling()
        if not n in _g_name2id:
            _g_name2id[n] = t
            
        return t

    @staticmethod
    def by_name(n):
        if n.startswith('::'):
            if n != '::':
                n = n[len('::'):]
                
        global _g_name2id
        if n in _g_name2id:
            return _g_name2id[n]
        
        for t in cxxtypes_itr():
            if t.spelling() == n:
                _g_name2id[n] = t
                return t
        return None
    
    def __repr__(self):
        return '<CxxType "%s" kind=%s id=%s>' % (
            self.spelling(), self.kind, self.cxx_id
            )
    
    def name(self, mod=0):
        name = self._attrs.get('name')
        if not name:
            name = ''
            if self.is_reference_type():
                #print "---",self.cxx_id,"is ref-type"
                t = CxxType.by_id(self._attrs['type'])
                name += t.spelling() + '&'

            if self.is_pointer_type():
                #print "---",self.cxx_id,"is ptr-type"
                t = CxxType.by_id(self._attrs['type'])
                name += t.spelling() + '*'

            if self.has_qualifiers():
                #print "---",self.cxx_id,"has qualifiers"

                cvr = ''
                if self.is_const_qualified():
                    cvr += 'const '

                if self.is_volatile_qualified():
                    cvr += 'volatile '

                if self.is_register_qualified():
                    cvr += 'register '

                t = CxxType.by_id(self._attrs['type'])
                name = cvr+t.spelling()

            if name == '':
                return "@@unnamed@@"
        
        kind = self.kind
        if kind in ('Destructor',) and name[0] != '~':
            name = '~'+name
        if kind in ('OperatorMethod', 'OperatorFunction'):
            if not name.startswith('operator'):
                name = 'operator' + name
        if kind in ('Converter',):
            name = 'operator '+self.return_type.name(mod)#spelling()
            if not name.startswith('operator'):
                name = 'operator '+name

        if mod & CxxName.S:
            scope = self.declaring_scope()
            if scope:
                scope_name = scope.spelling()
                if scope_name.startswith('::'):
                    scope_name = scope_name[len('::'):]
                name = '::'.join([scope_name, name])
            name = _normalize_class(name, alltempl=False)
            if name.startswith('::'):
                name = name[len('::'):]
                pass
            pass
        
        return name

    def spelling(self):
        #print '===spelling=== [%s]...' % self.cxx_id

        return self.name(CxxName.S)
    
        scope = self.declaring_scope()

        n = ''

        if 'name' in self._attrs:
            n = self.name(CxxName.S)
            pass

        if self.is_function_type():
            n = self.function_prototype()
            return n
        
        if self.is_reference_type():
            #print "---",self.cxx_id,"is ref-type"
            t = CxxType.by_id(self._attrs['type'])
            n += t.spelling() + '&'

        if self.is_pointer_type():
            #print "---",self.cxx_id,"is ptr-type"
            t = CxxType.by_id(self._attrs['type'])
            n += t.spelling() + '*'

        if self.has_qualifiers():
            #print "---",self.cxx_id,"has qualifiers"

            cvr = ''
            if self.is_const_qualified():
                cvr += 'const '
            
            if self.is_volatile_qualified():
                cvr += 'volatile '

            if self.is_register_qualified():
                cvr += 'register '
                
            t = CxxType.by_id(self._attrs['type'])
            n = cvr+t.spelling()
            
        return n
    
    def is_canonical(self):
        return self.cxx_id == self.canonical_type().cxx_id

    def canonical_type(self):
        if getattr(self, '_canonical_type', None) is None:
            pquals = []
            t = self
            while t._attrs.get('type'):
                if t.kind in ('PointerType',):
                    pquals.insert(0,'*')
                if t.kind in ('ReferenceType',):
                    pquals.insert(0, '&')
                t = CxxType.by_id(t._attrs['type'])
                
            self._canonical_type = CxxType.by_name(
                '%s%s' % (t.name(CxxName.S),
                          ''.join(pquals)))
        return self._canonical_type
    
    def is_const_qualified(self):
        if not self.has_qualifiers():
            return False
        return self._attrs.get('const') == '1'
        
    def is_volatile_qualified(self):
        if not self.has_qualifiers():
            return False
        return self._attrs.get('volatile') == '1'
        
    def is_register_qualified(self):
        if not self.has_qualifiers():
            return False
        return self._attrs.get('register') == '1'
        
    def has_qualifiers(self):
        return self.kind in (
            'CvQualifiedType',
            )

    def is_builtin_type(self):
        return self.kind in (
            'FundamentalType',
            )
    
    def is_compound_type(self):
        return self.kind in (
            'Class',
            'Struct',
            'Union',
            )

    def is_enum_type(self):
        return self.kind in (
            'Enumeration',
            )

    def is_function_type(self):
        return self.kind in ('Function',
                             'Method',
                             'OperatorMethod',
                             'OperatorFunction',
                             'Converter',
                             'Constructor',
                             'Destructor',

                             'FunctionType',
                             )

    def is_namespace_type(self):
        return self.kind in (
            'Namespace',
            )
    
    def is_pointer_type(self):
        return self.kind in (
            'PointerType',
            )

    def is_void_pointer_type(self):
        if not self.is_pointer_type():
            return False
        pointee_type = CxxType.by_id(self._attrs['type'])
        return pointee_type.spelling() == 'void'
    
    def is_reference_type(self):
        return self.kind in (
            'ReferenceType',
            )
    
    def is_structure_or_class_type(self):
        return self.kind in (
            'Struct',
            'Class',
            )

    def is_typedef_type(self):
        return self.kind in (
            'Typedef',
            )

    ## access specifiers ---
    def is_public(self):
        access = self._attrs.get('access', None)
        if access is None:
            raise RuntimeError(
                'no access specifier for [%s] (id=%s)' %
                (self.spelling(), self.cxx_id)
                )
        return access == 'public'
    
    def is_protected(self):
        access = self._attrs.get('access', None)
        if access is None:
            raise RuntimeError(
                'no access specifier for [%s] (id=%s)' %
                (self.spelling(), self.cxx_id)
                )
        return access == 'protected'

    def is_private(self):
        access = self._attrs.get('access', None)
        if access is None:
            raise RuntimeError(
                'no access specifier for [%s] (id=%s)' %
                (self.spelling(), self.cxx_id)
                )
        return access == 'private'
    ##
    
    def declaring_scope(self):
        if 'context' in self._attrs:
            return CxxType.by_id(self._attrs['context'])
        return None

    def function_prototype(self):
        proto = ''
        if self.return_type:
            if self.kind in ('Converter',):
                # don't add the converter return type in the prototype
                pass
            else:
                proto += self.return_type.spelling()+' '
                pass
        proto += self.function_name()
        proto += '('
        args = [a.spelling() for a in self.args_type]
        proto += ', '.join(args)
        proto += ')'

        if self._attrs.get('const') == '1':
            proto += ' const'

        if self._attrs.get('pure_virtual') == '1':
            proto += ' =0'

        return proto

    def function_name(self, mod = CxxName.S):
        n = ''
        fct_name = n = self.name(mod)

        if 0:
            scope = self.declaring_scope()
            if scope:
                n += scope.spelling()+'::'+fct_name
            else:
                n += fct_name
        return n
        
    pass # class CxxType

class CxxArgument(object):
    def __init__(self, elems, parent):
        self.fct = parent
        self._cxx_type = elems['type']
        self.default_value = elems.get('default', None)
        self.name = elems.get('name', None)
        return

    @property
    def cxx_type(self):
        return CxxType.by_id(self._cxx_type)
    
    def has_default(self):
        return not (self.default_value is None)

    def has_name(self):
        return not (self.name is None)

    def __repr__(self):
        s = '<CxxArgument type="%s"' % self.cxx_type.spelling()
        if self.has_name():
            s += ' name="%s"' % self.name
        if self.has_default():
            s += ' default=%r' % self.default_value
        return s+'>'

    def spelling(self):
        s = self.cxx_type.spelling()
        if self.has_name():
            s += ' '+self.name
        if self.has_default():
            s += ' = '+self.default_value
        return s
    
    pass # class CxxArgument

class CxxMember(object):
    def __init__(self, mbr, parent):
        self.scope = parent
        self._xref = mbr
        #self.cxx_id   = xid
        #self.cxx_type = CxxType.by_id(elems['type'])
        #self.name     = self.cxx_id._attrs.get('name', None)
        #self.offset   = None
        return

    @property
    def cxx_id(self):
        return CxxType.by_id(self._xref['attrs']['id'])
    
    def spelling(self):
        spelling = self.cxx_id.spelling()
        if self.cxx_id.kind in ('Field',):
            t = CxxType.by_id(self.cxx_id._attrs['type'])
            spelling = t.spelling()
        return spelling

    def is_public(self):
        return self.cxx_id._attrs.get('access') == 'public'

    def is_protected(self):
        return self.cxx_id._attrs.get('access') == 'protected'

    def is_private(self):
        return self.cxx_id._attrs.get('access') == 'private'

    def has_offset(self):
        return not (self.cxx_id._attrs.get('offset') is None)

    def offset(self):
        if self.has_offset():
            return int(self.cxx_id._attrs.get('offset'))
        return None

    def has_name(self):
        return not (self.cxx_id._attrs.get('name') is None)

    def name(self, mod=CxxName.S):
        return self.cxx_id.name(mod)
    
    def __repr__(self):
        s = '<CxxMember type="%s" kind=%s' % (self.spelling(),
                                              self.cxx_id.kind)
        if self.has_name():
            s += ' name="%s"' % self.name()
            
        if self.has_offset():
            s += ' offset=%s' % self.offset()

        access = self.cxx_id._attrs.get('access')
        if access:
            s += ' access=%s' % access
        s += '>'
        return s
    
    def is_data_member(self):
        return not self.is_function()

    def is_function(self):
        return self.cxx_id.is_function_type()

    pass # class CxxMember

class CxxBase(object):
    def __init__(self, xid, access, parent):
        self.scope = parent
        self._cxx_id = xid
        self.access = access

    @property
    def cxx_id(self):
        return CxxType.by_id(self._cxx_id)
    
    def name(self, mod=CxxName.S):
        n = self.cxx_id.name(mod)
        return n

    def __repr__(self):
        return '<CxxBase type="%s" access=%s>' % (self.name(), self.access)

    pass # CxxBase
        
class CxxEnumValue(object):

    def __init__(self, data, parent):
        self.scope = parent
        self.value = data['init']
        self._name  = data['name']

    def name(self, mod=CxxName.S):
        n = self.scope.name(mod) + '::' + self._name
        if n.startswith('::'):
            n = n[len('::'):]
        return n

    def spelling(self):
        return self.name()

    def __repr__(self):
        return '<CxxEnumValue type="%s" name="%s" value=%r>' % (
            self.scope.name(CxxName.S),
            self._name,
            self.value,
            )
