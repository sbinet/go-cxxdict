package gccxml

import (
	"testing"
)

var names_to_normalize [][]string = [][]string{
	// builtins
	{"long long unsigned int",
		"unsigned long long",
		"unsigned long long"},
	{"long long int",
		"long long",
		"long long"},
	{"unsigned short int",
		"unsigned short",
		"unsigned short"},
	{"short unsigned int",
		"unsigned short",
		"unsigned short"},
	{"short int",
		"short",
		"short"},
	{"long unsigned int",
		"unsigned long",
		"unsigned long"},
	{"unsigned long int",
		"unsigned long",
		"unsigned long"},
	{"long int",
		"long",
		"long"},


	// stl-vector
	{"vector<int,std::allocator<int> >",
		"vector<int>",
		"vector<int,std::allocator<int> >"},
	{"vector<long int,std::allocator<long int> >",
		"vector<long>",
		"vector<long,std::allocator<long> >"},
	{"vector<unsigned int,std::allocator<unsigned int> >",
		"vector<unsigned int>",
		"vector<unsigned int,std::allocator<unsigned int> >"},
	{"vector<std::vector<int,std::allocator<int> >,std::allocator<std::vector<int,std::allocator<int> > > >",
		"vector<std::vector<int> >",
		"vector<std::vector<int,std::allocator<int> >,std::allocator<std::vector<int,std::allocator<int> > > >"},


	// stl-set
	{"set<int,std::less<int>,std::allocator<int> >",
		"set<int>",
		"set<int,std::less<int>,std::allocator<int> >"},

	{"set<long int,std::less<long int>,std::allocator<long int> >",
		"set<long>",
		"set<long,std::less<long>,std::allocator<long> >"},

	// stl-map
	{"map<int,char,std::less<int>,std::allocator<std::pair<int,char> > >",
		"map<int,char>",
		"map<int,char,std::less<int>,std::allocator<std::pair<int,char> > >"},
	{"map<long int,unsigned long,std::less<long int>,std::allocator<std::pair<long int,unsigned long> > >",
		"map<long,unsigned long>",
		"map<long,unsigned long,std::less<long>,std::allocator<std::pair<long,unsigned long> > >"},
	{"map<long int,unsigned long int,std::less<long int>,std::allocator<std::pair<long int,unsigned long int> > >",
		"map<long,unsigned long>",
		"map<long,unsigned long,std::less<long>,std::allocator<std::pair<long,unsigned long> > >"},
}

func TestNormalizeName(t *testing.T) {
	alltmpl := false
	for _, v := range names_to_normalize {
		out := normalizeName(v[0], alltmpl)
		if out != v[1] {
			t.Errorf("expected [%s], got [%s] (input='%s')", v[1], out, v[0])
		}
	}
}

func TestNormalizeName_alltmpl(t *testing.T) {
	alltmpl := true
	for _, v := range names_to_normalize {
		out := normalizeName(v[0], alltmpl)
		if out != v[2] {
			t.Errorf("expected [%s], got [%s] (input='%s')", v[2], out, v[0])
		}
	}
}

