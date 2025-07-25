// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery
// template: testify

package mocks

import (
	"context"

	"github.com/cloudflare/cloudflare-go/v4/custom_hostnames"
	"github.com/cloudflare/cloudflare-go/v4/option"
	mock "github.com/stretchr/testify/mock"
	"github.com/theopenlane/core/pkg/corejobs/internal/cloudflare"
)

// NewMockCustomHostnamesService creates a new instance of MockCustomHostnamesService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockCustomHostnamesService(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockCustomHostnamesService {
	mock := &MockCustomHostnamesService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// MockCustomHostnamesService is an autogenerated mock type for the CustomHostnamesService type
type MockCustomHostnamesService struct {
	mock.Mock
}

type MockCustomHostnamesService_Expecter struct {
	mock *mock.Mock
}

func (_m *MockCustomHostnamesService) EXPECT() *MockCustomHostnamesService_Expecter {
	return &MockCustomHostnamesService_Expecter{mock: &_m.Mock}
}

// Delete provides a mock function for the type MockCustomHostnamesService
func (_mock *MockCustomHostnamesService) Delete(context1 context.Context, s string, customHostnameDeleteParams custom_hostnames.CustomHostnameDeleteParams, vs ...option.RequestOption) (*custom_hostnames.CustomHostnameDeleteResponse, error) {
	var tmpRet mock.Arguments
	if len(vs) > 0 {
		tmpRet = _mock.Called(context1, s, customHostnameDeleteParams, vs)
	} else {
		tmpRet = _mock.Called(context1, s, customHostnameDeleteParams)
	}
	ret := tmpRet

	if len(ret) == 0 {
		panic("no return value specified for Delete")
	}

	var r0 *custom_hostnames.CustomHostnameDeleteResponse
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, custom_hostnames.CustomHostnameDeleteParams, ...option.RequestOption) (*custom_hostnames.CustomHostnameDeleteResponse, error)); ok {
		return returnFunc(context1, s, customHostnameDeleteParams, vs...)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, custom_hostnames.CustomHostnameDeleteParams, ...option.RequestOption) *custom_hostnames.CustomHostnameDeleteResponse); ok {
		r0 = returnFunc(context1, s, customHostnameDeleteParams, vs...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custom_hostnames.CustomHostnameDeleteResponse)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, string, custom_hostnames.CustomHostnameDeleteParams, ...option.RequestOption) error); ok {
		r1 = returnFunc(context1, s, customHostnameDeleteParams, vs...)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockCustomHostnamesService_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type MockCustomHostnamesService_Delete_Call struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//   - context1 context.Context
//   - s string
//   - customHostnameDeleteParams custom_hostnames.CustomHostnameDeleteParams
//   - vs ...option.RequestOption
func (_e *MockCustomHostnamesService_Expecter) Delete(context1 interface{}, s interface{}, customHostnameDeleteParams interface{}, vs ...interface{}) *MockCustomHostnamesService_Delete_Call {
	return &MockCustomHostnamesService_Delete_Call{Call: _e.mock.On("Delete",
		append([]interface{}{context1, s, customHostnameDeleteParams}, vs...)...)}
}

func (_c *MockCustomHostnamesService_Delete_Call) Run(run func(context1 context.Context, s string, customHostnameDeleteParams custom_hostnames.CustomHostnameDeleteParams, vs ...option.RequestOption)) *MockCustomHostnamesService_Delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 context.Context
		if args[0] != nil {
			arg0 = args[0].(context.Context)
		}
		var arg1 string
		if args[1] != nil {
			arg1 = args[1].(string)
		}
		var arg2 custom_hostnames.CustomHostnameDeleteParams
		if args[2] != nil {
			arg2 = args[2].(custom_hostnames.CustomHostnameDeleteParams)
		}
		var arg3 []option.RequestOption
		var variadicArgs []option.RequestOption
		if len(args) > 3 {
			variadicArgs = args[3].([]option.RequestOption)
		}
		arg3 = variadicArgs
		run(
			arg0,
			arg1,
			arg2,
			arg3...,
		)
	})
	return _c
}

func (_c *MockCustomHostnamesService_Delete_Call) Return(customHostnameDeleteResponse *custom_hostnames.CustomHostnameDeleteResponse, err error) *MockCustomHostnamesService_Delete_Call {
	_c.Call.Return(customHostnameDeleteResponse, err)
	return _c
}

