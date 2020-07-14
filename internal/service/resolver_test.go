package service_test

import (
	"errors"
	"testing"

	"github.com/r-erema/wapi/internal/service"
	"github.com/r-erema/wapi/internal/testutil/mock"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewQRImgResolver(t *testing.T) {
	tests := []struct {
		name         string
		mocksFactory func(t *testing.T) *mock.MockFileSystem
		expectError  bool
	}{
		{
			name:         "OK",
			mocksFactory: resolverMocks,
			expectError:  false,
		},
		{
			name: "QR file doesn't exists",
			mocksFactory: func(t *testing.T) *mock.MockFileSystem {
				c := gomock.NewController(t)
				fs := mock.NewMockFileSystem(c)
				fs.EXPECT().Stat(gomock.Any()).Return(nil, nil)
				fs.EXPECT().IsNotExist(gomock.Any()).Return(true)
				fs.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(errors.New("something went wrong... "))
				return fs
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			resolver, err := service.NewQRImgResolver("/fake/path", tt.mocksFactory(t))
			if tt.expectError {
				assert.NotNil(t, err)
				assert.Nil(t, resolver)
			} else {
				assert.NotNil(t, resolver)
				assert.Nil(t, err)
			}
		})
	}
}

func TestResolveQrFilePath(t *testing.T) {
	resolver, err := service.NewQRImgResolver("/fake/path", resolverMocks(t))
	require.Nil(t, err)
	qrPath := resolver.ResolveQrFilePath("_wid_")
	assert.Equal(t, "/fake/path/qr__wid_.png", qrPath)
}

func resolverMocks(t *testing.T) *mock.MockFileSystem {
	c := gomock.NewController(t)
	fs := mock.NewMockFileSystem(c)
	fs.EXPECT().Stat(gomock.Any()).Return(nil, nil)
	fs.EXPECT().IsNotExist(gomock.Any()).Return(false)
	return fs
}
