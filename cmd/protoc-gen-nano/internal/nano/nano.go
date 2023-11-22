package nano

import (
	"github.com/pthethanh/nano/cmd/protoc-gen-nano/internal/generator"
	pb "google.golang.org/protobuf/types/descriptorpb"
)

func init() {
	generator.RegisterPlugin(new(nano))
}

// nano is an implementation of the Go protocol buffer compiler's
// plugin architecture.  It generates bindings for go-nano support.
type nano struct {
	gen *generator.Generator
}

// Name returns the name of this plugin, "nano".
func (g *nano) Name() string {
	return "nano"
}

// Init initializes the plugin.
func (g *nano) Init(gen *generator.Generator) {
	g.gen = gen
}

// P forwards to g.gen.P.
func (g *nano) P(args ...interface{}) { g.gen.P(args...) }

// Generate generates code for the services in the given file.
func (g *nano) Generate(file *generator.FileDescriptor) {
	if len(file.FileDescriptorProto.Service) == 0 {
		return
	}
	for i, service := range file.FileDescriptorProto.Service {
		g.generateService(file, service, i)
	}
}

// GenerateImports generates the import declaration for this file.
func (g *nano) GenerateImports(file *generator.FileDescriptor, imports map[generator.GoImportPath]generator.GoPackageName) {
	if len(file.FileDescriptorProto.Service) == 0 {
		return
	}
	g.P("import (")
	g.P("grpc ", `"google.golang.org/grpc"`)
	if g.gen.GenGW {
		g.P(`"context"`)
		g.P("grpc ", `"google.golang.org/grpc"`)
		g.P(`"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"`)
	}
	g.P(")")
	g.P()
}

// generateService generates all the code for the named service.
func (g *nano) generateService(file *generator.FileDescriptor, service *pb.ServiceDescriptorProto, index int) {

	origServiceName := service.GetName()

	serviceName := generator.CamelCase(origServiceName)
	serviceAlias := "Unimplemented" + serviceName + "Server"

	g.P("func (", serviceAlias, ") ServiceDesc() *grpc.ServiceDesc{")
	g.P("return &", serviceName, "_ServiceDesc")
	g.P("}")

	if g.gen.GenGW {
		g.P()
		g.P("func (", serviceAlias, ") RegisterWithEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption){")
		g.P("Register" + serviceName + "HandlerFromEndpoint(ctx, mux, endpoint, opts)")
		g.P("}")
	}
}
