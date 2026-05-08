package dto

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/H-BF/corlib/pkg/dict"
	"github.com/stretchr/testify/suite"
)

type (
	fromType struct {
		V int
	}

	toType struct {
		V int
	}

	fromNoConv struct {
		S string
	}

	toNoConv struct {
		S string
	}
)

type dtoImplTestSuite struct {
	suite.Suite
}

func Test_DTOImpl(t *testing.T) {
	suite.Run(t, new(dtoImplTestSuite))
}

func (sui *dtoImplTestSuite) SetupTest() {
	registry = dict.HDict[reflect.Type, any]{}
}

func (sui *dtoImplTestSuite) Test_RegisterAndConvert_Success() {
	Register(func(src fromType) (toType, error) {
		return toType{V: src.V + 1}, nil
	})

	var got toType
	err := Convert(fromType{V: 41}, &got)

	sui.Require().NoError(err)
	sui.Require().Equal(toType{V: 42}, got)
}

func (sui *dtoImplTestSuite) Test_Convert_NoRegisteredConverter() {
	var got toNoConv
	err := Convert(fromNoConv{S: "x"}, &got)

	sui.Require().Error(err)
	sui.Require().ErrorIs(err, ErrDTO)
	sui.Require().True(strings.Contains(err.Error(), "no converter registered"))
	sui.Require().Equal(toNoConv{}, got)
}

func (sui *dtoImplTestSuite) Test_Convert_PropagatesConverterError() {
	wantErr := errors.New("convert failed")
	Register(func(src fromType) (toType, error) {
		return toType(src), wantErr
	})

	var got toType
	err := Convert(fromType{V: 7}, &got)

	sui.Require().ErrorIs(err, wantErr)
	sui.Require().Equal(toType{V: 7}, got)
}
