package jrpc2server

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/erikdubbelboer/fasthttp"
)

var (
	// Precompute the reflect.Type of error and fasthttp.RequestCtx
	typeOfError   = reflect.TypeOf((*error)(nil)).Elem()
	typeOfRequest = reflect.TypeOf((*fasthttp.RequestCtx)(nil)).Elem()
)

// APIServer - main structure
type APIServer struct {
	services *ServiceMap
}

// ServiceMap is a registry for services.
type ServiceMap struct {
	mutex    sync.Mutex
	services map[string]*Service
}

// Service - sub struct
type Service struct {
	name     string                    // name of service
	rcvr     reflect.Value             // receiver of methods for the service
	rcvrType reflect.Type              // type of the receiver
	methods  map[string]*ServiceMethod // registered methods
}

// ServiceMethod - sub struct
type ServiceMethod struct {
	method    reflect.Method // receiver method
	argsType  reflect.Type   // type of the request argument
	replyType reflect.Type   // type of the response argument
}

// HasMethod returns true if the given method is registered.
//
// The method uses a dotted notation as in "Service.Method".
func (as *APIServer) hasMethod(method string) bool {
	if _, _, err := as.services.get(method); err == nil {
		return true
	}
	return false
}

// RegisterService adds a new service to the server.
//
// The name parameter is optional: if empty it will be inferred from
// the receiver type name.
//
// Methods from the receiver will be extracted if these rules are satisfied:
//
//    - The receiver is exported (begins with an upper case letter) or local
//      (defined in the package registering the service).
//    - The method name is exported.
//    - The method has three arguments: *http.Request, *args, *reply.
//    - All three arguments are pointers.
//    - The second and third arguments are exported or local.
//    - The method has return type error.
//
// All other methods are ignored.
func (as *APIServer) RegisterService(receiver interface{}, name string) error {
	return as.services.register(receiver, name)
}

// get returns a registered service given a method name.
//
// The method name uses a dotted notation as in "Service.Method".
func (m *ServiceMap) get(method string) (*Service, *ServiceMethod, error) {
	parts := strings.Split(method, ".")
	if len(parts) != 2 {
		err := fmt.Errorf("api: service/method request ill-formed: %q", method)
		return nil, nil, err
	}
	m.mutex.Lock()
	service := m.services[parts[0]]
	m.mutex.Unlock()
	if service == nil {
		err := fmt.Errorf("api: can't find service %q", method)
		return nil, nil, err
	}
	serviceMethod := service.methods[parts[1]]
	if serviceMethod == nil {
		err := fmt.Errorf("api: can't find method %q", method)
		return nil, nil, err
	}
	return service, serviceMethod, nil
}

// GetAllServices returns an all registered services
//
// The method name uses a dotted notation as in "Service.Method".
func (as *APIServer) GetAllServices() (map[string]*Service, error) {
	as.services.mutex.Lock()
	service := as.services.services
	as.services.mutex.Unlock()
	return service, nil
}

// register adds a new service using reflection to extract its methods.
func (m *ServiceMap) register(rcvr interface{}, name string) error {
	// Setup service.
	s := &Service{
		name:     name,
		rcvr:     reflect.ValueOf(rcvr),
		rcvrType: reflect.TypeOf(rcvr),
		methods:  make(map[string]*ServiceMethod),
	}

	if name == "" {
		s.name = reflect.Indirect(s.rcvr).Type().Name()
		if !isExported(s.name) {
			return fmt.Errorf("api: type %q is not exported", s.name)
		}
	}

	if s.name == "" {
		return fmt.Errorf("api: no service name for type %q", s.rcvrType.String())
	}

	// Setup methods.
	for i := 0; i < s.rcvrType.NumMethod(); i++ {
		method := s.rcvrType.Method(i)

		mtype := method.Type
		// Method must be exported.
		if method.PkgPath != "" {
			continue
		}
		// Method needs four ins: receiver, ps httprouter.Params, *http.Request, *args, *reply.
		if mtype.NumIn() != 4 {
			continue
		}

		// First argument must be a pointer and must be http.Request.
		reqType := mtype.In(1)
		if reqType.Kind() != reflect.Ptr || reqType.Elem() != typeOfRequest {
			continue
		}
		// Second argument must be a pointer and must be exported.
		args := mtype.In(2)
		if args.Kind() != reflect.Ptr || !isExportedOrBuiltin(args) {
			continue
		}
		// Third argument must be a pointer and must be exported.
		reply := mtype.In(3)
		if reply.Kind() != reflect.Ptr || !isExportedOrBuiltin(reply) {
			continue
		}
		// Method needs one out: error.
		if mtype.NumOut() != 1 {
			continue
		}
		if returnType := mtype.Out(0); returnType != typeOfError {
			continue
		}

		s.methods[method.Name] = &ServiceMethod{
			method:    method,
			argsType:  args.Elem(),
			replyType: reply.Elem(),
		}

	}
	if len(s.methods) == 0 {
		return fmt.Errorf("api: %q has no exported methods of suitable type",
			s.name)
	}
	// Add to the map.
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.services == nil {
		m.services = make(map[string]*Service)
	} else if _, ok := m.services[s.name]; ok {
		return fmt.Errorf("api: service already defined: %q", s.name)
	}
	m.services[s.name] = s
	return nil
}

