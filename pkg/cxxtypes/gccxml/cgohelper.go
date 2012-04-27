package gccxml

//
// static int _go_gccxml_char_is_signed()
// {
//   char c = 255;
//   if (c > 128) {
//     return 0;
//   }
//   return 1;
// }
import "C"

import (
	"bitbucket.org/binet/go-cxxdict/pkg/cxxtypes"
)

func gccxml_get_char_type() cxxtypes.TypeKind {

	char_tk := cxxtypes.TK_Char_S
	if C._go_gccxml_char_is_signed() == C.int(0) {
		char_tk = cxxtypes.TK_Char_U
	}

	return char_tk
}

// EOF
