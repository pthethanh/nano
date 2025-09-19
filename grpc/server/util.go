package server

import "fmt"

func getTypeName(s any) string {
	if nameSrv, ok := s.(interface{ Name() string }); ok {
		return nameSrv.Name()
	}
	if nameSrv, ok := s.(interface{ GetName() string }); ok {
		return nameSrv.GetName()
	}
	return fmt.Sprintf("%T", s)
}