// isExported returns true of a string is an exported (upper case) name.
func isExported(name string) bool {
	runez, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(runez)
}

// isExportedOrBuiltin returns true if a type is exported or a builtin.
func isExportedOrBuiltin(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return isExported(t.Name()) || t.PkgPath() == ""
}

// NewServer returns a new RPC server.
func NewServer() *APIServer {
	return &APIServer{
		services: new(ServiceMap),
	}
}

// APIHandler handle api request, process it and return result
func (as *APIServer) APIHandler(ctx *fasthttp.RequestCtx) {

	var err error

	var req *ServerRequest

	if ctx.UserValue("PrepareDataHandlerRequestRun").(int) == 0 {
		if ctx.UserValue("PrepareDataHandlerRequest").(*ServerRequest) == nil {
			PrepareDataHandler(ctx)
		}
	}

	if ctx.UserValue("PrepareDataHandlerRequestErr").(error) != nil {
		return
	}

	req = ctx.UserValue("PrepareDataHandlerRequest").(*ServerRequest)

	if req.Version != Version {
		err = &Error{
			Code:    JErrorInvalidReq,
			Message: "jsonrpc must be " + Version,
			Data:    req,
		}

		resp := &ServerResponse{
			Version: Version,
			ID:      req.ID,
			Error:   err.(*Error),
		}

		WriteResponse(ctx, 400, resp)
		return
	}

	serviceSpec, methodSpec, errGet := as.services.get(req.Method)

	if errGet != nil {
		err = &Error{
			Code:    JErrorInternal,
			Message: errGet.Error(),
			Data:    req,
		}

		resp := &ServerResponse{
			Version: Version,
			ID:      req.ID,
			Error:   err.(*Error),
		}

		WriteResponse(ctx, 400, resp)
		return
	}

	// Decode the args.
	args := reflect.New(methodSpec.argsType)
	if errRead := ReadRequestParams(req, args.Interface()); errRead != nil {

		err = &Error{
			Code:    JErrorInvalidReq,
			Message: errRead.Error(),
			Data:    req.Params,
		}

		resp := &ServerResponse{
			Version: Version,
			ID:      req.ID,
			Error:   err.(*Error),
		}

		WriteResponse(ctx, 400, resp)
		return
	}

	// Call the service method.
	reply := reflect.New(methodSpec.replyType)
	errValue := methodSpec.method.Func.Call([]reflect.Value{
		serviceSpec.rcvr,
		reflect.ValueOf(ctx),
		args,
		reply,
	})

	var errResult *Error
	errInter := errValue[0].Interface()
	if errInter != nil {
		errResult = errInter.(*Error)
	}

	if errResult != nil {

		resp := &ServerResponse{
			Version: Version,
			ID:      req.ID,
			Error:   errResult,
		}

		WriteResponse(ctx, 400, resp)
		return
	}

	resp := &ServerResponse{
		Version: Version,
		ID:      req.ID,
		Result:  reply.Interface(),
	}

	WriteResponse(ctx, 200, resp)
	return
}
