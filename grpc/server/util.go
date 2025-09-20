package server

import "fmt"

func getTypeName(s any) string {
	if nameSrv, ok := s.(interface{ Name() string }); ok {
		return nameSrv.Name()
	}
	return fmt.Sprintf("%T", s)
}
