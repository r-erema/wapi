package service_test

import (
	"testing"

	"github.com/r-erema/wapi/internal/infrastructure/whatsapp"
	"github.com/r-erema/wapi/internal/model"
	"github.com/r-erema/wapi/internal/service"
	"github.com/r-erema/wapi/internal/testutil/mock"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewDTO(t *testing.T) {
	dto := service.NewDTO(dtoMocks(t))
	assert.NotNil(t, dto)
}

func TestDTOSession(t *testing.T) {
	conn, sess := dtoMocks(t)
	dto := service.NewDTO(conn, sess)
	assert.IsType(t, sess, dto.Session())
}

func TestDTOWac(t *testing.T) {
	conn, sess := dtoMocks(t)
	dto := service.NewDTO(conn, sess)
	assert.IsType(t, conn, dto.Wac())
}

func dtoMocks(t *testing.T) (whatsapp.Conn, *model.WapiSession) {
	c := gomock.NewController(t)
	return mock.NewMockConn(c), &model.WapiSession{}
}
