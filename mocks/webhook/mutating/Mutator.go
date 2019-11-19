// Code generated by mockery v1.0.0
package mutating

import context "context"
import mock "github.com/stretchr/testify/mock"

import v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// Mutator is an autogenerated mock type for the Mutator type
type Mutator struct {
	mock.Mock
}

// Mutate provides a mock function with given fields: _a0, _a1
func (_m *Mutator) Mutate(_a0 context.Context, _a1 v1.Object) (bool, error) {
	ret := _m.Called(_a0, _a1)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, v1.Object) bool); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, v1.Object) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