func (_c *MockCustomHostnamesService_Delete_Call) RunAndReturn(run func(context1 context.Context, s string, customHostnameDeleteParams custom_hostnames.CustomHostnameDeleteParams, vs ...option.RequestOption) (*custom_hostnames.CustomHostnameDeleteResponse, error)) *MockCustomHostnamesService_Delete_Call {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function for the type MockCustomHostnamesService
func (_mock *MockCustomHostnamesService) Get(context1 context.Context, s string, customHostnameGetParams custom_hostnames.CustomHostnameGetParams, vs ...option.RequestOption) (*custom_hostnames.CustomHostnameGetResponse, error) {
	var tmpRet mock.Arguments
	if len(vs) > 0 {
		tmpRet = _mock.Called(context1, s, customHostnameGetParams, vs)
	} else {
		tmpRet = _mock.Called(context1, s, customHostnameGetParams)
	}
	ret := tmpRet

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 *custom_hostnames.CustomHostnameGetResponse
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, custom_hostnames.CustomHostnameGetParams, ...option.RequestOption) (*custom_hostnames.CustomHostnameGetResponse, error)); ok {
		return returnFunc(context1, s, customHostnameGetParams, vs...)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, custom_hostnames.CustomHostnameGetParams, ...option.RequestOption) *custom_hostnames.CustomHostnameGetResponse); ok {
		r0 = returnFunc(context1, s, customHostnameGetParams, vs...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custom_hostnames.CustomHostnameGetResponse)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, string, custom_hostnames.CustomHostnameGetParams, ...option.RequestOption) error); ok {
		r1 = returnFunc(context1, s, customHostnameGetParams, vs...)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockCustomHostnamesService_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type MockCustomHostnamesService_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - context1 context.Context
//   - s string
//   - customHostnameGetParams custom_hostnames.CustomHostnameGetParams
//   - vs ...option.RequestOption
func (_e *MockCustomHostnamesService_Expecter) Get(context1 interface{}, s interface{}, customHostnameGetParams interface{}, vs ...interface{}) *MockCustomHostnamesService_Get_Call {
	return &MockCustomHostnamesService_Get_Call{Call: _e.mock.On("Get",
		append([]interface{}{context1, s, customHostnameGetParams}, vs...)...)}
}

func (_c *MockCustomHostnamesService_Get_Call) Run(run func(context1 context.Context, s string, customHostnameGetParams custom_hostnames.CustomHostnameGetParams, vs ...option.RequestOption)) *MockCustomHostnamesService_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 context.Context
		if args[0] != nil {
			arg0 = args[0].(context.Context)
		}
		var arg1 string
		if args[1] != nil {
			arg1 = args[1].(string)
		}
		var arg2 custom_hostnames.CustomHostnameGetParams
		if args[2] != nil {
			arg2 = args[2].(custom_hostnames.CustomHostnameGetParams)
		}
		var arg3 []option.RequestOption
		var variadicArgs []option.RequestOption
		if len(args) > 3 {
			variadicArgs = args[3].([]option.RequestOption)
		}
		arg3 = variadicArgs
		run(
			arg0,
			arg1,
			arg2,
			arg3...,
		)
	})
	return _c
}

func (_c *MockCustomHostnamesService_Get_Call) Return(customHostnameGetResponse *custom_hostnames.CustomHostnameGetResponse, err error) *MockCustomHostnamesService_Get_Call {
	_c.Call.Return(customHostnameGetResponse, err)
	return _c
}

func (_c *MockCustomHostnamesService_Get_Call) RunAndReturn(run func(context1 context.Context, s string, customHostnameGetParams custom_hostnames.CustomHostnameGetParams, vs ...option.RequestOption) (*custom_hostnames.CustomHostnameGetResponse, error)) *MockCustomHostnamesService_Get_Call {
	_c.Call.Return(run)
	return _c
}