func eq_strslice(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := 0; i < len(s1); i++ {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}

type str_results struct {
	input string
	expected []string
}

func TestGetTemplateArgs(t *testing.T) {
	names := []str_results{
		{ input:"std::vector<int>",
			expected: []string{"int",},},

		{ input:"std::vector<unsigned int>",
			expected: []string{"unsigned int",},},

		{ input:"std::vector<int,std::allocator<int> >",
			expected: []string{"int", "std::allocator<int>"},},

		{ input:"std::vector<unsigned int,std::allocator<unsigned int> >",
			expected: []string{"unsigned int", "std::allocator<unsigned int>"},},

		{ input:"std::set<int>",
			expected: []string{"int",},},
		
		{ input:"std::set<int,std::less<int>,std::allocator<int> >",
			expected: []string{"int", "std::less<int>", "std::allocator<int>"},},
		
		{ input:"std::set<unsigned int,std::less<unsigned int>,std::allocator<unsigned int> >",
			expected: []string{"unsigned int", "std::less<unsigned int>", "std::allocator<unsigned int>"},},

		{ input: "std::map<int,long>",
			expected: []string{"int","long"}, },
		
		{ input: "std::map<int,unsigned long>",
			expected: []string{"int","unsigned long"}, },
		
		{ input: "std::map<int,unsigned long,std::less<int>,std::allocator<std::pair<int,unsigned long> > >",
			expected: []string{"int","unsigned long","std::less<int>","std::allocator<std::pair<int,unsigned long> >"}, },
		
	}
	for _,v := range names {
		out := getTemplateArgs(v.input)
		if !eq_strslice(out, v.expected) {
			t.Errorf("expected %v, got %v (input='%s')", v.expected, out, v.input)
		}
	}
}

var names_to_cls_normalize [][]string = [][]string{
	// builtins
	{"long long unsigned int",
		"unsigned long long",
		"unsigned long long"},
	{"long long int",
		"long long",
		"long long"},
	{"unsigned short int",
		"unsigned short",
		"unsigned short"},
	{"short unsigned int",
		"unsigned short",
		"unsigned short"},
	{"short int",
		"short",
		"short"},
	{"long unsigned int",
		"unsigned long",
		"unsigned long"},
	{"unsigned long int",
		"unsigned long",
		"unsigned long"},
	{"long int",
		"long",
		"long"},


	// stl-vector
	{"std::vector<int,std::allocator<int> >",
		"std::vector<int>",
		"std::vector<int,std::allocator<int> >"},
	{"std::vector<long int,std::allocator<long int> >",
		"std::vector<long>",
		"std::vector<long,std::allocator<long> >"},
	{"std::vector<unsigned int,std::allocator<unsigned int> >",
		"std::vector<unsigned int>",
		"std::vector<unsigned int,std::allocator<unsigned int> >"},
	{"std::vector<std::vector<int,std::allocator<int> >,std::allocator<std::vector<int,std::allocator<int> > > >",
		"std::vector<std::vector<int> >",
		"std::vector<std::vector<int,std::allocator<int> >,std::allocator<std::vector<int,std::allocator<int> > > >"},


	// stl-set
	{"std::set<int,std::less<int>,std::allocator<int> >",
		"std::set<int>",
		"std::set<int,std::less<int>,std::allocator<int> >"},

	{"std::set<long int,std::less<long int>,std::allocator<long int> >",
		"std::set<long>",
		"std::set<long,std::less<long>,std::allocator<long> >"},

	// stl-map
	{"std::map<int,char,std::less<int>,std::allocator<std::pair<int,char> > >",
		"std::map<int,char>",
		"std::map<int,char,std::less<int>,std::allocator<std::pair<int,char> > >"},
	{"std::map<long int,unsigned long,std::less<long int>,std::allocator<std::pair<long int,unsigned long> > >",
		"std::map<long,unsigned long>",
		"std::map<long,unsigned long,std::less<long>,std::allocator<std::pair<long,unsigned long> > >"},

	{"std::map<long int,unsigned long int,std::less<long int>,std::allocator<std::pair<long int,unsigned long int> > >",
		"std::map<long,unsigned long>",
		"std::map<long,unsigned long,std::less<long>,std::allocator<std::pair<long,unsigned long> > >"},

	// user class
	{"MyFooCls<int,std::less<int> >", "MyFooCls<int>", "MyFooCls<int,std::less<int> >"},

	// string
	{"std::string", "std::string", "std::string",},
	{"std::wstring", "std::wstring", "std::wstring",},

	{"std::basic_string<char,std::char_traits<char>,std::allocator<char> >", 
		"std::basic_string<char>", 
		"std::basic_string<char,std::char_traits<char>,std::allocator<char> >",},
}

func TestNormalizeClass(t *testing.T) {
	alltmpl := false
	for _,v := range names_to_cls_normalize {
		out := normalizeClass(v[0], alltmpl)
		if out != v[1] {
			t.Errorf("expected [%s], got [%s] (input='%s')", v[1], out, v[0])
		}
	}
}

func TestNormalizeClass_alltmpl(t *testing.T) {
	alltmpl := true
	for _,v := range names_to_cls_normalize {
		out := normalizeClass(v[0], alltmpl)
		if out != v[2] {
			t.Errorf("expected [%s], got [%s] (input='%s')", v[2], out, v[0])
		}
	}
}

func TestAddTemplateToName(t *testing.T) {
	tests := [][]string{
		{"tmpl_fct", 
			"void NS::tmpl_fct<int>()", 
			"tmpl_fct<int>"},
		{"tmpl_fct", 
			"void NS::tmpl_fct<unsigned int>()", 
			"tmpl_fct<unsigned int>"},
		{"tmpl_fct", 
			"void NS::tmpl_fct<int,unsigned int>()", 
			"tmpl_fct<int,unsigned int>"},
		{"tmpl_fct", 
			"void NS::tmpl_fct<int,unsigned int>(int a, unsigned int b)", 
			"tmpl_fct<int,unsigned int>"},
	}
	for _,v := range tests {
		out := addTemplateToName(v[0], v[1])
		if out != v[2] {
			t.Errorf("expected [%s], got [%s] (input: name='%s', demangled='%s')",
				v[2], out, v[0], v[1])
		}
	}
}
func init() {
	// test custom templated-class with user-provided template-defaults
	g_stldeftable["MyFooCls"] = []string{ "=", "std::less", }
}