// New provides a mock function for the type MockCustomHostnamesService
func (_mock *MockCustomHostnamesService) New(context1 context.Context, customHostnameNewParams custom_hostnames.CustomHostnameNewParams, vs ...option.RequestOption) (*custom_hostnames.CustomHostnameNewResponse, error) {
	var tmpRet mock.Arguments
	if len(vs) > 0 {
		tmpRet = _mock.Called(context1, customHostnameNewParams, vs)
	} else {
		tmpRet = _mock.Called(context1, customHostnameNewParams)
	}
	ret := tmpRet

	if len(ret) == 0 {
		panic("no return value specified for New")
	}

	var r0 *custom_hostnames.CustomHostnameNewResponse
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, custom_hostnames.CustomHostnameNewParams, ...option.RequestOption) (*custom_hostnames.CustomHostnameNewResponse, error)); ok {
		return returnFunc(context1, customHostnameNewParams, vs...)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, custom_hostnames.CustomHostnameNewParams, ...option.RequestOption) *custom_hostnames.CustomHostnameNewResponse); ok {
		r0 = returnFunc(context1, customHostnameNewParams, vs...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custom_hostnames.CustomHostnameNewResponse)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, custom_hostnames.CustomHostnameNewParams, ...option.RequestOption) error); ok {
		r1 = returnFunc(context1, customHostnameNewParams, vs...)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockCustomHostnamesService_New_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'New'
type MockCustomHostnamesService_New_Call struct {
	*mock.Call
}

// New is a helper method to define mock.On call
//   - context1 context.Context
//   - customHostnameNewParams custom_hostnames.CustomHostnameNewParams
//   - vs ...option.RequestOption
func (_e *MockCustomHostnamesService_Expecter) New(context1 interface{}, customHostnameNewParams interface{}, vs ...interface{}) *MockCustomHostnamesService_New_Call {
	return &MockCustomHostnamesService_New_Call{Call: _e.mock.On("New",
		append([]interface{}{context1, customHostnameNewParams}, vs...)...)}
}

func (_c *MockCustomHostnamesService_New_Call) Run(run func(context1 context.Context, customHostnameNewParams custom_hostnames.CustomHostnameNewParams, vs ...option.RequestOption)) *MockCustomHostnamesService_New_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 context.Context
		if args[0] != nil {
			arg0 = args[0].(context.Context)
		}
		var arg1 custom_hostnames.CustomHostnameNewParams
		if args[1] != nil {
			arg1 = args[1].(custom_hostnames.CustomHostnameNewParams)
		}
		var arg2 []option.RequestOption
		var variadicArgs []option.RequestOption
		if len(args) > 2 {
			variadicArgs = args[2].([]option.RequestOption)
		}
		arg2 = variadicArgs
		run(
			arg0,
			arg1,
			arg2...,
		)
	})
	return _c
}

func (_c *MockCustomHostnamesService_New_Call) Return(customHostnameNewResponse *custom_hostnames.CustomHostnameNewResponse, err error) *MockCustomHostnamesService_New_Call {
	_c.Call.Return(customHostnameNewResponse, err)
	return _c
}

func (_c *MockCustomHostnamesService_New_Call) RunAndReturn(run func(context1 context.Context, customHostnameNewParams custom_hostnames.CustomHostnameNewParams, vs ...option.RequestOption) (*custom_hostnames.CustomHostnameNewResponse, error)) *MockCustomHostnamesService_New_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockClient creates a new instance of MockClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockClient {
	mock := &MockClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// MockClient is an autogenerated mock type for the Client type
type MockClient struct {
	mock.Mock
}

type MockClient_Expecter struct {
	mock *mock.Mock
}

func (_m *MockClient) EXPECT() *MockClient_Expecter {
	return &MockClient_Expecter{mock: &_m.Mock}
}

// CustomHostnames provides a mock function for the type MockClient
func (_mock *MockClient) CustomHostnames() cloudflare.CustomHostnamesService {
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for CustomHostnames")
	}

	var r0 cloudflare.CustomHostnamesService
	if returnFunc, ok := ret.Get(0).(func() cloudflare.CustomHostnamesService); ok {
		r0 = returnFunc()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(cloudflare.CustomHostnamesService)
		}
	}
	return r0
}

// MockClient_CustomHostnames_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CustomHostnames'
type MockClient_CustomHostnames_Call struct {
	*mock.Call
}

// CustomHostnames is a helper method to define mock.On call
func (_e *MockClient_Expecter) CustomHostnames() *MockClient_CustomHostnames_Call {
	return &MockClient_CustomHostnames_Call{Call: _e.mock.On("CustomHostnames")}
}

func (_c *MockClient_CustomHostnames_Call) Run(run func()) *MockClient_CustomHostnames_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockClient_CustomHostnames_Call) Return(customHostnamesService cloudflare.CustomHostnamesService) *MockClient_CustomHostnames_Call {
	_c.Call.Return(customHostnamesService)
	return _c
}

func (_c *MockClient_CustomHostnames_Call) RunAndReturn(run func() cloudflare.CustomHostnamesService) *MockClient_CustomHostnames_Call {
	_c.Call.Return(run)
	return _c
}
